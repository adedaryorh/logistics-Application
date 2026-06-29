package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	platformredis "github.com/adedaryorh/logistics-platform/pkg/redis"
	_ "github.com/lib/pq"
	segmentkafka "github.com/segmentio/kafka-go"
)

func integrationEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_INTEGRATION") != "1" {
		t.Skip("set TEST_INTEGRATION=1 to run integration tests")
	}
}

func testContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

func postgresDSN() string {
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://postgres:test@localhost:5432/logistics?sslmode=disable"
}

func gatewayBaseURL() string {
	if v := os.Getenv("TEST_GATEWAY_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func logisticsBaseURL() string {
	if v := os.Getenv("TEST_LOGISTICS_URL"); v != "" {
		return v
	}
	return "http://localhost:8082"
}

func paymentBaseURL() string {
	if v := os.Getenv("TEST_PAYMENT_URL"); v != "" {
		return v
	}
	return "http://localhost:8084"
}

func kafkaBrokers() []string {
	if v := os.Getenv("KAFKA_INTEGRATION_BROKERS"); v != "" {
		return strings.Split(v, ",")
	}
	return []string{"localhost:9092"}
}

func internalToken() string {
	if v := os.Getenv("TEST_INTERNAL_AUTH_TOKEN"); v != "" {
		return v
	}
	return "internal-dev-token"
}

func openPostgres(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", postgresDSN())
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	ctx, cancel := testContext(t, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	return db
}

func openRedis(t *testing.T) platformredis.Store {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	store := platformredis.NewClient(platformredis.Options{Addr: addr})
	ctx, cancel := testContext(t, 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	return store
}

func doJSON(t *testing.T, method, url string, headers map[string]string, body any) (*http.Response, map[string]any) {
	t.Helper()
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("")
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = strings.NewReader(string(payload))
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	var decoded map[string]any
	raw, _ := io.ReadAll(resp.Body)
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &decoded)
	}
	return resp, decoded
}

func newKafkaReader(t *testing.T, topic, groupID string) *segmentkafka.Reader {
	t.Helper()
	return segmentkafka.NewReader(segmentkafka.ReaderConfig{
		Brokers:  kafkaBrokers(),
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 1024 * 1024,
	})
}
