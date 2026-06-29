package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMatchRoute(t *testing.T) {
	handler := &Handler{routes: defaultRoutes()}
	route, ok := handler.matchRoute("/api/v1/orders/123")
	if !ok {
		t.Fatal("expected route match")
	}
	if route.Target != "http://logistics-service:8082" {
		t.Fatalf("unexpected target %q", route.Target)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{}
	router := gin.New()
	router.Use(handler.securityHeaders())
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected X-Frame-Options DENY, got %q", rec.Header().Get("X-Frame-Options"))
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("expected Content-Security-Policy header")
	}
}

func TestRateLimitFallback(t *testing.T) {
	handler := &Handler{
		routes:      defaultRoutes(),
		rateBuckets: map[string]rateBucket{},
	}

	limit := rateLimit{Requests: 2, Window: time.Minute}
	if _, _, allowed := handler.checkRateLimit("POST /api/v1/auth/login", "127.0.0.1", limit); !allowed {
		t.Fatal("expected first request allowed")
	}
	if _, _, allowed := handler.checkRateLimit("POST /api/v1/auth/login", "127.0.0.1", limit); !allowed {
		t.Fatal("expected second request allowed")
	}
	if _, _, allowed := handler.checkRateLimit("POST /api/v1/auth/login", "127.0.0.1", limit); allowed {
		t.Fatal("expected third request to be rate limited")
	}
}
