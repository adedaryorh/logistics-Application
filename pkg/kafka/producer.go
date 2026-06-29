package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

// Producer wraps a Kafka producer
type Producer struct {
	w        *kafka.Writer
	registry *SchemaRegistryClient
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) *Producer {
	return &Producer{
		w: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) WithSchemaRegistryURL(baseURL string) *Producer {
	p.registry = NewSchemaRegistryClient(baseURL)
	return p
}

// Publish sends a message to the Kafka topic
func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value proto.Message) error {
	val, err := proto.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal Kafka message: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: val,
		Time:  time.Now().UTC(),
	}

	return p.writeWithRetry(ctx, msg)
}

func (p *Producer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	for _, msg := range msgs {
		if err := p.writeWithRetry(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (p *Producer) PublishRaw(ctx context.Context, topic string, key, value []byte) error {
	value = enrichMessageTrace(ctx, value)
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now().UTC(),
	}
	return p.writeWithRetry(ctx, msg)
}

func (p *Producer) writeWithRetry(ctx context.Context, msg kafka.Message) error {
	msg.Value = enrichMessageTrace(ctx, msg.Value)
	if event, err := DecodeEvent(msg.Value); err == nil && p.registry != nil {
		if err := p.registry.ValidateEvent(ctx, event); err != nil {
			observability.IncCounter("kafka_producer_messages_total", 1, map[string]string{"topic": msg.Topic, "status": "schema_failed"})
			return fmt.Errorf("validate Kafka event with schema registry: %w", err)
		}
	}

	backoff := 100 * time.Millisecond
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		start := time.Now()
		writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		lastErr = p.w.WriteMessages(writeCtx, msg)
		cancel()
		observability.ObserveHistogram("kafka_producer_duration_seconds", time.Since(start).Seconds(), map[string]string{
			"topic":   msg.Topic,
			"attempt": fmt.Sprintf("%d", attempt+1),
		})
		if lastErr == nil {
			observability.IncCounter("kafka_producer_messages_total", 1, map[string]string{"topic": msg.Topic, "status": "success"})
			return nil
		}
		observability.IncCounter("kafka_producer_retries_total", 1, map[string]string{"topic": msg.Topic})

		select {
		case <-ctx.Done():
			observability.IncCounter("kafka_producer_messages_total", 1, map[string]string{"topic": msg.Topic, "status": "canceled"})
			return fmt.Errorf("publish Kafka message: %w", ctx.Err())
		case <-time.After(backoff):
		}
		backoff *= 2
	}

	observability.IncCounter("kafka_producer_messages_total", 1, map[string]string{"topic": msg.Topic, "status": "failed"})
	return fmt.Errorf("publish Kafka message after retries: %w", lastErr)
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.w.Close()
}

func enrichMessageTrace(ctx context.Context, value []byte) []byte {
	event, err := DecodeEvent(value)
	if err != nil {
		return value
	}
	event = event.WithMetadata(observability.TraceHeadersFromContext(ctx))
	if traceID := observability.TraceIDFromContext(ctx); traceID != "" {
		event = event.WithMetadata(map[string]string{"trace_id": traceID})
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		return value
	}
	return encoded
}
