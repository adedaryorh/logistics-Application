package service

import (
	"context"
	"encoding/json"
	"testing"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
)

type stubPublisher struct {
	messages []publishedMessage
}

type publishedMessage struct {
	topic string
	key   string
	value []byte
}

func (p *stubPublisher) PublishRaw(_ context.Context, topic string, key, value []byte) error {
	p.messages = append(p.messages, publishedMessage{topic: topic, key: string(key), value: append([]byte(nil), value...)})
	return nil
}

func TestCreateOrder_Idempotency(t *testing.T) {
	svc := New(&platformconfig.Config{})
	input := CreateOrderInput{
		CustomerID:     "customer-1",
		Type:           "ride",
		Pickup:         model.Coordinate{Lat: 6.5, Lng: 3.3},
		Dropoff:        model.Coordinate{Lat: 6.6, Lng: 3.4},
		IdempotencyKey: "idem-1",
	}

	first, err := svc.CreateOrder(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateOrder() first error = %v", err)
	}
	second, err := svc.CreateOrder(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateOrder() second error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same order ID for idempotent request, got %q and %q", first.ID, second.ID)
	}
}

func TestRespondToAssignment_Accept(t *testing.T) {
	svc := New(&platformconfig.Config{})
	driver, err := svc.CreateDriver(context.Background(), DriverOnboardingInput{
		UserID:   "driver-user-1",
		FullName: "Jane Doe",
		Phone:    "+2348000000000",
		Type:     "ride",
		Lat:      6.5,
		Lng:      3.3,
	})
	if err != nil {
		t.Fatalf("CreateDriver() error = %v", err)
	}
	if _, err := svc.SetDriverAvailability(context.Background(), DriverAvailabilityInput{UserID: driver.UserID, Online: true}); err != nil {
		t.Fatalf("SetDriverAvailability() error = %v", err)
	}

	order, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		CustomerID:     "customer-1",
		Type:           "ride",
		Pickup:         model.Coordinate{Lat: 6.5, Lng: 3.3},
		Dropoff:        model.Coordinate{Lat: 6.51, Lng: 3.31},
		IdempotencyKey: "idem-accept",
	})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	assignments := svc.orderAssignments[order.ID]
	if len(assignments) == 0 {
		t.Fatal("expected seeded assignments")
	}

	accepted, err := svc.RespondToAssignment(context.Background(), AssignmentResponseInput{
		AssignmentID: assignments[0].ID,
		DriverUserID: driver.UserID,
		Accepted:     true,
	})
	if err != nil {
		t.Fatalf("RespondToAssignment() error = %v", err)
	}
	if accepted.Status != model.AssignmentStatusAccepted {
		t.Fatalf("expected accepted status, got %q", accepted.Status)
	}
	gotOrder, err := svc.GetOrder(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if gotOrder.Status != model.OrderStatusAssigned {
		t.Fatalf("expected order to be assigned, got %q", gotOrder.Status)
	}
}

func TestDrainOutboxPublishesOrderEvents(t *testing.T) {
	svc := New(&platformconfig.Config{})
	publisher := &stubPublisher{}

	order, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		CustomerID:     "customer-1",
		Type:           "ride",
		Pickup:         model.Coordinate{Lat: 6.5, Lng: 3.3},
		Dropoff:        model.Coordinate{Lat: 6.6, Lng: 3.4},
		IdempotencyKey: "idem-outbox",
	})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	count, err := svc.DrainOutbox(context.Background(), publisher, "platform.")
	if err != nil {
		t.Fatalf("DrainOutbox() error = %v", err)
	}
	if count == 0 {
		t.Fatal("expected published outbox events")
	}
	if publisher.messages[0].topic != "platform.order.created" {
		t.Fatalf("expected first topic platform.order.created, got %s", publisher.messages[0].topic)
	}
	if publisher.messages[0].key != order.ID {
		t.Fatalf("expected order key %s, got %s", order.ID, publisher.messages[0].key)
	}

	var payload map[string]any
	if err := json.Unmarshal(publisher.messages[0].value, &payload); err != nil {
		t.Fatalf("unmarshal published payload: %v", err)
	}
	if payload["aggregate_id"] != order.ID {
		t.Fatalf("expected aggregate_id %s, got %v", order.ID, payload["aggregate_id"])
	}
}
