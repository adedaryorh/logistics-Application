package service

import "context"

func (s *Service) HandleUserRegistered(ctx context.Context, userID, email string) error {
	s.mu.Lock()
	s.recordWorkflowLocked("user_onboarding", userID, "completed", map[string]any{"email": email})
	s.mu.Unlock()
	if email != "" {
		_, err := s.SendTemplateEmail(ctx, userID, email, "welcome", map[string]any{
			"Name":             email,
			"VerificationLink": "https://example.com/verify",
		})
		if err != nil {
			return err
		}
		return nil
	}

	s.RecordNotification(userID, "unknown", "welcome", "queued")
	return nil
}

func (s *Service) HandleOrderCreated(ctx context.Context, orderID, customerID string) error {
	s.mu.Lock()
	state := s.appendOrderEventLocked(orderID, "order.created", map[string]any{"customer_id": customerID})
	state.CustomerID = customerID
	state.DispatchStatus = "created"
	s.recordWorkflowLocked("order_lifecycle", orderID, "dispatching", map[string]any{"customer_id": customerID})
	s.mu.Unlock()
	s.RecordBackgroundJob("order_created_followup", map[string]any{
		"order_id":    orderID,
		"customer_id": customerID,
	})
	s.RecordNotification(customerID, "customer", "order_confirmed", "queued")
	return nil
}

func (s *Service) HandleOrderAssigned(ctx context.Context, orderID, driverID string) error {
	s.mu.Lock()
	state := s.appendOrderEventLocked(orderID, "order.assigned", map[string]any{"driver_id": driverID})
	state.DriverID = driverID
	state.DispatchStatus = "assigned"
	state.TrackingStatus = "active"
	s.recordWorkflowLocked("dispatch_assignment", orderID, "completed", map[string]any{"driver_id": driverID})
	s.mu.Unlock()
	s.RecordNotification(driverID, "driver", "order_assigned", "queued")
	s.RecordBackgroundJob("dispatch_assignment_observed", map[string]any{
		"order_id":  orderID,
		"driver_id": driverID,
	})
	return nil
}

func (s *Service) HandlePaymentFailed(ctx context.Context, orderID, reason string) error {
	s.mu.Lock()
	state := s.appendOrderEventLocked(orderID, "payment.failed", map[string]any{"reason": reason})
	state.PaymentStatus = "failed"
	incident := s.createIncidentLocked("payment", "high", orderID, "Payment failed for active order", "payment_failure_followup")
	state.CurrentIncident = incident.ID
	s.recordWorkflowLocked("payment_recovery", orderID, "attention_required", map[string]any{"reason": reason, "incident_id": incident.ID})
	s.mu.Unlock()
	s.RecordNotification(orderID, "customer", "payment_failed", "queued")
	s.RecordBackgroundJob("payment_failure_followup", map[string]any{
		"order_id": orderID,
		"reason":   reason,
	})
	return nil
}

func (s *Service) HandlePaymentCompleted(ctx context.Context, orderID string) error {
	s.mu.Lock()
	state := s.appendOrderEventLocked(orderID, "payment.completed", nil)
	state.PaymentStatus = "completed"
	s.recordWorkflowLocked("payment_confirmation", orderID, "completed", nil)
	s.mu.Unlock()
	s.RecordNotification(orderID, "customer", "payment_received", "queued")
	return nil
}

func (s *Service) HandleDispatchOffered(ctx context.Context, orderID, driverID string, attempt int) error {
	s.mu.Lock()
	state := s.appendOrderEventLocked(orderID, "dispatch.offered", map[string]any{"driver_id": driverID, "attempt": attempt})
	state.DispatchStatus = "offered"
	s.mu.Unlock()
	s.RecordBackgroundJob("dispatch_offer_followup", map[string]any{
		"order_id":  orderID,
		"driver_id": driverID,
		"attempt":   attempt,
	})
	return nil
}

func (s *Service) HandleOrderCancelled(ctx context.Context, orderID, reason string) error {
	s.mu.Lock()
	state := s.appendOrderEventLocked(orderID, "order.cancelled", map[string]any{"reason": reason})
	state.DispatchStatus = "cancelled"
	state.TrackingStatus = "ended"
	s.recordWorkflowLocked("order_cancellation", orderID, "completed", map[string]any{"reason": reason})
	s.mu.Unlock()
	s.RecordNotification(orderID, "customer", "order_cancelled", "queued")
	return nil
}
