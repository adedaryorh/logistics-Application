package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/segmentio/kafka-go"
)

// MessageHandler handles a consumed Kafka message
type MessageHandler func(ctx context.Context, msg kafka.Message) error

// Consumer wraps a Kafka consumer
type Consumer struct {
	brokers  []string
	dlq      string
	registry *SchemaRegistryClient
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string) *Consumer {
	return &Consumer{
		brokers: brokers,
		dlq:     "dead-letter",
	}
}

func (c *Consumer) WithSchemaRegistryURL(baseURL string) *Consumer {
	c.registry = NewSchemaRegistryClient(baseURL)
	return c
}

// Subscribe starts consuming messages from the topic
func (c *Consumer) Subscribe(ctx context.Context, topic, groupID string, handler MessageHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(c.brokers...),
		Topic:    topic + "." + c.dlq,
		Balancer: &kafka.LeastBytes{},
	}
	defer dlqWriter.Close()

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("fetch Kafka message: %w", err)
		}

		var handleErr error
		backoff := 100 * time.Millisecond
		for attempt := 0; attempt < 3; attempt++ {
			handleCtx := ctx
			start := time.Now()
			if c.registry != nil {
				event, err := DecodeEvent(m.Value)
				if err != nil {
					handleErr = err
				} else if err := c.registry.ValidateEvent(ctx, event); err != nil {
					handleErr = err
				} else {
					handleCtx = observability.ContextWithTraceHeaders(ctx, event.Metadata)
					handleErr = handler(handleCtx, m)
				}
			} else {
				if event, err := DecodeEvent(m.Value); err == nil {
					handleCtx = observability.ContextWithTraceHeaders(ctx, event.Metadata)
				}
				handleErr = handler(handleCtx, m)
			}
			observability.ObserveHistogram("kafka_consumer_handler_duration_seconds", time.Since(start).Seconds(), map[string]string{
				"topic":   topic,
				"attempt": fmt.Sprintf("%d", attempt+1),
			})
			if handleErr == nil {
				observability.IncCounter("kafka_consumer_messages_total", 1, map[string]string{"topic": topic, "status": "success"})
				break
			}
			observability.IncCounter("kafka_consumer_retries_total", 1, map[string]string{"topic": topic})
			select {
			case <-ctx.Done():
				observability.IncCounter("kafka_consumer_messages_total", 1, map[string]string{"topic": topic, "status": "canceled"})
				return fmt.Errorf("handle Kafka message: %w", ctx.Err())
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		if handleErr != nil {
			observability.IncCounter("kafka_consumer_messages_total", 1, map[string]string{"topic": topic, "status": "failed"})
			m.Topic = topic + "." + c.dlq
			if err := dlqWriter.WriteMessages(ctx, m); err != nil {
				return fmt.Errorf("write DLQ message: %w", err)
			}
			observability.IncCounter("kafka_consumer_dlq_total", 1, map[string]string{"topic": topic})
		}

		if err := reader.CommitMessages(ctx, m); err != nil {
			return fmt.Errorf("commit Kafka message: %w", err)
		}
		observability.IncCounter("kafka_consumer_commits_total", 1, map[string]string{"topic": topic})
	}
}

func (c *Consumer) SubscribeAll(ctx context.Context, subscriptions ...Subscription) error {
	if len(subscriptions) == 0 {
		<-ctx.Done()
		return ctx.Err()
	}

	errCh := make(chan error, len(subscriptions))
	var wg sync.WaitGroup
	for _, subscription := range subscriptions {
		subscription := subscription
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Subscribe(ctx, subscription.Topic, subscription.GroupID, subscription.Handler); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		return ctx.Err()
	case err := <-errCh:
		return err
	case <-done:
		return nil
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return nil
}
