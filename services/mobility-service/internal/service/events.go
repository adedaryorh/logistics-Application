package service

import "context"

func (s *Service) HandleOrderCreated(ctx context.Context, orderID string, pickupLat, pickupLng, dropoffLat, dropoffLng *float64) error {
	return s.prepareTrackingSession(ctx, orderID, pickupLat, pickupLng, dropoffLat, dropoffLng)
}

func (s *Service) HandleOrderAssigned(ctx context.Context, orderID string) error {
	return s.activateTrackingSession(orderID)
}

func (s *Service) HandleOrderCancelled(ctx context.Context, orderID string) error {
	return s.EndTrackingSession(ctx, orderID)
}

func (s *Service) HandlePaymentFailed(ctx context.Context, orderID string) error {
	return s.EndTrackingSession(ctx, orderID)
}
