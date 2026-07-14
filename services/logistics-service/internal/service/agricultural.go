package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/webhook"
	"github.com/google/uuid"
)

type CreateAgriculturalQuoteInput struct {
	PlatformService, PlatformUserID, FarmSenseRequestID, MarketplaceRequestID, IdempotencyKey string
	Pickup, Dropoff                                                                           model.Coordinate
	Shipment                                                                                  model.AgriculturalShipment
}
type CreateAgriculturalBookingInput struct{ PlatformService, QuoteID, IdempotencyKey string }
type RecordProofInput struct {
	OrderID, DriverUserID, Type, EvidenceURL, Notes, RecipientName string
	Coordinate                                                     model.Coordinate
	CapturedAt                                                     time.Time
}

func (s *Service) CreateAgriculturalQuote(_ context.Context, in CreateAgriculturalQuoteInput) (*model.AgriculturalQuote, error) {
	if in.PlatformService == "" || in.PlatformUserID == "" || in.FarmSenseRequestID == "" || in.IdempotencyKey == "" || !validShipment(in.Shipment) {
		return nil, platformerrors.ErrBadRequest
	}
	key := in.PlatformService + ":quote:" + in.IdempotencyKey
	s.mu.Lock()
	defer s.mu.Unlock()
	if id := s.quoteIdempotency[key]; id != "" {
		copy := *s.agriculturalQuotes[id]
		return &copy, nil
	}
	now := time.Now().UTC()
	quote := &model.AgriculturalQuote{ID: uuid.Must(uuid.NewV7()).String(), PlatformService: in.PlatformService, PlatformUserID: in.PlatformUserID, FarmSenseRequestID: in.FarmSenseRequestID, MarketplaceRequestID: in.MarketplaceRequestID, Pickup: withDerivedH3(in.Pickup), Dropoff: withDerivedH3(in.Dropoff), Shipment: in.Shipment, PriceMinor: agriculturalFare(in.Pickup, in.Dropoff, in.Shipment), Currency: "NGN", CreatedAt: now, ExpiresAt: now.Add(30 * time.Minute)}
	if err := s.store.SaveQuote(context.Background(), quote, in.IdempotencyKey); err != nil {
		return nil, fmt.Errorf("persist agricultural quote: %w", err)
	}
	s.agriculturalQuotes[quote.ID] = quote
	s.quoteIdempotency[key] = quote.ID
	copy := *quote
	return &copy, nil
}

func (s *Service) CreateAgriculturalBooking(ctx context.Context, in CreateAgriculturalBookingInput) (*model.Order, error) {
	if in.PlatformService == "" || in.QuoteID == "" || in.IdempotencyKey == "" {
		return nil, platformerrors.ErrBadRequest
	}
	key := in.PlatformService + ":booking:" + in.IdempotencyKey
	s.agriculturalBookingMu.Lock()
	defer s.agriculturalBookingMu.Unlock()
	s.mu.RLock()
	existing := s.bookingIdempotency[key]
	quote := s.agriculturalQuotes[in.QuoteID]
	s.mu.RUnlock()
	if existing != "" {
		return s.GetOrder(ctx, existing)
	}
	if quote == nil {
		return nil, platformerrors.ErrNotFound
	}
	if quote.PlatformService != in.PlatformService || time.Now().UTC().After(quote.ExpiresAt) {
		return nil, platformerrors.ErrConflict
	}
	customerID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("platform-user:"+quote.PlatformUserID)).String()
	orderKey := uuid.NewSHA1(uuid.NameSpaceURL, []byte(key)).String()
	order, err := s.CreateOrder(ctx, CreateOrderInput{CustomerID: customerID, Type: string(model.OrderTypeAgricultural), Pickup: quote.Pickup, Dropoff: quote.Dropoff, IdempotencyKey: orderKey})
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	stored := s.orders[order.ID]
	stored.PlatformUserID = quote.PlatformUserID
	stored.SourcePlatformService = quote.PlatformService
	stored.MarketplaceRequestID = quote.MarketplaceRequestID
	stored.FarmSenseRequestID = quote.FarmSenseRequestID
	shipment := quote.Shipment
	stored.AgriculturalShipment = &shipment
	stored.PriceMinor = quote.PriceMinor
	s.bookingIdempotency[key] = order.ID
	copy := *stored
	s.mu.Unlock()
	if err := s.store.SaveBooking(ctx, &copy, in.PlatformService, in.IdempotencyKey); err != nil {
		return nil, fmt.Errorf("persist agricultural booking: %w", err)
	}
	s.sendAgriculturalStatus(&copy, "agricultural.booking.created")
	return &copy, nil
}

func (s *Service) RecordDeliveryProof(_ context.Context, in RecordProofInput) (*model.Order, error) {
	if in.OrderID == "" || in.DriverUserID == "" || (in.Type != "pickup" && in.Type != "delivery") || in.EvidenceURL == "" {
		return nil, platformerrors.ErrBadRequest
	}
	s.mu.Lock()
	order := s.orders[in.OrderID]
	driverID := s.driversByUserID[in.DriverUserID]
	if order == nil {
		s.mu.Unlock()
		return nil, platformerrors.ErrNotFound
	}
	if driverID == "" || order.DriverID == nil || *order.DriverID != driverID {
		s.mu.Unlock()
		return nil, platformerrors.ErrUnauthorized
	}
	expected, next := model.OrderStatusAssigned, model.OrderStatusPickedUp
	if in.Type == "delivery" {
		expected, next = model.OrderStatusPickedUp, model.OrderStatusDelivered
		if strings.TrimSpace(in.RecipientName) == "" {
			s.mu.Unlock()
			return nil, platformerrors.ErrBadRequest
		}
	}
	if order.Status != expected {
		s.mu.Unlock()
		return nil, platformerrors.ErrConflict
	}
	if in.CapturedAt.IsZero() {
		in.CapturedAt = time.Now().UTC()
	}
	proof := model.DeliveryProof{ID: uuid.Must(uuid.NewV7()).String(), OrderID: order.ID, Type: in.Type, EvidenceURL: in.EvidenceURL, Notes: in.Notes, RecipientName: in.RecipientName, Coordinate: withDerivedH3(in.Coordinate), CapturedAt: in.CapturedAt, DriverID: driverID}
	previous := order.Status
	order.Status = next
	order.UpdatedAt = time.Now().UTC()
	order.Proofs = append(order.Proofs, proof)
	s.appendOrderHistoryLocked(order.ID, &previous, next, nil, &in.DriverUserID, order.UpdatedAt)
	s.appendOutboxLocked(order.ID, "order."+string(next), map[string]any{"order_id": order.ID, "proof_id": proof.ID})
	s.broadcastLocked(order)
	copy := *order
	s.mu.Unlock()
	if copy.AgriculturalShipment != nil {
		s.sendAgriculturalStatus(&copy, "agricultural.delivery."+string(next))
	}
	return &copy, nil
}

func validShipment(v model.AgriculturalShipment) bool {
	units := map[string]bool{"kg": true, "tonne": true, "crate": true, "bag": true, "basket": true, "litre": true, "unit": true}
	packaging := map[string]bool{"loose": true, "bag": true, "crate": true, "carton": true, "basket": true, "pallet": true, "refrigerated_container": true}
	if strings.TrimSpace(v.ProduceType) == "" || v.Quantity <= 0 || !units[v.QuantityUnit] || !packaging[v.Packaging] || v.PickupWindow.StartAt.IsZero() || !v.PickupWindow.EndAt.After(v.PickupWindow.StartAt) || !v.DeliveryWindow.EndAt.After(v.DeliveryWindow.StartAt) || v.DeliveryWindow.StartAt.Before(v.PickupWindow.StartAt) {
		return false
	}
	if v.RequiresRefrigeration {
		return v.ColdChainMinC != nil && v.ColdChainMaxC != nil && *v.ColdChainMinC < *v.ColdChainMaxC
	}
	return v.ColdChainMinC == nil && v.ColdChainMaxC == nil
}

func agriculturalFare(pickup, dropoff model.Coordinate, shipment model.AgriculturalShipment) int64 {
	return estimateFare(pickup, dropoff, nil) + int64(shipment.Quantity*75)
}

func (s *Service) sendAgriculturalStatus(order *model.Order, _ string) {
	if s.cfg == nil || s.cfg.Security.AgriculturalWebhookURL == "" || s.cfg.Security.AgriculturalWebhookSigningSecret == "" {
		return
	}
	eventID := uuid.Must(uuid.NewV7()).String()
	body, _ := json.Marshal(map[string]any{"event_id": eventID, "event_type": "logistics.delivery.status_changed", "occurred_at": time.Now().UTC(), "source": "logistics", "platform_user_id": order.PlatformUserID, "farmsense_request_id": order.FarmSenseRequestID, "marketplace_request_id": order.MarketplaceRequestID, "logistics_delivery_id": order.ID, "data": map[string]any{"status": normalizedStatus(order.Status), "internal_status": order.Status}})
	if err := s.store.EnqueueWebhook(context.Background(), order.ID, "logistics.delivery.status_changed", body); err != nil {
		return
	}
	if s.cfg.Database.PersistenceMode == "sql" || s.cfg.Server.Environment == "production" {
		return
	}
	url, secret := s.cfg.Security.AgriculturalWebhookURL, s.cfg.Security.AgriculturalWebhookSigningSecret
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = webhook.Deliver(ctx, &http.Client{Timeout: 10 * time.Second}, url, secret, eventID, body)
	}()
}

func normalizedStatus(status model.OrderStatus) string {
	switch status {
	case model.OrderStatusPending:
		return "requested"
	case model.OrderStatusAwaitingPayment:
		return "quoted"
	case model.OrderStatusPaid:
		return "confirmed"
	case model.OrderStatusDispatching:
		return "booked"
	case model.OrderStatusAssigned:
		return "provider_assigned"
	case model.OrderStatusPickedUp:
		return "picked_up"
	case model.OrderStatusDelivered:
		return "delivered"
	case model.OrderStatusCancelled:
		return "cancelled"
	case model.OrderStatusFailed:
		return "failed"
	default:
		return "failed"
	}
}

func NormalizedStatus(status model.OrderStatus) string { return normalizedStatus(status) }

var _ = fmt.Sprintf
