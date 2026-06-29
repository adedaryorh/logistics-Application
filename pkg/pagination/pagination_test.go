package pagination

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEncodeDecodeCursor(t *testing.T) {
	t.Parallel()

	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("NewV7 returned error: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Nanosecond)
	cursor := EncodeCursor(now, id)

	gotTime, gotID, err := DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("DecodeCursor returned error: %v", err)
	}

	if !gotTime.Equal(now) {
		t.Fatalf("expected time %v, got %v", now, gotTime)
	}
	if gotID != id {
		t.Fatalf("expected id %v, got %v", id, gotID)
	}
}

func TestDecodeCursorInvalidLength(t *testing.T) {
	t.Parallel()

	if _, _, err := DecodeCursor("abcd"); err == nil {
		t.Fatal("expected invalid cursor length error")
	}
}
