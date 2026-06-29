package service

import (
	"context"
	"encoding/json"
	"time"

	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
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
		observability.SetGauge("service_outbox_depth", 0, map[string]string{"service": "logistics-service"})
		return 0, nil
	}
	observability.SetGauge("service_outbox_depth", float64(len(events)), map[string]string{"service": "logistics-service"})

	published := 0
	for _, event := range events {
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "logistics-service", "status": "marshal_failed"})
			return published, err
		}
		envelope, err := platformkafka.NewEventWithPayload("logistics-service", event.AggregateID, event.EventType, payload)
		if err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "logistics-service", "status": "envelope_failed"})
			return published, err
		}
		encoded, err := json.Marshal(envelope)
		if err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "logistics-service", "status": "encode_failed"})
			return published, err
		}
		if err := publisher.PublishRaw(ctx, topicPrefix+event.EventType, []byte(event.AggregateID), encoded); err != nil {
			s.requeueOutbox(events[published:])
			observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "logistics-service", "status": "publish_failed"})
			return published, err
		}
		observability.IncCounter("service_outbox_events_total", 1, map[string]string{"service": "logistics-service", "status": "published"})
		published++
	}

	observability.SetGauge("service_outbox_depth", 0, map[string]string{"service": "logistics-service"})
	return published, nil
}

func (s *Service) popOutbox() []model.OutboxEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.outboxEvents) == 0 {
		return nil
	}

	events := append([]model.OutboxEvent(nil), s.outboxEvents...)
	s.outboxEvents = s.outboxEvents[:0]
	return events
}

func (s *Service) requeueOutbox(events []model.OutboxEvent) {
	if len(events) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.outboxEvents = append(events, s.outboxEvents...)
}
