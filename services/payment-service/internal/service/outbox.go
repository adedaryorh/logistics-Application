package service

import (
	"context"
	"encoding/json"
	"time"

	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
)

type EventPublisher interface {
	PublishRaw(ctx context.Context, topic string, key, value []byte) error
}

func (s *Service) StartOutboxWorker(ctx context.Context, publisher EventPublisher, pollInterval time.Duration, topicPrefix string) error {
	if publisher == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if _, err := s.DrainOutbox(ctx, publisher, topicPrefix); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Service) DrainOutbox(ctx context.Context, publisher EventPublisher, topicPrefix string) (int, error) {
	events := s.popOutbox()
	if len(events) == 0 {
		observability.SetGauge("service_outbox_depth", 0, map[string]string{"service": "payment-service"})
		return 0, nil
	}
	observability.SetGauge("service_outbox_depth", float64(len(events)), map[string]string{"service": "payment-service"})

	published := 0
	for _, event := range events {
		payload, err := json.Marshal(event["payload"])
		if err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "payment-service", "status": "marshal_failed"})
			return published, err
		}
		envelope, err := platformkafka.NewEventWithPayload(
			"payment-service",
			firstNonEmpty(
				stringKey(event["aggregate_id"]),
				stringKey(event["transaction_id"]),
				stringKey(event["order_id"]),
				stringKey(event["user_id"]),
			),
			stringKey(event["event_type"]),
			payload,
		)
		if err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "payment-service", "status": "envelope_failed"})
			return published, err
		}
		encoded, err := json.Marshal(envelope)
		if err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "payment-service", "status": "encode_failed"})
			return published, err
		}
		nestedPayload, _ := event["payload"].(map[string]any)
		key := firstNonEmpty(
			stringKey(event["aggregate_id"]),
			stringKey(event["transaction_id"]),
			stringKey(event["order_id"]),
			stringKey(event["user_id"]),
			stringKey(nestedPayload["transaction_id"]),
			stringKey(nestedPayload["order_id"]),
			stringKey(nestedPayload["user_id"]),
		)
		if err := publisher.PublishRaw(ctx, topicPrefix+stringKey(event["event_type"]), []byte(key), encoded); err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "payment-service", "status": "publish_failed"})
			return published, err
		}
		observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "payment-service", "status": "published"})
		published++
	}

	observability.SetGauge("service_outbox_depth", 0, map[string]string{"service": "payment-service"})
	return published, nil
}

func (s *Service) popOutbox() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.outbox) == 0 {
		return nil
	}

	events := append([]map[string]any(nil), s.outbox...)
	s.outbox = s.outbox[:0]
	return events
}

func (s *Service) requeueOutbox(events []map[string]any) {
	if len(events) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.outbox = append(events, s.outbox...)
}

func stringKey(value any) string {
	asString, _ := value.(string)
	return asString
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "event"
}
