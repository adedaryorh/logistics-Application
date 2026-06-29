package pagination

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Page represents a paginated slice of items
type Page[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
}

// Cursor represents a cursor for pagination
type Cursor struct {
	Time  time.Time
	ID    uuid.UUID
	Other []byte // for additional flexibility
}

// EncodeCursor encodes a cursor to a string
func EncodeCursor(t time.Time, id uuid.UUID) string {
	t = t.UTC()

	// Encode time as Unix nanoseconds
	var timeBuf [8]byte
	binary.BigEndian.PutUint64(timeBuf[:], uint64(t.UnixNano()))

	// Convert UUID to bytes
	idBytes := id[:]

	// Combine time and ID
	var buf []byte
	buf = append(buf, timeBuf[:]...)
	buf = append(buf, idBytes[:]...)

	// Base64 encode
	return base64.RawURLEncoding.EncodeToString(buf)
}

// DecodeCursor decodes a cursor string to time and UUID
func DecodeCursor(s string) (time.Time, uuid.UUID, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	if len(data) != 24 {
		return time.Time{}, uuid.Nil, errors.New("invalid cursor length")
	}

	// First 8 bytes are time nanoseconds
	timeBytes := data[:8]
	idBytes := data[8:]

	// Decode time
	nano := binary.BigEndian.Uint64(timeBytes)
	t := time.Unix(0, int64(nano)).UTC()

	// Decode UUID
	var id uuid.UUID
	copy(id[:], idBytes)

	return t, id, nil
}

// paginate is a helper function for slice pagination
func paginate[T any](items []T, pageSize int, after *Cursor) ([]T, *Cursor, bool) {
	if len(items) == 0 {
		return []T{}, nil, false
	}

	start := 0
	if after != nil {
		// Find the index after the cursor
		for i := range items {
			// This is a simplified implementation - in reality,
			// you'd compare based on the actual fields used in the cursor
			if i == 0 { // Skip first item if it matches cursor (simplified)
				start = i + 1
				break
			}
		}
	}

	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	result := items[start:end]

	var nextCursor *Cursor
	hasMore := false
	if end < len(items) {
		hasMore = true
		nextCursor = &Cursor{
			Time: time.Now().UTC(),
			ID:   uuid.New(),
		}
	}

	return result, nextCursor, hasMore
}
