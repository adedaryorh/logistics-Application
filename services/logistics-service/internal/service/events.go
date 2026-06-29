package service

import (
	"context"
	"time"

	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
)

func (s *Service) HandlePaymentCompleted(ctx context.Context, orderID string) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return platformerrors.ErrNotFound
	}
	if order.Status == model.OrderStatusAwaitingPayment || order.Status == model.OrderStatusPending {
		prev := order.Status
		order.Status = model.OrderStatusPaid
		order.UpdatedAt = time.Now().UTC()
		s.appendOrderHistoryLocked(order.ID, &prev, order.Status, nil, nil, order.UpdatedAt)
		s.broadcastLocked(order)
	}
	return nil
}

func (s *Service) HandlePaymentFailed(ctx context.Context, orderID, reason string) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return platformerrors.ErrNotFound
	}
	if order.Status == model.OrderStatusCancelled || order.Status == model.OrderStatusDelivered {
		return nil
	}

	prev := order.Status
	order.Status = model.OrderStatusFailed
	order.UpdatedAt = time.Now().UTC()
	reasonCopy := reason
	if reasonCopy == "" {
		reasonCopy = "payment_failed"
	}
	order.CancellationReason = &reasonCopy
	s.appendOrderHistoryLocked(order.ID, &prev, order.Status, &reasonCopy, nil, order.UpdatedAt)
	s.appendOutboxLocked(order.ID, "order.payment_failed", map[string]any{
		"order_id": order.ID,
		"reason":   reasonCopy,
	})
	s.broadcastLocked(order)
	return nil
}

func (s *Service) HandlePaymentRefunded(ctx context.Context, orderID string) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return platformerrors.ErrNotFound
	}
	if order.Status == model.OrderStatusDelivered || order.Status == model.OrderStatusCancelled {
		return nil
	}

	prev := order.Status
	reason := "payment_refunded"
	order.Status = model.OrderStatusCancelled
	order.CancellationReason = &reason
	order.UpdatedAt = time.Now().UTC()
	s.appendOrderHistoryLocked(order.ID, &prev, order.Status, &reason, nil, order.UpdatedAt)
	s.appendOutboxLocked(order.ID, "order.cancelled", map[string]any{
		"order_id": order.ID,
		"reason":   reason,
	})
	s.broadcastLocked(order)
	return nil
}
