package tests

import (
	"net/http"
	"testing"
)

func TestHTTPContracts_CreateOrderAndInitializePayment(t *testing.T) {
	integrationEnabled(t)

	orderResp, orderBody := doJSON(t, http.MethodPost, logisticsBaseURL()+"/api/v1/orders", map[string]string{
		"X-User-ID": "contract-user-1",
	}, map[string]any{
		"type":            "food",
		"idempotency_key": "contract-order-1",
		"pickup":          map[string]any{"lat": 6.52, "lng": 3.37},
		"dropoff":         map[string]any{"lat": 6.53, "lng": 3.38},
		"items":           []map[string]any{{"id": "item-1", "name": "Rice", "quantity": 1, "price_minor": 1500, "currency": "NGN"}},
	})
	if orderResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected order status 201, got %d body=%v", orderResp.StatusCode, orderBody)
	}
	assertEnvelope(t, orderBody)

	data, ok := orderBody["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected order data object, got %#v", orderBody["data"])
	}
	orderObj := data["order"].(map[string]any)
	orderID := orderObj["id"].(string)

	paymentResp, paymentBody := doJSON(t, http.MethodPost, paymentBaseURL()+"/api/v1/payments/initialize", nil, map[string]any{
		"order_id":        orderID,
		"customer_id":     "contract-user-1",
		"amount_minor":    1500,
		"provider":        "paystack",
		"idempotency_key": "contract-payment-1",
		"customer_email":  "contract@example.com",
		"customer_name":   "Contract User",
	})
	if paymentResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected payment status 201, got %d body=%v", paymentResp.StatusCode, paymentBody)
	}
	assertEnvelope(t, paymentBody)
}

func assertEnvelope(t *testing.T, body map[string]any) {
	t.Helper()
	for _, key := range []string{"success", "data", "error", "meta"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("expected response envelope key %q in %#v", key, body)
		}
	}
}
