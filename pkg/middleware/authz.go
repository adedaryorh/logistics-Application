package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/gin-gonic/gin"
)

type anomalyRecord struct {
	Count   int
	ResetAt time.Time
	BlockTo time.Time
}

var (
	anomalyMu   sync.Mutex
	anomalyByIP = map[string]anomalyRecord{}
)

func RequireBearer(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseBearerClaims(c.Request.Context(), c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(platformerrors.ErrUnauthorized.HTTPStatus, errorBody(c, platformerrors.ErrUnauthorized))
			return
		}
		if len(roles) > 0 && !allowedAnyRole(claims.Role, roles...) {
			c.AbortWithStatusJSON(platformerrors.ErrForbidden.HTTPStatus, errorBody(c, platformerrors.ErrForbidden))
			return
		}
		c.Set("auth_claims", claims)
		c.Set("auth_subject", claims.Subject)
		c.Next()
	}
}

func InternalOnly(cfg *platformconfig.Config, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil || strings.TrimSpace(cfg.Security.InternalToken) == "" {
			c.AbortWithStatusJSON(platformerrors.ErrInternal.HTTPStatus, errorBody(c, platformerrors.ErrInternal))
			return
		}
		if subtleHeaderMismatch(c.GetHeader("X-Internal-Token"), cfg.Security.InternalToken) {
			c.AbortWithStatusJSON(platformerrors.ErrUnauthorized.HTTPStatus, errorBody(c, platformerrors.ErrUnauthorized))
			return
		}
		if cfg.Security.MTLSRequired {
			if c.Request.TLS == nil || len(c.Request.TLS.VerifiedChains) == 0 {
				if !strings.EqualFold(c.GetHeader(cfg.Security.MTLSClientHeader), "true") {
					c.AbortWithStatusJSON(platformerrors.ErrUnauthorized.HTTPStatus, errorBody(c, platformerrors.ErrUnauthorized))
					return
				}
			}
		}
		if len(roles) > 0 {
			role := c.GetHeader("X-Internal-Role")
			if !allowedAnyRole(role, roles...) {
				c.AbortWithStatusJSON(platformerrors.ErrForbidden.HTTPStatus, errorBody(c, platformerrors.ErrForbidden))
				return
			}
		}
		c.Next()
	}
}

func AbuseProtection(cfg *platformconfig.Config) gin.HandlerFunc {
	threshold := 8
	window := 5 * time.Minute
	if cfg != nil {
		if cfg.Security.AnomalyThreshold > 0 {
			threshold = cfg.Security.AnomalyThreshold
		}
		if cfg.Security.AnomalyWindowSeconds > 0 {
			window = time.Duration(cfg.Security.AnomalyWindowSeconds) * time.Second
		}
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now().UTC()

		anomalyMu.Lock()
		record := anomalyByIP[ip]
		if !record.BlockTo.IsZero() && now.Before(record.BlockTo) {
			anomalyMu.Unlock()
			c.AbortWithStatusJSON(platformerrors.ErrRateLimit.HTTPStatus, errorBody(c, platformerrors.ErrRateLimit))
			return
		}
		if record.ResetAt.IsZero() || now.After(record.ResetAt) {
			record = anomalyRecord{ResetAt: now.Add(window)}
		}
		anomalyByIP[ip] = record
		anomalyMu.Unlock()

		c.Next()

		status := c.Writer.Status()
		if status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusTooManyRequests {
			anomalyMu.Lock()
			record := anomalyByIP[ip]
			if record.ResetAt.IsZero() || now.After(record.ResetAt) {
				record = anomalyRecord{ResetAt: now.Add(window)}
			}
			record.Count++
			if record.Count >= threshold {
				record.BlockTo = now.Add(window)
			}
			anomalyByIP[ip] = record
			anomalyMu.Unlock()
		}
	}
}

func parseBearerClaims(ctx context.Context, header string) (*platformauth.Claims, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, platformerrors.ErrUnauthorized
	}
	return platformauth.ValidateTokenWithOptions(parts[1], platformauth.ValidationOptions{
		ExpectedIssuer:   "logistics-platform",
		ExpectedAudience: "logistics-platform",
	})
}

func allowedAnyRole(role string, allowed ...string) bool {
	return platformauth.RoleAllowed(role, allowed...)
}

func subtleHeaderMismatch(got, expected string) bool {
	return strings.TrimSpace(got) == "" || strings.TrimSpace(got) != strings.TrimSpace(expected)
}

func errorBody(c *gin.Context, apiErr *platformerrors.APIError) gin.H {
	return gin.H{
		"success": false,
		"data":    nil,
		"error": gin.H{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		},
		"meta": gin.H{
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now().UTC(),
			"version":    "v1",
		},
	}
}
