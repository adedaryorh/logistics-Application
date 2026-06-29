package tracking

import "testing"

func TestHubBroadcastsToSubscribers(t *testing.T) {
	hub := NewHub()
	ch := make(chan DriverLocationMsg, 1)
	hub.Subscribe("order-1", ch)

	hub.BroadcastDriverLocation("order-1", DriverLocationMsg{DriverID: "driver-1", OrderID: "order-1"})

	select {
	case msg := <-ch:
		if msg.DriverID != "driver-1" {
			t.Fatalf("expected driver-1, got %q", msg.DriverID)
		}
	default:
		t.Fatal("expected broadcast message")
	}
}
