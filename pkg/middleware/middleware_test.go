package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformlogger "github.com/adedaryorh/logistics-platform/pkg/logger"
	"github.com/gin-gonic/gin"
)

func TestRequestIDSetsHeader(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)
	engine.Use(RequestID())
	engine.GET("/", func(c *gin.Context) {
		if c.GetString("request_id") == "" {
			t.Fatal("expected request_id in context")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestRecoveryHandlesPanic(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	log, err := platformlogger.New("svc", "test", "v1")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)
	engine.Use(RequestID(), Recovery(log))
	engine.GET("/", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestMetricsCollectorTracksRequests(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	collector := NewMetricsCollector()

	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)
	engine.Use(collector.Middleware())
	engine.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/123", nil)
	engine.ServeHTTP(recorder, req)

	body := collector.RenderPrometheus("logistics-service")
	if !strings.Contains(body, `http_requests_total{service="logistics-service",method="GET",path="/orders/:id",status="202"} 1`) {
		t.Fatalf("expected route-based request counter, got %s", body)
	}
	if !strings.Contains(body, `http_request_duration_microseconds_total{service="logistics-service",method="GET",path="/orders/:id",status="202"}`) {
		t.Fatalf("expected duration counter, got %s", body)
	}
}

func TestMetricsEndpointRendersPrometheus(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	collector := NewMetricsCollector()

	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)
	engine.Use(collector.Middleware())
	engine.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	engine.GET("/metrics", collector.Endpoint("identity-service"))

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	engine.ServeHTTP(recorder, req)

	body := recorder.Body.String()
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", contentType)
	}
	if !strings.Contains(body, `http_requests_total{service="identity-service",method="GET",path="/healthz",status="200"} 1`) {
		t.Fatalf("expected health metric in body, got %s", body)
	}
}
