package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	appLogger "github.com/adedaryorh/logistics-platform/pkg/logger"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	traceIDKey   contextKey = "trace_id"
)

// RequestID generates a UUID v7 and sets it in the context and header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.NewV7()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate request id"})
			return
		}

		requestID := id.String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Logger returns a middleware that logs HTTP requests
func Logger(log *appLogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		reqLog := log.WithRequestID(c.Request.Context())
		reqLog.Info("http request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()),
		)
	}
}

// getRequestID retrieves the request ID from context
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// Tracing creates a middleware that starts OpenTelemetry spans
func Tracing(tracerProvider trace.TracerProvider) gin.HandlerFunc {
	if tracerProvider == nil {
		tracerProvider = otel.GetTracerProvider()
	}

	tracer := tracerProvider.Tracer("github.com/adedaryorh/logistics-platform")
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		ctx, span := tracer.Start(ctx, "HTTP "+c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()

		traceID := span.SpanContext().TraceID().String()
		ctx = context.WithValue(ctx, traceIDKey, traceID)
		c.Header("X-Trace-ID", traceID)
		if requestID := getRequestID(c); requestID != "" {
			span.SetAttributes(attribute.String("app.request_id", requestID))
		}

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
			attribute.String("net.peer.ip", c.ClientIP()),
			attribute.String("app.trace_id", traceID),
		)

		// Update context with span
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Set response status
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)
		observability.IncCounter("http_requests_traced_total", 1, map[string]string{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": fmt.Sprintf("%d", c.Writer.Status()),
		})
	}
}

// Recovery recovers from panics and logs the stack trace
func Recovery(log *appLogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := make([]byte, 4<<10) // 4 KB
				length := runtime.Stack(stack, true)
				log.WithRequestID(c.Request.Context()).Error("recovered from panic",
					zap.Any("error", err),
					zap.String("stack", string(stack[:length])),
					zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"data":    nil,
					"error": gin.H{
						"code":    "internal_error",
						"message": fmt.Sprintf("%v", err),
					},
					"meta": gin.H{
						"request_id": getRequestID(c),
						"timestamp":  time.Now().UTC(),
						"version":    "v1",
					},
				})
			}
		}()
		c.Next()
	}
}
