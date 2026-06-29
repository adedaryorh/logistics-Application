package logger

import (
	"context"
	"testing"
)

func TestWithContextFields(t *testing.T) {
	t.Parallel()

	log, err := New("svc", "test", "v1")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, traceIDKey, "trace-1")
	ctx = context.WithValue(ctx, requestIDKey, "request-1")
	ctx = context.WithValue(ctx, userIDKey, "user-1")

	if got := log.WithTraceID(ctx); got == nil {
		t.Fatal("expected logger from WithTraceID")
	}
	if got := log.WithRequestID(ctx); got == nil {
		t.Fatal("expected logger from WithRequestID")
	}
	if got := log.WithUserID(ctx); got == nil {
		t.Fatal("expected logger from WithUserID")
	}
}
