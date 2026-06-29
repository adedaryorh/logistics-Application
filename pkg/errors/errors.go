package errors

import (
	"fmt"
	"net/http"
)

var (
	// ErrNotFound represents a resource not found error
	ErrNotFound = NewError("not_found", "resource not found", http.StatusNotFound)

	// ErrUnauthorized represents an unauthorized error
	ErrUnauthorized = NewError("unauthorized", "unauthorized", http.StatusUnauthorized)

	// ErrForbidden represents a forbidden error
	ErrForbidden = NewError("forbidden", "forbidden", http.StatusForbidden)

	// ErrConflict represents a conflict error
	ErrConflict = NewError("conflict", "resource conflict", http.StatusConflict)

	// ErrBadRequest represents a bad request error
	ErrBadRequest = NewError("bad_request", "bad request", http.StatusBadRequest)

	// ErrInternal represents an internal server error
	ErrInternal = NewError("internal_error", "internal server error", http.StatusInternalServerError)

	// ErrRateLimit represents a rate limit exceeded error
	ErrRateLimit = NewError("rate_limit_exceeded", "rate limit exceeded", http.StatusTooManyRequests)
)

// ErrorCode represents an error code
type ErrorCode string

// APIError represents an API error with code, message, and HTTP status
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
	// HTTPStatus is the HTTP status code to return
	HTTPStatus int `json:"-"`
}

// NewError creates a new APIError
func NewError(code ErrorCode, message string, status int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    nil,
		HTTPStatus: status,
	}
}

func WithDetails(apiErr *APIError, details any) *APIError {
	if apiErr == nil {
		return nil
	}
	return &APIError{
		Code:       apiErr.Code,
		Message:    apiErr.Message,
		Details:    details,
		HTTPStatus: apiErr.HTTPStatus,
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	return string(e.Code) + ": " + e.Message
}

// Is returns true if the target error matches this error code
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		return e.Code == t.Code
	}
	return false
}

// Wrap wraps an error with context while preserving the original error
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return &APIError{
		Code:       ErrorCode("wrapped_error"),
		Message:    context + ": " + err.Error(),
		HTTPStatus: http.StatusInternalServerError,
	}
}

// Wrapf wraps an error with a formatted context message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return Wrap(err, fmt.Sprintf(format, args...))
}
