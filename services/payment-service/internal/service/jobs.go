package service

import (
	"context"
	"errors"
	"time"
)

func (s *Service) StartOperationsWorker(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if _, err := s.ReconcileTransactions(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
