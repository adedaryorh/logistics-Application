package service

import (
	"context"
	"testing"
)

func TestScrubString(t *testing.T) {
	got := scrubString("email jane@example.com phone +2348012345678")
	if got == "email jane@example.com phone +2348012345678" {
		t.Fatal("expected pii to be scrubbed")
	}
}

func TestAuthenticateRejectsInvalidKey(t *testing.T) {
	svc := New(nil)
	if _, err := svc.Authenticate("bad-key"); err == nil {
		t.Fatal("expected invalid key to fail")
	}
}

func TestCallToolWritesAudit(t *testing.T) {
	svc := New(nil)
	record, err := svc.Authenticate("ops-key")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if _, err := svc.CallTool(context.Background(), record, "health_check", map[string]any{"email": "jane@example.com"}); err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if len(svc.AuditLog()) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(svc.AuditLog()))
	}
}

func TestRBACBlocksUnauthorizedTool(t *testing.T) {
	svc := New(nil)
	record, err := svc.Authenticate("support-key")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if _, err := svc.CallTool(context.Background(), record, "payment_status", map[string]any{"transaction_id": "tx-1"}); err == nil {
		t.Fatal("expected unauthorized tool call to fail")
	}
}

func TestSensitiveToolRequiresReason(t *testing.T) {
	svc := New(nil)
	record, err := svc.Authenticate("finance-key")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if _, err := svc.CallTool(context.Background(), record, "payment_status", map[string]any{"transaction_id": "tx-1"}); err == nil {
		t.Fatal("expected missing reason to fail")
	}
}

func TestRequiredFieldValidation(t *testing.T) {
	svc := New(nil)
	record, err := svc.Authenticate("ops-key")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if _, err := svc.CallTool(context.Background(), record, "get_metrics", map[string]any{}); err == nil {
		t.Fatal("expected missing required field to fail")
	}
}
