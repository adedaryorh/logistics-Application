package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
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

func TestWebhookDeduplication(t *testing.T) {
	svc := New(&platformconfig.Config{})
	tx, err := svc.InitializePayment(context.Background(), InitializeInput{
		OrderID:        "order-1",
		CustomerID:     "customer-1",
		AmountMinor:    2000,
		Currency:       "NGN",
		Provider:       "paystack",
		IdempotencyKey: "idem-1",
		CustomerEmail:  "customer@example.com",
		CustomerName:   "Customer One",
	})
	if err != nil {
		t.Fatalf("InitializePayment() error = %v", err)
	}

	payload := []byte(`{"event_id":"evt-1","event_type":"payment.success","provider_tx_id":"` + tx.ProviderTxID + `","amount_minor":2000,"currency":"NGN"}`)
	mac := hmac.New(sha512.New, []byte("paystack-secret"))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	processed, err := svc.HandleWebhook(context.Background(), WebhookInput{
		Provider:  "paystack",
		Payload:   payload,
		Signature: signature,
	})
	if err != nil || !processed {
		t.Fatalf("expected first webhook to process, processed=%v err=%v", processed, err)
	}

	processed, err = svc.HandleWebhook(context.Background(), WebhookInput{
		Provider:  "paystack",
		Payload:   payload,
		Signature: signature,
	})
	if err != nil {
		t.Fatalf("second webhook error = %v", err)
	}
	if processed {
		t.Fatal("expected duplicate webhook to be ignored")
	}
}

func TestDrainOutboxPublishesEvents(t *testing.T) {
	svc := New(&platformconfig.Config{})
	publisher := &stubPublisher{}

	tx, err := svc.InitializePayment(context.Background(), InitializeInput{
		OrderID:        "order-1",
		CustomerID:     "customer-1",
		AmountMinor:    5000,
		Currency:       "NGN",
		Provider:       "stripe",
		IdempotencyKey: "idem-publish",
		CustomerEmail:  "customer@example.com",
		CustomerName:   "Customer One",
	})
	if err != nil {
		t.Fatalf("InitializePayment() error = %v", err)
	}

	count, err := svc.DrainOutbox(context.Background(), publisher, "platform.")
	if err != nil {
		t.Fatalf("DrainOutbox() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 published event, got %d", count)
	}
	if len(publisher.messages) != 1 {
		t.Fatalf("expected exactly one message, got %d", len(publisher.messages))
	}
	if publisher.messages[0].topic != "platform.payment.initialized" {
		t.Fatalf("expected payment topic, got %s", publisher.messages[0].topic)
	}
	if publisher.messages[0].key != tx.ID {
		t.Fatalf("expected transaction key %s, got %s", tx.ID, publisher.messages[0].key)
	}

	var payload map[string]any
	if err := json.Unmarshal(publisher.messages[0].value, &payload); err != nil {
		t.Fatalf("unmarshal published payload: %v", err)
	}
	if payload["event_type"] != "payment.initialized" {
		t.Fatalf("expected event_type payment.initialized, got %v", payload["event_type"])
	}
}

func TestReconcileTransactionsCompletesInitializedPayment(t *testing.T) {
	svc := New(&platformconfig.Config{})
	tx, err := svc.InitializePayment(context.Background(), InitializeInput{
		OrderID:        "order-reconcile",
		CustomerID:     "customer-1",
		AmountMinor:    2000,
		Currency:       "NGN",
		Provider:       "flutterwave",
		IdempotencyKey: "idem-reconcile",
		CustomerEmail:  "customer@example.com",
		CustomerName:   "Customer",
	})
	if err != nil {
		t.Fatalf("InitializePayment() error = %v", err)
	}

	result, err := svc.ReconcileTransactions(context.Background())
	if err != nil {
		t.Fatalf("ReconcileTransactions() error = %v", err)
	}
	if result.Checked == 0 || result.Updated == 0 {
		t.Fatalf("expected reconciliation updates, got %+v", result)
	}

	updated, err := svc.GetPayment(context.Background(), tx.ID)
	if err != nil {
		t.Fatalf("GetPayment() error = %v", err)
	}
	if updated.Status != "completed" {
		t.Fatalf("expected completed status after reconciliation, got %q", updated.Status)
	}
}

func TestCreatePayoutDebitsWalletAndWritesAudit(t *testing.T) {
	svc := New(&platformconfig.Config{})
	if err := svc.ReleaseFunds(context.Background(), "missing", "user-1"); err == nil {
		t.Fatal("expected missing transaction error")
	}

	tx, err := svc.InitializePayment(context.Background(), InitializeInput{
		OrderID:        "order-payout",
		CustomerID:     "user-1",
		AmountMinor:    5000,
		Currency:       "NGN",
		Provider:       "paystack",
		IdempotencyKey: "idem-payout-seed",
		CustomerEmail:  "user@example.com",
		CustomerName:   "User One",
	})
	if err != nil {
		t.Fatalf("InitializePayment() error = %v", err)
	}
	payload := []byte(`{"event_id":"evt-payout","event_type":"payment.success","provider_tx_id":"` + tx.ProviderTxID + `","amount_minor":5000,"currency":"NGN"}`)
	mac := hmac.New(sha512.New, []byte("paystack-secret"))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	processed, err := svc.HandleWebhook(context.Background(), WebhookInput{
		Provider:  "paystack",
		Payload:   payload,
		Signature: signature,
	})
	if err != nil || !processed {
		t.Fatalf("HandleWebhook() error = %v processed=%v", err, processed)
	}
	if err := svc.ReleaseFunds(context.Background(), tx.ID, "user-1"); err != nil {
		t.Fatalf("ReleaseFunds() error = %v", err)
	}

	payout, err := svc.CreatePayout(context.Background(), PayoutInput{
		UserID:         "user-1",
		Provider:       "paystack",
		AmountMinor:    2000,
		Currency:       "NGN",
		DestinationRef: "bank-123",
		IdempotencyKey: "idem-payout-1",
	})
	if err != nil {
		t.Fatalf("CreatePayout() error = %v", err)
	}
	if payout.Status == "" {
		t.Fatal("expected payout status")
	}

	records := svc.ListAuditRecords(context.Background())
	if len(records) == 0 {
		t.Fatal("expected audit records")
	}
}

func TestStartOperationsWorkerRunsReconciliation(t *testing.T) {
	svc := New(&platformconfig.Config{})
	_, err := svc.InitializePayment(context.Background(), InitializeInput{
		OrderID:        "order-worker",
		CustomerID:     "customer-1",
		AmountMinor:    2000,
		Currency:       "NGN",
		Provider:       "stripe",
		IdempotencyKey: "idem-worker",
		CustomerEmail:  "worker@example.com",
		CustomerName:   "Worker",
	})
	if err != nil {
		t.Fatalf("InitializePayment() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_ = svc.StartOperationsWorker(ctx, 10*time.Millisecond)

	records := svc.ListAuditRecords(context.Background())
	if len(records) == 0 {
		t.Fatal("expected reconciliation audit records")
	}
}
