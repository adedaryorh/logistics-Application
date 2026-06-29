package service

import (
	"context"
	"testing"
)

func TestPaymentFailureCreatesIncidentAndTrace(t *testing.T) {
	svc := New(nil)

	if err := svc.HandleOrderCreated(context.Background(), "order-1", "customer-1"); err != nil {
		t.Fatalf("HandleOrderCreated() error = %v", err)
	}
	if err := svc.HandlePaymentFailed(context.Background(), "order-1", "provider_timeout"); err != nil {
		t.Fatalf("HandlePaymentFailed() error = %v", err)
	}

	trace, err := svc.GetOrderTrace(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrderTrace() error = %v", err)
	}
	if trace.PaymentStatus != "failed" {
		t.Fatalf("expected failed payment status, got %q", trace.PaymentStatus)
	}
	if trace.CurrentIncident == "" {
		t.Fatal("expected current incident to be set")
	}

	incidents := svc.ListIncidents(context.Background())
	if len(incidents) == 0 {
		t.Fatal("expected incident to be recorded")
	}
}

func TestIncidentAckAndResolve(t *testing.T) {
	svc := New(nil)
	if err := svc.HandlePaymentFailed(context.Background(), "order-2", "declined"); err != nil {
		t.Fatalf("HandlePaymentFailed() error = %v", err)
	}
	incident := svc.ListIncidents(context.Background())[0]

	acked, err := svc.AcknowledgeIncident(context.Background(), incident.ID)
	if err != nil {
		t.Fatalf("AcknowledgeIncident() error = %v", err)
	}
	if acked.Status != "acknowledged" {
		t.Fatalf("expected acknowledged status, got %q", acked.Status)
	}

	resolved, err := svc.ResolveIncident(context.Background(), incident.ID)
	if err != nil {
		t.Fatalf("ResolveIncident() error = %v", err)
	}
	if resolved.Status != "resolved" {
		t.Fatalf("expected resolved status, got %q", resolved.Status)
	}
}
