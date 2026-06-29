package service

import (
	"context"
	"time"

	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
)

func (s *Service) HandleOrderCancelled(ctx context.Context, orderID string) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, tx := range s.transactions {
		if tx.OrderID != orderID {
			continue
		}
		if tx.Status == "failed" || tx.Status == "refunded" {
			return nil
		}
		tx.Status = "failed"
		tx.UpdatedAt = time.Now().UTC()
		s.outbox = append(s.outbox, map[string]any{
			"aggregate_id": tx.ID,
			"event_type":   "payment.cancelled",
			"payload": map[string]any{
				"transaction_id": tx.ID,
				"order_id":       tx.OrderID,
				"reason":         "order_cancelled",
			},
		})
		return nil
	}

	return nil
}
