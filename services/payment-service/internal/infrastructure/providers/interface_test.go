package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCircuitBreakerOpensAfterFiveFailures(t *testing.T) {
	provider := &BaseProvider{name: "test"}

	for i := 0; i < 5; i++ {
		if _, err := provider.Initialize(context.Background(), InitializeRequest{}); err == nil {
			t.Fatal("expected initialize failure")
		}
	}

	if _, err := provider.Initialize(context.Background(), InitializeRequest{
		TransactionID: "tx-1",
		AmountMinor:   1000,
		Currency:      "NGN",
	}); err == nil {
		t.Fatal("expected circuit breaker error")
	}
}

func TestInitializeUsesLiveHTTPAndIdempotencyCache(t *testing.T) {
	serverCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalls++
		if got := r.Header.Get("Idempotency-Key"); got != "idem-live" {
			t.Fatalf("expected idempotency header idem-live, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"provider_tx_id":"live-tx-1","checkout_url":"https://example.com/checkout","access_code":"access-1"}`))
	}))
	defer server.Close()

	provider := &BaseProvider{
		name:          "stripe",
		secret:        "stripe-secret",
		apiKey:        "sk_test",
		baseURL:       server.URL,
		initializeURL: "/payments",
		httpClient:    server.Client(),
		initCache:     map[string]InitializeResponse{},
		refundCache:   map[string]RefundResponse{},
		payoutCache:   map[string]PayoutResponse{},
	}

	first, err := provider.Initialize(context.Background(), InitializeRequest{
		TransactionID:  "tx-1",
		AmountMinor:    1000,
		Currency:       "NGN",
		CustomerEmail:  "person@example.com",
		IdempotencyKey: "idem-live",
	})
	if err != nil {
		t.Fatalf("Initialize() first error = %v", err)
	}
	second, err := provider.Initialize(context.Background(), InitializeRequest{
		TransactionID:  "tx-1",
		AmountMinor:    1000,
		Currency:       "NGN",
		CustomerEmail:  "person@example.com",
		IdempotencyKey: "idem-live",
	})
	if err != nil {
		t.Fatalf("Initialize() second error = %v", err)
	}
	if first.ProviderTxID != second.ProviderTxID {
		t.Fatalf("expected cached response match, got %q and %q", first.ProviderTxID, second.ProviderTxID)
	}
	if serverCalls != 1 {
		t.Fatalf("expected exactly one live provider call, got %d", serverCalls)
	}
}

func TestParseWebhookSupportsProviderPayloadVariants(t *testing.T) {
	provider := &BaseProvider{name: "stripe", secret: "stripe-secret"}
	payload := []byte(`{"id":"evt_1","type":"payment_intent.succeeded","data":{"object":{"id":"pi_1"},"amount":2500,"currency":"NGN"}}`)
	mac := hmac.New(sha256.New, []byte("stripe-secret"))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	event, err := provider.ParseWebhook(payload, signature)
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if event.EventType != "payment.completed" {
		t.Fatalf("expected payment.completed, got %q", event.EventType)
	}
	if event.ProviderTxID != "pi_1" {
		t.Fatalf("expected provider tx id pi_1, got %q", event.ProviderTxID)
	}
}

func TestParseWebhook_VerifiesSignature(t *testing.T) {
	payload := []byte(`{"event_id":"evt-1","event_type":"payment.success","provider_tx_id":"tx-1","amount_minor":1000,"currency":"NGN"}`)

	flutterwave := &BaseProvider{name: "flutterwave", secret: "flutterwave-secret"}
	if _, err := flutterwave.ParseWebhook(payload, "flutterwave-secret"); err != nil {
		t.Fatalf("flutterwave ParseWebhook() error = %v", err)
	}
	if _, err := flutterwave.ParseWebhook(payload, "wrong"); err == nil {
		t.Fatal("expected flutterwave invalid signature error")
	}

	paystackSig := func() string {
		mac := hmac.New(sha512.New, []byte("paystack-secret"))
		mac.Write(payload)
		return hex.EncodeToString(mac.Sum(nil))
	}()
	paystack := &BaseProvider{name: "paystack", secret: "paystack-secret"}
	if _, err := paystack.ParseWebhook(payload, paystackSig); err != nil {
		t.Fatalf("paystack ParseWebhook() error = %v", err)
	}

	stripeSig := func() string {
		mac := hmac.New(sha256.New, []byte("stripe-secret"))
		mac.Write(payload)
		return hex.EncodeToString(mac.Sum(nil))
	}()
	stripe := &BaseProvider{name: "stripe", secret: "stripe-secret"}
	if _, err := stripe.ParseWebhook(payload, stripeSig); err != nil {
		t.Fatalf("stripe ParseWebhook() error = %v", err)
	}
}
