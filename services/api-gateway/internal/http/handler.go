package http

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	platformredis "github.com/adedaryorh/logistics-platform/pkg/redis"
)

type Route struct {
	Prefix string
	Target string
	Auth   bool
	Role   string
}

type rateLimit struct {
	Requests int
	Window   time.Duration
}

type rateBucket struct {
	Count   int
	ResetAt time.Time
}

type Handler struct {
	routes      []Route
	rateMu      sync.Mutex
	rateBuckets map[string]rateBucket
	redis       platformredis.Store
	cfg         *platformconfig.Config
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	handler := &Handler{
		routes:      defaultRoutes(),
		rateBuckets: map[string]rateBucket{},
		redis:       platformredis.NewStoreFromConfig(cfg),
		cfg:         cfg,
	}

	router.Use(handler.cors())
	router.Use(handler.securityHeaders())
	router.Use(handler.maxBodySize(1 << 20))
	router.Use(handler.routeGuard())
	router.GET("/readyz", func(c *gin.Context) {
		writeSuccess(c, http.StatusOK, gin.H{"status": "ready"})
	})
	router.NoRoute(handler.proxyRequest)
}

func defaultRoutes() []Route {
	return []Route{
		{Prefix: "/api/v1/auth", Target: "http://identity-service:8081", Auth: false},
		{Prefix: "/api/v1/users", Target: "http://identity-service:8081", Auth: true},
		{Prefix: "/api/v1/orders", Target: "http://logistics-service:8082", Auth: true},
		{Prefix: "/api/v1/drivers", Target: "http://logistics-service:8082", Auth: true},
		{Prefix: "/api/v1/merchants", Target: "http://logistics-service:8082", Auth: true},
		{Prefix: "/api/v1/payments", Target: "http://payment-service:8084", Auth: true},
		{Prefix: "/api/v1/webhooks", Target: "http://payment-service:8084", Auth: false},
		{Prefix: "/api/v1/admin", Target: "http://operations-service:8085", Auth: true, Role: "admin"},
		{Prefix: "/api/v1/mobility", Target: "http://mobility-service:8083", Auth: true},
	}
}

func (h *Handler) routeGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		route, ok := h.matchRoute(c.Request.URL.Path)
		if !ok {
			writeError(c, platformerrors.ErrNotFound)
			c.Abort()
			return
		}

		if !h.allowRequest(c, route) {
			c.Abort()
			return
		}

		c.Set("gateway_route", route)
		c.Next()
	}
}

func (h *Handler) allowRequest(c *gin.Context, route Route) bool {
	limit := h.limitFor(c.Request.Method, c.Request.URL.Path, route.Auth)
	identifier := c.ClientIP()
	if route.Auth {
		claims, err := parseBearer(c.GetHeader("Authorization"))
		if err != nil {
			writeError(c, platformerrors.ErrUnauthorized)
			return false
		}
		identifier = claims.Subject
		c.Set("auth_claims", claims)
		c.Request.Header.Set("X-User-ID", claims.Subject)
		c.Request.Header.Set("X-User-Role", claims.Role)
		if route.Role != "" && !platformauth.RoleAllowed(claims.Role, route.Role) {
			writeError(c, platformerrors.ErrUnauthorized)
			return false
		}
	}

	remaining, resetAt, allowed := h.checkRateLimit(c.Request.Method+" "+c.Request.URL.Path, identifier, limit)
	c.Header("X-RateLimit-Limit", intToString(limit.Requests))
	c.Header("X-RateLimit-Remaining", intToString(remaining))
	c.Header("X-RateLimit-Reset", intToString(int(resetAt.Unix())))
	if !allowed {
		c.Header("Retry-After", intToString(int(time.Until(resetAt).Seconds())))
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"data":    nil,
			"error": gin.H{
				"code":    "rate_limited",
				"message": "rate limit exceeded",
			},
			"meta": gin.H{"timestamp": time.Now().UTC()},
		})
		return false
	}

	return true
}

func (h *Handler) proxyRequest(c *gin.Context) {
	route := c.MustGet("gateway_route").(Route)
	targetURL, err := url.Parse(route.Target)
	if err != nil {
		writeError(c, platformerrors.ErrInternal)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
		req.Header.Set("X-Request-ID", c.GetString("request_id"))
		req.Header.Set("X-Trace-ID", c.Writer.Header().Get("X-Trace-ID"))
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusServiceUnavailable)
		_, _ = rw.Write([]byte(`{"success":false,"data":null,"error":{"code":"upstream_unavailable","message":"upstream unavailable"},"meta":{"version":"v1"}}`))
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) matchRoute(path string) (Route, bool) {
	for _, route := range h.routes {
		if strings.HasPrefix(path, route.Prefix) {
			return route, true
		}
	}
	return Route{}, false
}

func (h *Handler) limitFor(method, path string, authenticated bool) rateLimit {
	key := method + " " + path
	switch {
	case strings.HasPrefix(key, "POST /api/v1/auth/login"):
		return rateLimit{Requests: 5, Window: 15 * time.Minute}
	case strings.HasPrefix(key, "POST /api/v1/auth/magic-link/send"):
		return rateLimit{Requests: 3, Window: 10 * time.Minute}
	case strings.HasPrefix(key, "POST /api/v1/auth/register"):
		return rateLimit{Requests: 10, Window: time.Hour}
	case strings.HasPrefix(key, "POST /api/v1/orders"):
		return rateLimit{Requests: 20, Window: time.Minute}
	case authenticated:
		return rateLimit{Requests: 300, Window: time.Minute}
	default:
		return rateLimit{Requests: 60, Window: time.Minute}
	}
}

func (h *Handler) checkRateLimit(endpointKey, identifier string, limit rateLimit) (remaining int, resetAt time.Time, allowed bool) {
	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	now := time.Now().UTC()
	key := endpointKey + ":" + identifier
	if h.redis != nil {
		count, resetAt, err := h.redis.IncrWithTTL(context.Background(), key, limit.Window)
		if err == nil {
			if count > limit.Requests {
				return 0, resetAt, false
			}
			return limit.Requests - count, resetAt, true
		}
	}
	bucket := h.rateBuckets[key]
	if bucket.ResetAt.IsZero() || now.After(bucket.ResetAt) {
		bucket = rateBucket{Count: 0, ResetAt: now.Add(limit.Window)}
	}
	if bucket.Count >= limit.Requests {
		return 0, bucket.ResetAt, false
	}
	bucket.Count++
	h.rateBuckets[key] = bucket
	return limit.Requests - bucket.Count, bucket.ResetAt, true
}

func (h *Handler) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := ""
		for _, candidate := range h.cfg.Security.CORSAllowedOrigins {
			if origin != "" && origin == candidate {
				allowedOrigin = origin
				break
			}
		}
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Trace-ID, X-CSRF-Token")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Max-Age", "600")
		if origin != "" && allowedOrigin == "" && (c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete) {
			writeError(c, platformerrors.ErrForbidden)
			c.Abort()
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *Handler) securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func (h *Handler) maxBodySize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

func parseBearer(header string) (*platformauth.Claims, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, platformerrors.ErrUnauthorized
	}
	claims, err := platformauth.ValidateTokenWithOptions(parts[1], platformauth.ValidationOptions{
		ExpectedIssuer:   "logistics-platform",
		ExpectedAudience: "logistics-platform",
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
		"error":   nil,
		"meta":    gin.H{"timestamp": time.Now().UTC(), "version": "v1"},
	})
}

func writeError(c *gin.Context, apiErr *platformerrors.APIError) {
	c.JSON(apiErr.HTTPStatus, gin.H{
		"success": false,
		"data":    nil,
		"error":   gin.H{"code": apiErr.Code, "message": apiErr.Message},
		"meta":    gin.H{"timestamp": time.Now().UTC(), "version": "v1"},
	})
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
