package service

import (
	"context"
	"testing"
	"time"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
)

func TestAgriculturalQuoteAndBookingAreIdempotent(t *testing.T) {
	svc := New(&platformconfig.Config{})
	now := time.Now().UTC().Add(time.Hour)
	weight := 850.0
	in := CreateAgriculturalQuoteInput{PlatformService: "farmsense", PlatformUserID: "platform-user-1", FarmSenseRequestID: "fs-1", MarketplaceRequestID: "market-1", IdempotencyKey: "quote-key", Pickup: model.Coordinate{Lat: 6.45, Lng: 3.4}, Dropoff: model.Coordinate{Lat: 6.6, Lng: 3.5}, Shipment: model.AgriculturalShipment{ProduceType: "tomatoes", Quantity: 20, QuantityUnit: "crate", WeightKG: &weight, Packaging: "crate", PickupWindow: model.TimeWindow{StartAt: now, EndAt: now.Add(time.Hour)}, DeliveryWindow: model.TimeWindow{StartAt: now.Add(2 * time.Hour), EndAt: now.Add(4 * time.Hour)}}}
	q1, err := svc.CreateAgriculturalQuote(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	q2, err := svc.CreateAgriculturalQuote(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if q1.ID != q2.ID {
		t.Fatal("quote idempotency returned a different quote")
	}
	b := CreateAgriculturalBookingInput{PlatformService: "farmsense", QuoteID: q1.ID, IdempotencyKey: "booking-key"}
	o1, err := svc.CreateAgriculturalBooking(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}
	o2, err := svc.CreateAgriculturalBooking(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}
	if o1.ID != o2.ID || o1.Type != model.OrderTypeAgricultural || o1.MarketplaceRequestID != "market-1" || o1.AgriculturalShipment == nil {
		t.Fatalf("unexpected booking: %#v", o1)
	}
}

func TestAgriculturalColdChainRequiresTemperatureRange(t *testing.T) {
	now := time.Now().UTC().Add(time.Hour)
	shipment := model.AgriculturalShipment{ProduceType: "milk", Quantity: 10, QuantityUnit: "litre", Packaging: "refrigerated_container", RequiresRefrigeration: true, PickupWindow: model.TimeWindow{StartAt: now, EndAt: now.Add(time.Hour)}, DeliveryWindow: model.TimeWindow{StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour)}}
	if validShipment(shipment) {
		t.Fatal("expected missing cold-chain temperatures to fail")
	}
	min, max := 2.0, 6.0
	shipment.ColdChainMinC = &min
	shipment.ColdChainMaxC = &max
	if !validShipment(shipment) {
		t.Fatal("expected complete cold-chain metadata to pass")
	}
}
