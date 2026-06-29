package providers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Provider interface {
	Name() string
	Initialize(ctx context.Context, req InitializeRequest) (InitializeResponse, error)
	Verify(ctx context.Context, txRef string) (VerifyResponse, error)
	Refund(ctx context.Context, req RefundRequest) (RefundResponse, error)
	Payout(ctx context.Context, req PayoutRequest) (PayoutResponse, error)
	ParseWebhook(payload []byte, signature string) (WebhookEvent, error)
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type ProviderErrorKind string

const (
	ProviderErrorBadRequest   ProviderErrorKind = "bad_request"
	ProviderErrorUnauthorized ProviderErrorKind = "unauthorized"
	ProviderErrorConflict     ProviderErrorKind = "conflict"
	ProviderErrorRateLimited  ProviderErrorKind = "rate_limited"
	ProviderErrorTemporary    ProviderErrorKind = "temporary"
	ProviderErrorUnknown      ProviderErrorKind = "unknown"
)

type ProviderError struct {
	Provider   string
	Kind       ProviderErrorKind
	Message    string
	StatusCode int
	Retryable  bool
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s provider error (%s): %s", e.Provider, e.Kind, e.Message)
}

type InitializeRequest struct {
	TransactionID  string
	AmountMinor    int64
	Currency       string
	CustomerEmail  string
	CustomerName   string
	CallbackURL    string
	Metadata       map[string]string
	IdempotencyKey string
}

type InitializeResponse struct {
	ProviderTxID string
	CheckoutURL  string
	AccessCode   string
	RawResponse  map[string]any
}

type VerifyResponse struct {
	ProviderTxID string
	Status       string
	AmountMinor  int64
	Currency     string
	PaidAt       time.Time
	RawResponse  map[string]any
}

type RefundRequest struct {
	TransactionID  string
	AmountMinor    int64
	Reason         string
	IdempotencyKey string
}

type RefundResponse struct {
	ProviderRefundID string
	Status           string
	RawResponse      map[string]any
}

type PayoutRequest struct {
	Reference      string
	UserID         string
	AmountMinor    int64
	Currency       string
	DestinationRef string
	IdempotencyKey string
}

type PayoutResponse struct {
	ProviderPayoutID string
	Status           string
	RawResponse      map[string]any
}

type WebhookEvent struct {
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	ProviderTxID string          `json:"provider_tx_id"`
	AmountMinor  int64           `json:"amount_minor"`
	Currency     string          `json:"currency"`
	RawPayload   json.RawMessage `json:"raw_payload"`
}

type BaseProvider struct {
	name          string
	secret        string
	apiKey        string
	baseURL       string
	initializeURL string
	verifyURL     string
	refundURL     string
	payoutURL     string
	httpClient    HTTPDoer
	mu            sync.Mutex
	failures      int
	openUntil     time.Time
	initCache     map[string]InitializeResponse
	refundCache   map[string]RefundResponse
	payoutCache   map[string]PayoutResponse
}

func NewFlutterwaveProvider() Provider {
	return newBaseProvider("flutterwave", "FLUTTERWAVE")
}

func NewPaystackProvider() Provider {
	return newBaseProvider("paystack", "PAYSTACK")
}

func NewStripeProvider() Provider {
	return newBaseProvider("stripe", "STRIPE")
}

func newBaseProvider(name, prefix string) Provider {
	baseURL := strings.TrimSpace(os.Getenv(prefix + "_BASE_URL"))
	return &BaseProvider{
		name:          name,
		secret:        envOrDefault(prefix+"_WEBHOOK_SECRET", strings.ToLower(name)+"-secret"),
		apiKey:        strings.TrimSpace(os.Getenv(prefix + "_API_KEY")),
		baseURL:       baseURL,
		initializeURL: strings.TrimSpace(os.Getenv(prefix + "_INITIALIZE_URL")),
		verifyURL:     strings.TrimSpace(os.Getenv(prefix + "_VERIFY_URL")),
		refundURL:     strings.TrimSpace(os.Getenv(prefix + "_REFUND_URL")),
		payoutURL:     strings.TrimSpace(os.Getenv(prefix + "_PAYOUT_URL")),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		initCache:   map[string]InitializeResponse{},
		refundCache: map[string]RefundResponse{},
		payoutCache: map[string]PayoutResponse{},
	}
}

func (p *BaseProvider) Name() string { return p.name }

func (p *BaseProvider) Initialize(ctx context.Context, req InitializeRequest) (InitializeResponse, error) {
	if err := p.beforeCall(); err != nil {
		return InitializeResponse{}, err
	}
	if req.AmountMinor <= 0 || req.TransactionID == "" || req.CustomerEmail == "" {
		p.recordFailure()
		return InitializeResponse{}, &ProviderError{Provider: p.name, Kind: ProviderErrorBadRequest, Message: "invalid initialize request", StatusCode: http.StatusBadRequest}
	}

	if cached, ok := p.loadInitializeCache(req.IdempotencyKey); ok {
		return cached, nil
	}

	if p.canUseLiveAPI() {
		resp, err := p.doJSON(ctx, http.MethodPost, p.resolveURL(p.initializeURL, "/payments"), req, req.IdempotencyKey)
		if err != nil {
			p.recordFailure()
			return InitializeResponse{}, err
		}
		result := InitializeResponse{
			ProviderTxID: stringValue(resp, "provider_tx_id", req.TransactionID+"-"+p.name),
			CheckoutURL:  stringValue(resp, "checkout_url", p.fallbackCheckoutURL(req.TransactionID)),
			AccessCode:   stringValue(resp, "access_code", req.TransactionID+"-access"),
			RawResponse:  resp,
		}
		p.storeInitializeCache(req.IdempotencyKey, result)
		p.recordSuccess()
		return result, nil
	}

	p.recordSuccess()
	result := InitializeResponse{
		ProviderTxID: req.TransactionID + "-" + p.name,
		CheckoutURL:  p.fallbackCheckoutURL(req.TransactionID),
		AccessCode:   req.TransactionID + "-access",
		RawResponse:  map[string]any{"mode": "simulated"},
	}
	p.storeInitializeCache(req.IdempotencyKey, result)
	return result, nil
}

func (p *BaseProvider) Verify(ctx context.Context, txRef string) (VerifyResponse, error) {
	if err := p.beforeCall(); err != nil {
		return VerifyResponse{}, err
	}
	if txRef == "" {
		p.recordFailure()
		return VerifyResponse{}, &ProviderError{Provider: p.name, Kind: ProviderErrorBadRequest, Message: "missing tx ref", StatusCode: http.StatusBadRequest}
	}

	if p.canUseLiveAPI() {
		resp, err := p.doJSON(ctx, http.MethodGet, p.resolveURL(p.verifyURL, "/payments/"+txRef), nil, "")
		if err != nil {
			p.recordFailure()
			return VerifyResponse{}, err
		}
		p.recordSuccess()
		return VerifyResponse{
			ProviderTxID: stringValue(resp, "provider_tx_id", txRef),
			Status:       stringValue(resp, "status", "success"),
			AmountMinor:  int64Value(resp, "amount_minor", 0),
			Currency:     stringValue(resp, "currency", "NGN"),
			PaidAt:       time.Now().UTC(),
			RawResponse:  resp,
		}, nil
	}

	p.recordSuccess()
	return VerifyResponse{
		ProviderTxID: txRef,
		Status:       "success",
		AmountMinor:  0,
		Currency:     "NGN",
		PaidAt:       time.Now().UTC(),
		RawResponse:  map[string]any{"mode": "simulated"},
	}, nil
}

func (p *BaseProvider) Refund(ctx context.Context, req RefundRequest) (RefundResponse, error) {
	if err := p.beforeCall(); err != nil {
		return RefundResponse{}, err
	}
	if req.TransactionID == "" || req.AmountMinor <= 0 {
		p.recordFailure()
		return RefundResponse{}, &ProviderError{Provider: p.name, Kind: ProviderErrorBadRequest, Message: "invalid refund request", StatusCode: http.StatusBadRequest}
	}

	if cached, ok := p.loadRefundCache(req.IdempotencyKey); ok {
		return cached, nil
	}

	if p.canUseLiveAPI() {
		resp, err := p.doJSON(ctx, http.MethodPost, p.resolveURL(p.refundURL, "/refunds"), req, req.IdempotencyKey)
		if err != nil {
			p.recordFailure()
			return RefundResponse{}, err
		}
		result := RefundResponse{
			ProviderRefundID: stringValue(resp, "provider_refund_id", req.TransactionID+"-refund"),
			Status:           stringValue(resp, "status", "completed"),
			RawResponse:      resp,
		}
		p.storeRefundCache(req.IdempotencyKey, result)
		p.recordSuccess()
		return result, nil
	}

	p.recordSuccess()
	result := RefundResponse{
		ProviderRefundID: req.TransactionID + "-refund",
		Status:           "completed",
		RawResponse:      map[string]any{"mode": "simulated"},
	}
	p.storeRefundCache(req.IdempotencyKey, result)
	return result, nil
}

func (p *BaseProvider) Payout(ctx context.Context, req PayoutRequest) (PayoutResponse, error) {
	if err := p.beforeCall(); err != nil {
		return PayoutResponse{}, err
	}
	if req.Reference == "" || req.UserID == "" || req.AmountMinor <= 0 {
		p.recordFailure()
		return PayoutResponse{}, &ProviderError{Provider: p.name, Kind: ProviderErrorBadRequest, Message: "invalid payout request", StatusCode: http.StatusBadRequest}
	}

	if cached, ok := p.loadPayoutCache(req.IdempotencyKey); ok {
		return cached, nil
	}

	if p.canUseLiveAPI() {
		resp, err := p.doJSON(ctx, http.MethodPost, p.resolveURL(p.payoutURL, "/payouts"), req, req.IdempotencyKey)
		if err != nil {
			p.recordFailure()
			return PayoutResponse{}, err
		}
		result := PayoutResponse{
			ProviderPayoutID: stringValue(resp, "provider_payout_id", req.Reference+"-payout"),
			Status:           stringValue(resp, "status", "processing"),
			RawResponse:      resp,
		}
		p.storePayoutCache(req.IdempotencyKey, result)
		p.recordSuccess()
		return result, nil
	}

	p.recordSuccess()
	result := PayoutResponse{
		ProviderPayoutID: req.Reference + "-payout",
		Status:           "processing",
		RawResponse:      map[string]any{"mode": "simulated"},
	}
	p.storePayoutCache(req.IdempotencyKey, result)
	return result, nil
}

func (p *BaseProvider) ParseWebhook(payload []byte, signature string) (WebhookEvent, error) {
	if signature == "" {
		return WebhookEvent{}, fmt.Errorf("%s webhook: missing signature", p.name)
	}
	if err := p.verifySignature(payload, signature); err != nil {
		return WebhookEvent{}, err
	}

	var envelope map[string]any
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return WebhookEvent{}, fmt.Errorf("%s webhook: parse payload: %w", p.name, err)
	}

	event := WebhookEvent{
		EventID:      firstString(envelope, "event_id", "id", "data.id"),
		EventType:    normalizeEventType(p.name, firstString(envelope, "event_type", "event", "type")),
		ProviderTxID: firstString(envelope, "provider_tx_id", "tx_ref", "data.tx_ref", "data.reference", "data.object.id", "data.id"),
		AmountMinor:  firstInt64(envelope, "amount_minor", "amount", "data.amount"),
		Currency:     firstString(envelope, "currency", "data.currency"),
		RawPayload:   json.RawMessage(payload),
	}
	if event.EventID == "" {
		event.EventID = event.ProviderTxID + ":" + event.EventType
	}
	return event, nil
}

func (p *BaseProvider) beforeCall() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if time.Now().UTC().Before(p.openUntil) {
		return &ProviderError{Provider: p.name, Kind: ProviderErrorTemporary, Message: "circuit open", Retryable: true}
	}
	return nil
}

func (p *BaseProvider) recordFailure() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failures++
	if p.failures >= 5 {
		p.openUntil = time.Now().UTC().Add(30 * time.Second)
		p.failures = 0
	}
}

func (p *BaseProvider) recordSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failures = 0
}

func (p *BaseProvider) canUseLiveAPI() bool {
	return p.httpClient != nil && p.baseURL != "" && p.apiKey != ""
}

func (p *BaseProvider) doJSON(ctx context.Context, method, url string, body any, idempotencyKey string) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, &ProviderError{Provider: p.name, Kind: ProviderErrorBadRequest, Message: "marshal request body"}
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, &ProviderError{Provider: p.name, Kind: ProviderErrorUnknown, Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	setIdempotencyHeaders(req.Header, p.name, idempotencyKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, &ProviderError{Provider: p.name, Kind: ProviderErrorTemporary, Message: err.Error(), Retryable: true}
	}
	defer resp.Body.Close()

	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil && err != io.EOF {
		return nil, &ProviderError{Provider: p.name, Kind: ProviderErrorUnknown, Message: "decode response failed"}
	}
	if resp.StatusCode >= 400 {
		return nil, mapProviderHTTPError(p.name, resp.StatusCode, decoded)
	}
	return decoded, nil
}

func mapProviderHTTPError(provider string, status int, payload map[string]any) error {
	message := stringValue(payload, "message", "provider request failed")
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return &ProviderError{Provider: provider, Kind: ProviderErrorUnauthorized, Message: message, StatusCode: status}
	case status == http.StatusConflict:
		return &ProviderError{Provider: provider, Kind: ProviderErrorConflict, Message: message, StatusCode: status}
	case status == http.StatusTooManyRequests:
		return &ProviderError{Provider: provider, Kind: ProviderErrorRateLimited, Message: message, StatusCode: status, Retryable: true}
	case status >= 500:
		return &ProviderError{Provider: provider, Kind: ProviderErrorTemporary, Message: message, StatusCode: status, Retryable: true}
	case status >= 400:
		return &ProviderError{Provider: provider, Kind: ProviderErrorBadRequest, Message: message, StatusCode: status}
	default:
		return &ProviderError{Provider: provider, Kind: ProviderErrorUnknown, Message: message, StatusCode: status}
	}
}

func setIdempotencyHeaders(header http.Header, provider, key string) {
	if key == "" {
		return
	}
	switch provider {
	case "stripe":
		header.Set("Idempotency-Key", key)
	case "paystack":
		header.Set("X-Idempotency-Key", key)
	default:
		header.Set("Idempotency-Key", key)
	}
}

func (p *BaseProvider) resolveURL(customPath, fallbackPath string) string {
	if customPath != "" {
		if strings.HasPrefix(customPath, "http://") || strings.HasPrefix(customPath, "https://") {
			return customPath
		}
		return strings.TrimRight(p.baseURL, "/") + "/" + strings.TrimLeft(customPath, "/")
	}
	return strings.TrimRight(p.baseURL, "/") + fallbackPath
}

func (p *BaseProvider) fallbackCheckoutURL(transactionID string) string {
	return "https://payments.example/" + p.name + "/" + transactionID
}

func (p *BaseProvider) verifySignature(payload []byte, signature string) error {
	switch p.name {
	case "flutterwave":
		if !hmac.Equal([]byte(signature), []byte(p.secret)) {
			return fmt.Errorf("%s webhook: invalid signature", p.name)
		}
		return nil
	case "paystack":
		mac := hmac.New(sha512.New, []byte(p.secret))
		_, _ = mac.Write(payload)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			return fmt.Errorf("%s webhook: invalid signature", p.name)
		}
		return nil
	case "stripe":
		mac := hmac.New(sha256.New, []byte(p.secret))
		_, _ = mac.Write(payload)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			return fmt.Errorf("%s webhook: invalid signature", p.name)
		}
		return nil
	default:
		return fmt.Errorf("%s webhook: unsupported provider", p.name)
	}
}

func (p *BaseProvider) loadInitializeCache(key string) (InitializeResponse, bool) {
	if key == "" {
		return InitializeResponse{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	resp, ok := p.initCache[key]
	return resp, ok
}

func (p *BaseProvider) storeInitializeCache(key string, resp InitializeResponse) {
	if key == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.initCache[key] = resp
}

func (p *BaseProvider) loadRefundCache(key string) (RefundResponse, bool) {
	if key == "" {
		return RefundResponse{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	resp, ok := p.refundCache[key]
	return resp, ok
}

func (p *BaseProvider) storeRefundCache(key string, resp RefundResponse) {
	if key == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refundCache[key] = resp
}

func (p *BaseProvider) loadPayoutCache(key string) (PayoutResponse, bool) {
	if key == "" {
		return PayoutResponse{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	resp, ok := p.payoutCache[key]
	return resp, ok
}

func (p *BaseProvider) storePayoutCache(key string, resp PayoutResponse) {
	if key == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.payoutCache[key] = resp
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func stringValue(m map[string]any, key, fallback string) string {
	if value, ok := m[key]; ok {
		if cast, ok := value.(string); ok && cast != "" {
			return cast
		}
	}
	return fallback
}

func int64Value(m map[string]any, key string, fallback int64) int64 {
	if value, ok := m[key]; ok {
		switch cast := value.(type) {
		case float64:
			return int64(cast)
		case int64:
			return cast
		case int:
			return int64(cast)
		}
	}
	return fallback
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := nestedValue(payload, key); value != nil {
			if cast, ok := value.(string); ok && cast != "" {
				return cast
			}
		}
	}
	return ""
}

func firstInt64(payload map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if value := nestedValue(payload, key); value != nil {
			switch cast := value.(type) {
			case float64:
				return int64(cast)
			case int64:
				return cast
			case int:
				return int64(cast)
			}
		}
	}
	return 0
}

func nestedValue(payload map[string]any, dotted string) any {
	parts := strings.Split(dotted, ".")
	var current any = payload
	for _, part := range parts {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = obj[part]
	}
	return current
}

func normalizeEventType(provider, raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch provider {
	case "flutterwave":
		if value == "charge.completed" || value == "successful" {
			return "payment.completed"
		}
	case "paystack":
		if value == "charge.success" {
			return "payment.completed"
		}
	case "stripe":
		if value == "payment_intent.succeeded" || value == "charge.succeeded" {
			return "payment.completed"
		}
	}
	switch value {
	case "payment.success":
		return "payment.completed"
	case "payment.failed", "charge.failed", "payment_intent.payment_failed":
		return "payment.failed"
	}
	return value
}
