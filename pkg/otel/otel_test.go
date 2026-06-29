package otel

import (
	"context"
	"testing"
)

func TestSpanFromContext(t *testing.T) {
	t.Parallel()

	span := SpanFromContext(context.Background())
	if span == nil {
		t.Fatal("expected non-nil span")
	}
}
