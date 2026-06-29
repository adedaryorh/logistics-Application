package kafka

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSchemaRegistryClientValidateEvent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subjects/order_created.v1/versions/1" {
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subject":"order_created.v1","version":1}`))
	}))
	defer server.Close()

	event, err := NewEventWithPayload("logistics-service", "order-1", "order.created", []byte(`{"order_id":"order-1","customer_id":"user-1","type":"food","status":"dispatching","price_minor":2500,"workflow_id":"wf-1","idempotency_key":"idem-1"}`))
	if err != nil {
		t.Fatalf("NewEventWithPayload() error = %v", err)
	}

	client := NewSchemaRegistryClient(server.URL)
	if err := client.ValidateEvent(context.Background(), event); err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}
}

func TestSchemaRegistryClientValidateEventMissingSubject(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	event, err := NewEventWithPayload("payment-service", "tx-1", "payment.completed", []byte(`{"transaction_id":"tx-1","order_id":"order-1","amount_minor":5000,"currency":"NGN"}`))
	if err != nil {
		t.Fatalf("NewEventWithPayload() error = %v", err)
	}

	client := NewSchemaRegistryClient(server.URL)
	if err := client.ValidateEvent(context.Background(), event); err == nil {
		t.Fatal("expected schema registry error")
	}
}
