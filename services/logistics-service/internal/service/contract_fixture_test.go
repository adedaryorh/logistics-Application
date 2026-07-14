package service

import (
	"encoding/json"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
	"os"
	"reflect"
	"testing"
)

func TestCanonicalLifecycleFixture(t *testing.T) {
	body, err := os.ReadFile("../../../../contracts/v1/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Statuses  []string `json:"lifecycle"`
		Logistics struct {
			Quote   string `json:"quote"`
			Booking string `json:"booking"`
			Status  string `json:"status"`
		} `json:"logistics"`
	}
	if json.Unmarshal(body, &fixture) != nil {
		t.Fatal("invalid lifecycle fixture")
	}
	want := []string{"requested", "quoted", "confirmed", "booked", "provider_assigned", "picked_up", "in_transit", "delivered", "cancelled", "failed", "disputed"}
	if !reflect.DeepEqual(fixture.Statuses, want) {
		t.Fatalf("contract lifecycle drift: %v", fixture.Statuses)
	}
	if fixture.Logistics.Quote != "POST /internal/v1/agricultural/quotes" ||
		fixture.Logistics.Booking != "POST /internal/v1/agricultural/bookings" ||
		fixture.Logistics.Status != "GET /internal/v1/agricultural/bookings/{logistics_delivery_id}" {
		t.Fatalf("contract route drift: %+v", fixture.Logistics)
	}
	mapped := map[model.OrderStatus]string{model.OrderStatusPending: "requested", model.OrderStatusAwaitingPayment: "quoted", model.OrderStatusPaid: "confirmed", model.OrderStatusDispatching: "booked", model.OrderStatusAssigned: "provider_assigned", model.OrderStatusPickedUp: "picked_up", model.OrderStatusDelivered: "delivered", model.OrderStatusCancelled: "cancelled", model.OrderStatusFailed: "failed"}
	for internal, canonical := range mapped {
		if got := NormalizedStatus(internal); got != canonical {
			t.Fatalf("%s mapped to %s", internal, got)
		}
	}
}

func TestCanonicalFixturesAreValidJSON(t *testing.T) {
	for _, name := range []string{"service_request_create.json", "logistics_quote_request.json", "logistics_booking_request.json", "event_envelope.json"} {
		body, err := os.ReadFile("../../../../contracts/v1/" + name)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if json.Unmarshal(body, &value) != nil {
			t.Fatalf("invalid fixture %s", name)
		}
	}
}
