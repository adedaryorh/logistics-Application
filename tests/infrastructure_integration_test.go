package tests

import (
	"context"
	"testing"
	"time"

	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	segmentkafka "github.com/segmentio/kafka-go"
)

func TestInfrastructureIntegration(t *testing.T) {
	integrationEnabled(t)

	db := openPostgres(t)
	defer db.Close()

	ctx, cancel := testContext(t, 5*time.Second)
	defer cancel()
	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		t.Fatalf("postgres select 1: %v", err)
	}
	if one != 1 {
		t.Fatalf("expected postgres result 1, got %d", one)
	}

	redis := openRedis(t)
	key := "integration:test:key"
	if err := redis.Set(ctx, key, "ok", 5*time.Second); err != nil {
		t.Fatalf("redis set: %v", err)
	}
	value, ok, err := redis.Get(ctx, key)
	if err != nil || !ok {
		t.Fatalf("redis get: ok=%v err=%v", ok, err)
	}
	if value != "ok" {
		t.Fatalf("expected redis value ok, got %q", value)
	}

	topic := "platform.integration.infrastructure"
	payload, err := platformkafka.NewEvent("integration-test", "infra-1", "payment.completed", map[string]any{
		"transaction_id": "tx-infra-1",
		"order_id":       "order-infra-1",
		"amount_minor":   2500,
		"currency":       "NGN",
	})
	if err != nil {
		t.Fatalf("build kafka event: %v", err)
	}

	producer := platformkafka.NewProducer(kafkaBrokers())
	defer producer.Close()
	if err := producer.PublishRaw(ctx, topic, []byte("infra-1"), payload); err != nil {
		t.Fatalf("publish kafka event: %v", err)
	}
}

func TestKafkaPoisonMessageDLQ(t *testing.T) {
	integrationEnabled(t)

	topic := "platform.integration.poison"
	dlqTopic := topic + ".dead-letter"
	consumer := platformkafka.NewConsumer(kafkaBrokers())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Subscribe(ctx, topic, "integration-poison-group", func(ctx context.Context, msg segmentkafka.Message) error {
			_, err := platformkafka.DecodeEvent(msg.Value)
			return err
		})
	}()

	time.Sleep(200 * time.Millisecond)

	producer := platformkafka.NewProducer(kafkaBrokers())
	defer producer.Close()
	if err := producer.PublishRaw(context.Background(), topic, []byte("bad"), []byte(`{"not":"valid"}`)); err != nil {
		t.Fatalf("publish poison message: %v", err)
	}

	reader := newKafkaReader(t, dlqTopic, "integration-poison-reader")
	defer reader.Close()
	readCtx, readCancel := testContext(t, 10*time.Second)
	defer readCancel()
	msg, err := reader.ReadMessage(readCtx)
	if err != nil {
		t.Fatalf("read dlq message: %v", err)
	}
	if string(msg.Key) != "bad" {
		t.Fatalf("expected dlq key bad, got %q", string(msg.Key))
	}

	cancel()
	select {
	case <-time.After(2 * time.Second):
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("consumer returned error: %v", err)
		}
	}
}
