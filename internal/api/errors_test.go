package api

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"rate limit matches sentinel", &RateLimitError{}, ErrRateLimited, true},
		{"rate limit also matches request fail", &RateLimitError{}, ErrRequestFail, true},
		{"auth matches sentinel", &AuthError{}, ErrAuthFailed, true},
		{"auth also matches request fail", &AuthError{}, ErrRequestFail, true},
		{"api error matches request fail", &APIError{}, ErrRequestFail, true},
		{"api error does not match auth", &APIError{}, ErrAuthFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, errors.Is(tt.err, tt.target))
		})
	}
}

func TestErrorsAs(t *testing.T) {
	err := TranslateAPIError(429, "slow down")

	var rateLimit *RateLimitError
	assert.True(t, errors.As(err, &rateLimit))
	assert.Equal(t, 429, rateLimit.StatusCode)

	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr))
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"429 is retryable", 429, true},
		{"500 is retryable", 500, true},
		{"502 is retryable", 502, true},
		{"503 is retryable", 503, true},
		{"529 is retryable", 529, true},
		{"400 is not retryable", 400, false},
		{"401 is not retryable", 401, false},
		{"403 is not retryable", 403, false},
		{"404 is not retryable", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TranslateAPIError(tt.code, "test")
			assert.Equal(t, tt.want, IsRetryable(err))
		})
	}
}

func TestTranslateAPIError(t *testing.T) {
	t.Run("401 produces AuthError", func(t *testing.T) {
		err := TranslateAPIError(401, "bad key")
		var authErr *AuthError
		assert.True(t, errors.As(err, &authErr))
		assert.False(t, IsRetryable(err))
	})

	t.Run("429 produces RateLimitError", func(t *testing.T) {
		err := TranslateAPIError(429, "slow down")
		var rlErr *RateLimitError
		assert.True(t, errors.As(err, &rlErr))
		assert.True(t, IsRetryable(err))
	})

	t.Run("500 produces generic APIError", func(t *testing.T) {
		err := TranslateAPIError(500, "internal error")
		var apiErr *APIError
		assert.True(t, errors.As(err, &apiErr))
		assert.True(t, IsRetryable(err))
	})

	t.Run("400 produces non-retryable APIError", func(t *testing.T) {
		err := TranslateAPIError(400, "bad request")
		var apiErr *APIError
		assert.True(t, errors.As(err, &apiErr))
		assert.False(t, IsRetryable(err))
	})
}

func TestRateLimitError_RetryAfter(t *testing.T) {
	err := &RateLimitError{
		APIError:   APIError{StatusCode: 429, Message: "slow down"},
		RetryAfter: 5 * time.Second,
	}
	assert.Contains(t, err.Error(), "5s")
}
