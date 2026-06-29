package middleware

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/gin-gonic/gin"
)

type metricKey struct {
	method string
	path   string
	status int
}

type metricValue struct {
	count          uint64
	durationMicros uint64
}

type MetricsCollector struct {
	inFlight atomic.Int64

	mu       sync.RWMutex
	requests map[metricKey]metricValue
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requests: make(map[metricKey]metricValue),
	}
}

func (m *MetricsCollector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.inFlight.Add(1)
		defer m.inFlight.Add(-1)

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		m.record(c.Request.Method, route, c.Writer.Status(), time.Since(start))
	}
}

func (m *MetricsCollector) Endpoint(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4")
		c.String(http.StatusOK, m.RenderPrometheus(service))
	}
}

func (m *MetricsCollector) record(method, path string, status int, duration time.Duration) {
	roundedMicros := uint64(duration.Microseconds())
	if roundedMicros == 0 {
		roundedMicros = 1
	}

	key := metricKey{method: method, path: path, status: status}

	m.mu.Lock()
	defer m.mu.Unlock()

	value := m.requests[key]
	value.count++
	value.durationMicros += roundedMicros
	m.requests[key] = value
}

func (m *MetricsCollector) RenderPrometheus(service string) string {
	var builder strings.Builder

	builder.WriteString("# HELP http_requests_in_flight In-flight requests for service.\n")
	builder.WriteString("# TYPE http_requests_in_flight gauge\n")
	builder.WriteString(fmt.Sprintf("http_requests_in_flight{service=%q} %d\n", service, m.inFlight.Load()))
	builder.WriteString("# HELP http_requests_total Total HTTP requests for service.\n")
	builder.WriteString("# TYPE http_requests_total counter\n")
	builder.WriteString("# HELP http_request_duration_microseconds_total Cumulative request duration in microseconds.\n")
	builder.WriteString("# TYPE http_request_duration_microseconds_total counter\n")

	m.mu.RLock()
	keys := make([]metricKey, 0, len(m.requests))
	for key := range m.requests {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].status < keys[j].status
	})

	for _, key := range keys {
		value := m.requests[key]
		labels := fmt.Sprintf("{service=%q,method=%q,path=%q,status=%q}", service, key.method, key.path, strconv.Itoa(key.status))
		builder.WriteString(fmt.Sprintf("http_requests_total%s %d\n", labels, value.count))
		builder.WriteString(fmt.Sprintf("http_request_duration_microseconds_total%s %d\n", labels, value.durationMicros))
	}
	m.mu.RUnlock()

	builder.WriteString(observability.Default().RenderPrometheus(service))

	return builder.String()
}
