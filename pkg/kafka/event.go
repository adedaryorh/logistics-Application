package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/observability"
)

const DefaultSchemaVersion = 1

type Event struct {
	EventType     string            `json:"event_type"`
	AggregateID   string            `json:"aggregate_id,omitempty"`
	Producer      string            `json:"producer"`
	Schema        string            `json:"schema"`
	SchemaVersion int               `json:"schema_version"`
	OccurredAt    time.Time         `json:"occurred_at"`
	Payload       json.RawMessage   `json:"payload"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type Subscription struct {
	Topic   string
	GroupID string
	Handler MessageHandler
}

func NewEvent(producer, aggregateID, eventType string, payload any) ([]byte, error) {
	if strings.TrimSpace(producer) == "" {
		return nil, fmt.Errorf("producer is required")
	}
	if strings.TrimSpace(eventType) == "" {
		return nil, fmt.Errorf("event type is required")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal event payload: %w", err)
	}

	envelope, err := NewEventWithPayload(producer, aggregateID, eventType, payloadBytes)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal event envelope: %w", err)
	}
	return encoded, nil
}

func NewEventFromContext(ctx context.Context, producer, aggregateID, eventType string, payload any) ([]byte, error) {
	encoded, err := NewEvent(producer, aggregateID, eventType, payload)
	if err != nil {
		return nil, err
	}
	event, err := DecodeEvent(encoded)
	if err != nil {
		return nil, err
	}
	event.Metadata = mergeMetadata(event.Metadata, observability.TraceHeadersFromContext(ctx))
	encoded, err = json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal contextual event envelope: %w", err)
	}
	return encoded, nil
}

func NewEventWithPayload(producer, aggregateID, eventType string, payload []byte) (Event, error) {
	if strings.TrimSpace(producer) == "" {
		return Event{}, fmt.Errorf("producer is required")
	}
	if strings.TrimSpace(eventType) == "" {
		return Event{}, fmt.Errorf("event type is required")
	}
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}

	event := Event{
		EventType:     eventType,
		AggregateID:   aggregateID,
		Producer:      producer,
		Schema:        schemaName(eventType),
		SchemaVersion: DefaultSchemaVersion,
		OccurredAt:    time.Now().UTC(),
		Payload:       append(json.RawMessage(nil), payload...),
	}
	if err := ValidateEventContract(event); err != nil {
		return Event{}, fmt.Errorf("build event envelope: %w", err)
	}
	return event, nil
}

func (e Event) WithMetadata(metadata map[string]string) Event {
	e.Metadata = mergeMetadata(e.Metadata, metadata)
	return e
}

func DecodeEvent(data []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return Event{}, fmt.Errorf("decode event envelope: %w", err)
	}
	if strings.TrimSpace(event.EventType) == "" {
		return Event{}, fmt.Errorf("decode event envelope: missing event_type")
	}
	if strings.TrimSpace(event.Schema) == "" {
		return Event{}, fmt.Errorf("decode event envelope: missing schema")
	}
	if event.SchemaVersion != DefaultSchemaVersion {
		return Event{}, fmt.Errorf("decode event envelope: unsupported schema version %d", event.SchemaVersion)
	}
	if len(event.Payload) == 0 {
		return Event{}, fmt.Errorf("decode event envelope: missing payload")
	}
	if expected := schemaName(event.EventType); event.Schema != expected {
		return Event{}, fmt.Errorf("decode event envelope: schema %q does not match expected %q", event.Schema, expected)
	}
	if err := ValidateEventContract(event); err != nil {
		return Event{}, fmt.Errorf("decode event envelope: %w", err)
	}
	return event, nil
}

func schemaName(eventType string) string {
	name := strings.TrimSpace(strings.ReplaceAll(eventType, ".", "_"))
	if name == "" {
		return "event"
	}
	return name + ".v1"
}

func mergeMetadata(existing, extra map[string]string) map[string]string {
	if len(existing) == 0 && len(extra) == 0 {
		return nil
	}
	merged := make(map[string]string, len(existing)+len(extra))
	for key, value := range existing {
		merged[key] = value
	}
	for key, value := range extra {
		if strings.TrimSpace(value) == "" {
			continue
		}
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
