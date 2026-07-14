package service

import (
	"context"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/webhook"
	"net/http"
	"time"
)

func (s *Service) StartWebhookWorker(ctx context.Context) {
	if s.cfg == nil || s.cfg.Security.AgriculturalWebhookURL == "" || s.cfg.Security.AgriculturalWebhookSigningSecret == "" {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.drainWebhooks(ctx)
		}
	}
}
func (s *Service) drainWebhooks(ctx context.Context) {
	items, err := s.store.ClaimWebhooks(ctx, 25)
	if err != nil {
		return
	}
	for _, item := range items {
		deliveryErr := webhook.Deliver(ctx, &http.Client{Timeout: 10 * time.Second}, s.cfg.Security.AgriculturalWebhookURL, s.cfg.Security.AgriculturalWebhookSigningSecret, item.ID, item.Payload)
		_ = s.store.CompleteWebhook(ctx, item.ID, deliveryErr, 10)
	}
}
