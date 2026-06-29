package config

import (
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("PORT", "9099")
	t.Setenv("ENV", "test")
	t.Setenv("JAEGER_ENDPOINT", "jaeger:4317")
	t.Setenv("DEFAULT_CURRENCY", "USD")
	t.Setenv("KAFKA_ENABLED", "true")
	t.Setenv("KAFKA_TOPIC_PREFIX", "events.")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Server.Port != "9099" {
		t.Fatalf("expected port 9099, got %s", cfg.Server.Port)
	}
	if cfg.Server.Environment != "test" {
		t.Fatalf("expected env test, got %s", cfg.Server.Environment)
	}
	if cfg.Telemetry.JaegerEndpoint != "jaeger:4317" {
		t.Fatalf("expected jaeger endpoint jaeger:4317, got %s", cfg.Telemetry.JaegerEndpoint)
	}
	if cfg.Money.DefaultCurrency != "USD" {
		t.Fatalf("expected default currency USD, got %s", cfg.Money.DefaultCurrency)
	}
	if !cfg.Kafka.Enabled {
		t.Fatal("expected Kafka to be enabled from env")
	}
	if cfg.Kafka.TopicPrefix != "events." {
		t.Fatalf("expected Kafka topic prefix events., got %s", cfg.Kafka.TopicPrefix)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}

	content := []byte("PORT: \"7070\"\nENV: \"staging\"\n")
	if err := os.WriteFile(file.Name(), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cfg, err := LoadConfig(file.Name())
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Server.Port != "7070" {
		t.Fatalf("expected port 7070, got %s", cfg.Server.Port)
	}
	if cfg.Server.Environment != "staging" {
		t.Fatalf("expected env staging, got %s", cfg.Server.Environment)
	}
}
