package kafka

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewEventRejectsMissingContractFields(t *testing.T) {
	t.Parallel()

	_, err := NewEvent("payment-service", "tx-1", "payment.completed", map[string]any{
		"transaction_id": "tx-1",
		"order_id":       "order-1",
		"amount_minor":   5000,
	})
	if err == nil {
		t.Fatal("expected contract validation error")
	}
	if !strings.Contains(err.Error(), "currency") {
		t.Fatalf("expected missing currency error, got %v", err)
	}
}

func TestDecodeEventRejectsWrongProducer(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(Event{
		EventType:     "order.created",
		AggregateID:   "order-1",
		Producer:      "payment-service",
		Schema:        "order_created.v1",
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{"order_id":"order-1","customer_id":"user-1","type":"food","status":"dispatching","price_minor":2500,"workflow_id":"wf-1","idempotency_key":"idem-1"}`),
	})
	if err != nil {
		t.Fatalf("marshal test event: %v", err)
	}

	_, err = DecodeEvent(raw)
	if err == nil {
		t.Fatal("expected producer validation error")
	}
	if !strings.Contains(err.Error(), `expected "logistics-service"`) {
		t.Fatalf("expected producer mismatch error, got %v", err)
	}
}

func TestNewEventAcceptsKnownContract(t *testing.T) {
	t.Parallel()

	raw, err := NewEvent("identity-service", "user-1", "user.registered", map[string]any{
		"user_id": "user-1",
		"email":   "person@example.com",
		"role":    "customer",
	})
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	event, err := DecodeEvent(raw)
	if err != nil {
		t.Fatalf("DecodeEvent() error = %v", err)
	}
	if event.EventType != "user.registered" {
		t.Fatalf("expected user.registered, got %q", event.EventType)
	}
}
