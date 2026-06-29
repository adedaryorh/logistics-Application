package errors

import (
	stderrors "errors"
	"net/http"
	"testing"
)

func TestNewError(t *testing.T) {
	t.Parallel()

	err := NewError("custom", "failed", http.StatusBadRequest)
	if err.Code != "custom" {
		t.Fatalf("expected code custom, got %s", err.Code)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
}

func TestWrapf(t *testing.T) {
	t.Parallel()

	err := Wrapf(stderrors.New("boom"), "while doing %s", "work")
	if err == nil {
		t.Fatal("expected wrapped error")
	}
	if got := err.Error(); got != "wrapped_error: while doing work: boom" {
		t.Fatalf("unexpected wrapped error %q", got)
	}
}

func TestIs(t *testing.T) {
	t.Parallel()

	if !stderrors.Is(ErrNotFound, NewError("not_found", "missing", http.StatusNotFound)) {
		t.Fatal("expected errors.Is to match by code")
	}
}
