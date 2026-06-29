package tests

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"testing"
	"time"
)

func TestOrderToPaymentWorkflowE2E(t *testing.T) {
	integrationEnabled(t)

	orderResp, orderBody := doJSON(t, http.MethodPost, logisticsBaseURL()+"/api/v1/orders", map[string]string{
		"X-User-ID": "e2e-user-1",
	}, map[string]any{
		"type":            "food",
		"idempotency_key": "e2e-order-1",
		"pickup":          map[string]any{"lat": 6.52, "lng": 3.37},
		"dropoff":         map[string]any{"lat": 6.53, "lng": 3.38},
		"items":           []map[string]any{{"id": "item-1", "name": "Rice", "quantity": 1, "price_minor": 2500, "currency": "NGN"}},
	})
	if orderResp.StatusCode != http.StatusCreated {
		t.Fatalf("create order failed: status=%d body=%v", orderResp.StatusCode, orderBody)
	}
	orderID := orderBody["data"].(map[string]any)["order"].(map[string]any)["id"].(string)

	initResp, initBody := doJSON(t, http.MethodPost, paymentBaseURL()+"/api/v1/payments/initialize", nil, map[string]any{
		"order_id":        orderID,
		"customer_id":     "e2e-user-1",
		"amount_minor":    2500,
		"provider":        "paystack",
		"idempotency_key": "e2e-payment-1",
		"customer_email":  "e2e@example.com",
		"customer_name":   "E2E User",
	})
	if initResp.StatusCode != http.StatusCreated {
		t.Fatalf("initialize payment failed: status=%d body=%v", initResp.StatusCode, initBody)
	}
	payment := initBody["data"].(map[string]any)["payment"].(map[string]any)
	paymentID := payment["id"].(string)
	providerTxID := payment["provider_tx_id"].(string)

	webhookPayload := `{"event_id":"evt-e2e-1","event_type":"payment.success","provider_tx_id":"` + providerTxID + `","amount_minor":2500,"currency":"NGN"}`
	mac := hmac.New(sha512.New, []byte("paystack-secret"))
	_, _ = mac.Write([]byte(webhookPayload))
	signature := hex.EncodeToString(mac.Sum(nil))
	webhookResp, webhookBody := doJSON(t, http.MethodPost, paymentBaseURL()+"/api/v1/webhooks/paystack", map[string]string{
		"X-Signature": signature,
	}, map[string]any{
		"event_id":       "evt-e2e-1",
		"event_type":     "payment.success",
		"provider_tx_id": providerTxID,
		"amount_minor":   2500,
		"currency":       "NGN",
	})
	if webhookResp.StatusCode != http.StatusOK {
		t.Fatalf("webhook failed: status=%d body=%v", webhookResp.StatusCode, webhookBody)
	}

	time.Sleep(200 * time.Millisecond)

	statusResp, statusBody := doJSON(t, http.MethodGet, paymentBaseURL()+"/api/v1/payments/"+paymentID, nil, nil)
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("get payment failed: status=%d body=%v", statusResp.StatusCode, statusBody)
	}
	status := statusBody["data"].(map[string]any)["payment"].(map[string]any)["status"].(string)
	if status != "completed" {
		t.Fatalf("expected completed payment status, got %q", status)
	}
}
