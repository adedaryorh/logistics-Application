package kafka

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
)

func TestKafkaRealBrokerIntegration(t *testing.T) {
	brokers := os.Getenv("KAFKA_INTEGRATION_BROKERS")
	if brokers == "" {
		t.Skip("set KAFKA_INTEGRATION_BROKERS to run real Kafka integration test")
	}
	brokerList := strings.Split(brokers, ",")

	topic := "platform.integration.test"
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	producer := NewProducer(brokerList)
	defer producer.Close()

	payload := map[string]any{
		"order_id": "order-int-1",
		"status":   "created",
	}
	encoded, err := NewEvent("integration-test", "order-int-1", "order.created", payload)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}
	if err := producer.PublishRaw(ctx, topic, []byte("order-int-1"), encoded); err != nil {
		t.Fatalf("PublishRaw() error = %v", err)
	}

	reader := segmentkafka.NewReader(segmentkafka.ReaderConfig{
		Brokers: brokerList,
		// Use a dedicated consumer group against the real broker so the test
		// verifies end-to-end publish/consume behavior through Kafka itself.
		Topic:    topic,
		GroupID:  "integration-test-group",
		MinBytes: 1,
		MaxBytes: 1024 * 1024,
	})
	defer reader.Close()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	event, err := DecodeEvent(msg.Value)
	if err != nil {
		t.Fatalf("DecodeEvent() error = %v", err)
	}
	if event.EventType != "order.created" {
		t.Fatalf("expected order.created event, got %q", event.EventType)
	}

	var decoded map[string]any
	if err := json.Unmarshal(event.Payload, &decoded); err != nil {
		t.Fatalf("unmarshal payload error = %v", err)
	}
	if decoded["order_id"] != "order-int-1" {
		t.Fatalf("expected payload order_id order-int-1, got %#v", decoded["order_id"])
	}
}
