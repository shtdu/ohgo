package api

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for programmatic matching via errors.Is.
var (
	ErrAuthFailed  = errors.New("authentication failed")
	ErrRateLimited = errors.New("rate limited")
	ErrRequestFail = errors.New("request failed")
)

// APIError is the base error for upstream API failures.
type APIError struct {
	StatusCode int
	Message    string
	Retryable  bool
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error: status %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) Is(target error) bool {
	return target == ErrRequestFail
}

// As allows errors.As to match embedded APIError types.
func (e *APIError) As(target any) bool {
	if t, ok := target.(**APIError); ok {
		*t = e
		return true
	}
	return false
}

// RateLimitError indicates a 429 response.
type RateLimitError struct {
	APIError
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited: status %d: %s (retry after %s)", e.StatusCode, e.Message, e.RetryAfter)
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited || e.APIError.Is(target)
}

// AuthError indicates a 401/403 response.
type AuthError struct {
	APIError
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: status %d: %s", e.StatusCode, e.Message)
}

func (e *AuthError) Is(target error) bool {
	return target == ErrAuthFailed || e.APIError.Is(target)
}

// IsRetryable returns true if the error represents a transient failure.
func IsRetryable(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.StatusCode {
	case 429, 500, 502, 503, 529:
		return true
	default:
		return false
	}
}

// TranslateAPIError maps an HTTP status code and body to a typed error.
func TranslateAPIError(statusCode int, body string) error {
	switch statusCode {
	case 401, 403:
		return &AuthError{
			APIError: APIError{
				StatusCode: statusCode,
				Message:    body,
				Retryable:  false,
			},
		}
	case 429:
		return &RateLimitError{
			APIError: APIError{
				StatusCode: statusCode,
				Message:    body,
				Retryable:  true,
			},
		}
	default:
		retryable := statusCode >= 500
		return &APIError{
			StatusCode: statusCode,
			Message:    body,
			Retryable:  retryable,
		}
	}
}
