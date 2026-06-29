package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/adedaryorh/logistics-platform/services/payment-service/internal/infrastructure/providers"
)

type Transaction struct {
	ID             string         `json:"id"`
	OrderID        string         `json:"order_id"`
	CustomerID     string         `json:"customer_id"`
	AmountMinor    int64          `json:"amount_minor"`
	Currency       string         `json:"currency"`
	Provider       string         `json:"provider"`
	ProviderTxID   string         `json:"provider_tx_id,omitempty"`
	Status         string         `json:"status"`
	CheckoutURL    string         `json:"checkout_url,omitempty"`
	IdempotencyKey string         `json:"idempotency_key"`
	ProviderError  string         `json:"provider_error,omitempty"`
	SettledAt      *time.Time     `json:"settled_at,omitempty"`
	LastVerifiedAt *time.Time     `json:"last_verified_at,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Refund struct {
	ID               string    `json:"id"`
	TransactionID    string    `json:"transaction_id"`
	AmountMinor      int64     `json:"amount_minor"`
	Reason           string    `json:"reason,omitempty"`
	Status           string    `json:"status"`
	ProviderRefundID string    `json:"provider_refund_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type WalletBalance struct {
	UserID       string `json:"user_id"`
	BalanceMinor int64  `json:"balance_minor"`
	Currency     string `json:"currency"`
}

type WalletLedgerEntry struct {
	ID           string    `json:"id"`
	WalletUserID string    `json:"wallet_user_id"`
	Type         string    `json:"type"`
	AmountMinor  int64     `json:"amount_minor"`
	RefID        string    `json:"ref_id,omitempty"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Payout struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Provider         string    `json:"provider"`
	AmountMinor      int64     `json:"amount_minor"`
	Currency         string    `json:"currency"`
	Status           string    `json:"status"`
	DestinationRef   string    `json:"destination_ref"`
	ProviderPayoutID string    `json:"provider_payout_id,omitempty"`
	IdempotencyKey   string    `json:"idempotency_key"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AuditRecord struct {
	ID          string         `json:"id"`
	Action      string         `json:"action"`
	ActorID     string         `json:"actor_id,omitempty"`
	ReferenceID string         `json:"reference_id,omitempty"`
	Provider    string         `json:"provider,omitempty"`
	Status      string         `json:"status"`
	Details     map[string]any `json:"details,omitempty"`
	PrevHash    string         `json:"prev_hash,omitempty"`
	Hash        string         `json:"hash"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ReconciliationResult struct {
	Checked    int      `json:"checked"`
	Updated    int      `json:"updated"`
	Failed     int      `json:"failed"`
	References []string `json:"references"`
}

type InitializeInput struct {
	OrderID        string
	CustomerID     string
	AmountMinor    int64
	Currency       string
	Provider       string
	IdempotencyKey string
	CustomerEmail  string
	CustomerName   string
}

type RefundInput struct {
	TransactionID string
	AmountMinor   int64
	Reason        string
}

type WebhookInput struct {
	Provider  string
	Payload   []byte
	Signature string
}

type PayoutInput struct {
	UserID         string
	Provider       string
	AmountMinor    int64
	Currency       string
	DestinationRef string
	IdempotencyKey string
}

type Service struct {
	cfg              *platformconfig.Config
	providers        map[string]providers.Provider
	mu               sync.RWMutex
	transactions     map[string]*Transaction
	transactionByKey map[string]string
	refunds          map[string][]Refund
	webhookDedup     map[string]struct{}
	wallets          map[string]*WalletBalance
	walletLedger     map[string][]WalletLedgerEntry
	payouts          map[string]*Payout
	audits           []AuditRecord
	lastAuditHash    string
	outbox           []map[string]any
}

func New(cfg *platformconfig.Config) *Service {
	return &Service{
		cfg: cfg,
		providers: map[string]providers.Provider{
			"flutterwave": providers.NewFlutterwaveProvider(),
			"paystack":    providers.NewPaystackProvider(),
			"stripe":      providers.NewStripeProvider(),
		},
		transactions:     map[string]*Transaction{},
		transactionByKey: map[string]string{},
		refunds:          map[string][]Refund{},
		webhookDedup:     map[string]struct{}{},
		wallets:          map[string]*WalletBalance{},
		walletLedger:     map[string][]WalletLedgerEntry{},
		payouts:          map[string]*Payout{},
		audits:           []AuditRecord{},
		lastAuditHash:    "",
		outbox:           []map[string]any{},
	}
}

func (s *Service) InitializePayment(ctx context.Context, input InitializeInput) (*Transaction, error) {
	if input.OrderID == "" || input.Provider == "" || input.IdempotencyKey == "" || input.AmountMinor <= 0 {
		return nil, platformerrors.ErrBadRequest
	}
	provider, ok := s.providers[input.Provider]
	if !ok {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	if existingID, ok := s.transactionByKey[input.IdempotencyKey]; ok {
		txCopy := *s.transactions[existingID]
		s.mu.Unlock()
		return &txCopy, nil
	}
	s.mu.Unlock()

	now := time.Now().UTC()
	tx := &Transaction{
		ID:             "tx-" + input.IdempotencyKey,
		OrderID:        input.OrderID,
		CustomerID:     input.CustomerID,
		AmountMinor:    input.AmountMinor,
		Currency:       defaultCurrency(input.Currency),
		Provider:       input.Provider,
		Status:         "pending",
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	response, err := provider.Initialize(ctx, providers.InitializeRequest{
		TransactionID:  tx.ID,
		AmountMinor:    tx.AmountMinor,
		Currency:       tx.Currency,
		CustomerEmail:  input.CustomerEmail,
		CustomerName:   input.CustomerName,
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		s.mu.Lock()
		s.appendAuditLocked("payment.initialize", tx.ID, input.CustomerID, tx.Provider, "failed", map[string]any{"error": err.Error()})
		s.mu.Unlock()
		observability.IncCounter("payments_initialized_total", 1, map[string]string{"provider": input.Provider, "status": "failed"})
		return nil, mapProviderError(err)
	}

	tx.ProviderTxID = response.ProviderTxID
	tx.CheckoutURL = response.CheckoutURL
	tx.Status = "initialized"
	tx.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	s.transactions[tx.ID] = tx
	s.transactionByKey[input.IdempotencyKey] = tx.ID
	s.outbox = append(s.outbox, map[string]any{
		"aggregate_id": tx.ID,
		"event_type":   "payment.initialized",
		"payload": map[string]any{
			"transaction_id": tx.ID,
			"order_id":       tx.OrderID,
			"amount_minor":   tx.AmountMinor,
			"currency":       tx.Currency,
			"provider":       tx.Provider,
		},
	})
	s.appendAuditLocked("payment.initialize", tx.ID, input.CustomerID, tx.Provider, "completed", map[string]any{"provider_tx_id": tx.ProviderTxID})
	s.mu.Unlock()
	observability.IncCounter("payments_initialized_total", 1, map[string]string{"provider": tx.Provider, "status": "success"})

	txCopy := *tx
	return &txCopy, nil
}

func (s *Service) GetPayment(ctx context.Context, transactionID string) (*Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tx, ok := s.transactions[transactionID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	copyTx := *tx
	return &copyTx, nil
}

func (s *Service) RefundPayment(ctx context.Context, input RefundInput) (*Refund, error) {
	if input.TransactionID == "" {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.RLock()
	tx, ok := s.transactions[input.TransactionID]
	s.mu.RUnlock()
	if !ok {
		return nil, platformerrors.ErrNotFound
	}

	amount := input.AmountMinor
	if amount <= 0 {
		amount = tx.AmountMinor
	}
	provider := s.providers[tx.Provider]
	response, err := provider.Refund(ctx, providers.RefundRequest{
		TransactionID:  tx.ID,
		AmountMinor:    amount,
		Reason:         input.Reason,
		IdempotencyKey: tx.ID + ":refund:" + input.Reason,
	})
	if err != nil {
		s.mu.Lock()
		s.appendAuditLocked("payment.refund", tx.ID, tx.CustomerID, tx.Provider, "failed", map[string]any{"error": err.Error(), "amount_minor": amount})
		s.mu.Unlock()
		observability.IncCounter("payments_refunded_total", 1, map[string]string{"provider": tx.Provider, "status": "failed"})
		return nil, mapProviderError(err)
	}

	refund := Refund{
		ID:               "refund-" + tx.ID + "-" + time.Now().UTC().Format("150405"),
		TransactionID:    tx.ID,
		AmountMinor:      amount,
		Reason:           input.Reason,
		Status:           response.Status,
		ProviderRefundID: response.ProviderRefundID,
		CreatedAt:        time.Now().UTC(),
	}

	s.mu.Lock()
	s.refunds[tx.ID] = append(s.refunds[tx.ID], refund)
	tx.Status = "refunded"
	tx.UpdatedAt = time.Now().UTC()
	s.outbox = append(s.outbox, map[string]any{
		"aggregate_id": tx.ID,
		"event_type":   "payment.refunded",
		"payload": map[string]any{
			"transaction_id": tx.ID,
			"order_id":       tx.OrderID,
			"amount_minor":   amount,
		},
	})
	s.appendAuditLocked("payment.refund", tx.ID, tx.CustomerID, tx.Provider, response.Status, map[string]any{"provider_refund_id": response.ProviderRefundID})
	s.mu.Unlock()
	observability.IncCounter("payments_refunded_total", 1, map[string]string{"provider": tx.Provider, "status": response.Status})

	return &refund, nil
}

func (s *Service) HandleWebhook(ctx context.Context, input WebhookInput) (bool, error) {
	provider, ok := s.providers[input.Provider]
	if !ok {
		return false, platformerrors.ErrBadRequest
	}
	event, err := provider.ParseWebhook(input.Payload, input.Signature)
	if err != nil {
		s.mu.Lock()
		s.appendAuditLocked("payment.webhook", "", "", input.Provider, "failed", map[string]any{"error": err.Error()})
		s.mu.Unlock()
		return false, fmt.Errorf("parse webhook: %w", err)
	}

	dedupKey := input.Provider + ":" + event.EventID
	s.mu.Lock()
	if _, exists := s.webhookDedup[dedupKey]; exists {
		s.mu.Unlock()
		return false, nil
	}
	s.webhookDedup[dedupKey] = struct{}{}
	defer s.mu.Unlock()

	for _, tx := range s.transactions {
		if tx.ProviderTxID != event.ProviderTxID {
			continue
		}
		switch event.EventType {
		case "payment.success", "payment.completed":
			tx.Status = "completed"
			s.outbox = append(s.outbox, map[string]any{"aggregate_id": tx.ID, "event_type": "payment.completed", "payload": map[string]any{"transaction_id": tx.ID, "order_id": tx.OrderID, "amount_minor": tx.AmountMinor, "currency": tx.Currency}})
			observability.IncCounter("payments_completed_total", 1, map[string]string{"provider": tx.Provider, "source": "webhook"})
		case "payment.failed":
			tx.Status = "failed"
			tx.ProviderError = "provider_webhook"
			s.outbox = append(s.outbox, map[string]any{"aggregate_id": tx.ID, "event_type": "payment.failed", "payload": map[string]any{"transaction_id": tx.ID, "order_id": tx.OrderID, "reason": "provider_webhook"}})
			observability.IncCounter("payments_failed_total", 1, map[string]string{"provider": tx.Provider, "source": "webhook"})
		}
		now := time.Now().UTC()
		tx.UpdatedAt = now
		tx.LastVerifiedAt = &now
		s.appendAuditLocked("payment.webhook", tx.ID, tx.CustomerID, tx.Provider, tx.Status, map[string]any{"event_id": event.EventID, "event_type": event.EventType})
		return true, nil
	}

	s.appendAuditLocked("payment.webhook", "", "", input.Provider, "unmatched", map[string]any{"event_id": event.EventID, "provider_tx_id": event.ProviderTxID})
	return true, nil
}

func (s *Service) ReleaseFunds(ctx context.Context, transactionID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, ok := s.transactions[transactionID]
	if !ok {
		return platformerrors.ErrNotFound
	}
	if tx.Status != "completed" && tx.Status != "initialized" {
		return platformerrors.ErrConflict
	}
	wallet := s.ensureWalletLocked(userID, tx.Currency)
	wallet.BalanceMinor += tx.AmountMinor
	entry := WalletLedgerEntry{
		ID:           "ledger-" + transactionID,
		WalletUserID: userID,
		Type:         "credit",
		AmountMinor:  tx.AmountMinor,
		RefID:        transactionID,
		Description:  "payment release",
		CreatedAt:    time.Now().UTC(),
	}
	s.walletLedger[userID] = append(s.walletLedger[userID], entry)
	now := time.Now().UTC()
	tx.SettledAt = &now
	s.outbox = append(s.outbox, map[string]any{"aggregate_id": transactionID, "event_type": "wallet.credited", "payload": map[string]any{"user_id": userID, "amount_minor": tx.AmountMinor, "currency": tx.Currency, "ref_id": transactionID}})
	s.appendAuditLocked("payment.release", transactionID, userID, tx.Provider, "completed", map[string]any{"amount_minor": tx.AmountMinor})
	observability.IncCounter("wallet_releases_total", 1, map[string]string{"provider": tx.Provider, "currency": tx.Currency})
	observability.SetGauge("wallet_balance_minor", float64(wallet.BalanceMinor), map[string]string{"user_id": userID, "currency": wallet.Currency})
	return nil
}

func (s *Service) CancelPayment(ctx context.Context, transactionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, ok := s.transactions[transactionID]
	if !ok {
		return platformerrors.ErrNotFound
	}
	tx.Status = "failed"
	tx.ProviderError = "cancelled"
	tx.UpdatedAt = time.Now().UTC()
	s.appendAuditLocked("payment.cancel", transactionID, tx.CustomerID, tx.Provider, "completed", nil)
	return nil
}

func (s *Service) ReconcileTransactions(ctx context.Context) (*ReconciliationResult, error) {
	result := &ReconciliationResult{References: []string{}}

	s.mu.RLock()
	candidates := make([]*Transaction, 0, len(s.transactions))
	for _, tx := range s.transactions {
		if tx.ProviderTxID == "" {
			continue
		}
		if tx.Status == "completed" || tx.Status == "failed" || tx.Status == "refunded" {
			continue
		}
		candidates = append(candidates, tx)
	}
	s.mu.RUnlock()

	for _, tx := range candidates {
		result.Checked++
		provider := s.providers[tx.Provider]
		verify, err := provider.Verify(ctx, tx.ProviderTxID)
		if err != nil {
			result.Failed++
			s.mu.Lock()
			tx.ProviderError = err.Error()
			s.appendAuditLocked("payment.reconcile", tx.ID, tx.CustomerID, tx.Provider, "failed", map[string]any{"error": err.Error()})
			s.mu.Unlock()
			observability.IncCounter("payment_reconciliation_checks_total", 1, map[string]string{"provider": tx.Provider, "status": "failed"})
			continue
		}

		s.mu.Lock()
		now := time.Now().UTC()
		tx.LastVerifiedAt = &now
		switch verify.Status {
		case "success", "completed":
			tx.Status = "completed"
			result.Updated++
			result.References = append(result.References, tx.ID)
		case "failed":
			tx.Status = "failed"
			result.Updated++
			result.References = append(result.References, tx.ID)
		}
		tx.UpdatedAt = now
		s.appendAuditLocked("payment.reconcile", tx.ID, tx.CustomerID, tx.Provider, tx.Status, map[string]any{"provider_tx_id": tx.ProviderTxID})
		s.mu.Unlock()
		observability.IncCounter("payment_reconciliation_checks_total", 1, map[string]string{"provider": tx.Provider, "status": tx.Status})
	}

	return result, nil
}

func (s *Service) CreatePayout(ctx context.Context, input PayoutInput) (*Payout, error) {
	if input.UserID == "" || input.Provider == "" || input.AmountMinor <= 0 || input.IdempotencyKey == "" || input.DestinationRef == "" {
		return nil, platformerrors.ErrBadRequest
	}
	provider, ok := s.providers[input.Provider]
	if !ok {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	wallet := s.ensureWalletLocked(input.UserID, defaultCurrency(input.Currency))
	if wallet.BalanceMinor < input.AmountMinor {
		s.mu.Unlock()
		return nil, platformerrors.ErrConflict
	}
	for _, payout := range s.payouts {
		if payout.IdempotencyKey == input.IdempotencyKey {
			copyPayout := *payout
			s.mu.Unlock()
			return &copyPayout, nil
		}
	}
	wallet.BalanceMinor -= input.AmountMinor
	s.mu.Unlock()

	payoutID := "payout-" + input.IdempotencyKey
	response, err := provider.Payout(ctx, providers.PayoutRequest{
		Reference:      payoutID,
		UserID:         input.UserID,
		AmountMinor:    input.AmountMinor,
		Currency:       defaultCurrency(input.Currency),
		DestinationRef: input.DestinationRef,
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		s.mu.Lock()
		wallet := s.ensureWalletLocked(input.UserID, defaultCurrency(input.Currency))
		wallet.BalanceMinor += input.AmountMinor
		s.appendAuditLocked("payment.payout", payoutID, input.UserID, input.Provider, "failed", map[string]any{"error": err.Error()})
		s.mu.Unlock()
		observability.IncCounter("payouts_total", 1, map[string]string{"provider": input.Provider, "status": "failed"})
		return nil, mapProviderError(err)
	}

	now := time.Now().UTC()
	payout := &Payout{
		ID:               payoutID,
		UserID:           input.UserID,
		Provider:         input.Provider,
		AmountMinor:      input.AmountMinor,
		Currency:         defaultCurrency(input.Currency),
		Status:           response.Status,
		DestinationRef:   input.DestinationRef,
		ProviderPayoutID: response.ProviderPayoutID,
		IdempotencyKey:   input.IdempotencyKey,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.mu.Lock()
	s.payouts[payout.ID] = payout
	s.walletLedger[input.UserID] = append(s.walletLedger[input.UserID], WalletLedgerEntry{
		ID:           "ledger-" + payout.ID,
		WalletUserID: input.UserID,
		Type:         "debit",
		AmountMinor:  input.AmountMinor,
		RefID:        payout.ID,
		Description:  "payout",
		CreatedAt:    now,
	})
	s.appendAuditLocked("payment.payout", payout.ID, input.UserID, input.Provider, response.Status, map[string]any{"provider_payout_id": response.ProviderPayoutID})
	s.mu.Unlock()
	observability.IncCounter("payouts_total", 1, map[string]string{"provider": input.Provider, "status": response.Status})

	copyPayout := *payout
	return &copyPayout, nil
}

func (s *Service) ListAuditRecords(ctx context.Context) []AuditRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]AuditRecord(nil), s.audits...)
}

func (s *Service) ensureWalletLocked(userID, currency string) *WalletBalance {
	wallet, ok := s.wallets[userID]
	if !ok {
		wallet = &WalletBalance{UserID: userID, Currency: defaultCurrency(currency)}
		s.wallets[userID] = wallet
	}
	return wallet
}

func (s *Service) appendAuditLocked(action, referenceID, actorID, provider, status string, details map[string]any) {
	record := AuditRecord{
		ID:          fmt.Sprintf("audit-%d", len(s.audits)+1),
		Action:      action,
		ActorID:     actorID,
		ReferenceID: referenceID,
		Provider:    provider,
		Status:      status,
		Details:     details,
		PrevHash:    s.lastAuditHash,
		CreatedAt:   time.Now().UTC(),
	}
	record.Hash = hashAuditRecord(record)
	s.lastAuditHash = record.Hash
	s.audits = append(s.audits, record)
}

func defaultCurrency(currency string) string {
	if currency == "" {
		return "NGN"
	}
	return currency
}

func mapProviderError(err error) error {
	if err == nil {
		return nil
	}
	providerErr, ok := err.(*providers.ProviderError)
	if !ok {
		return fmt.Errorf("provider request failed: %w", err)
	}
	switch providerErr.Kind {
	case providers.ProviderErrorBadRequest:
		return platformerrors.ErrBadRequest
	case providers.ProviderErrorUnauthorized:
		return platformerrors.ErrUnauthorized
	case providers.ProviderErrorConflict:
		return platformerrors.ErrConflict
	case providers.ProviderErrorRateLimited:
		return platformerrors.ErrRateLimit
	default:
		return fmt.Errorf("provider request failed: %w", err)
	}
}

func hashAuditRecord(record AuditRecord) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		record.ID,
		record.Action,
		record.ReferenceID,
		record.ActorID,
		record.Provider,
		record.Status,
		record.PrevHash,
	)))
	return hex.EncodeToString(sum[:])
}
