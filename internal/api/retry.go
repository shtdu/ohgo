package api

import (
	"math/rand/v2"
	"time"
)

const (
	baseRetryDelay = 1 * time.Second
	maxRetryDelay  = 30 * time.Second
	maxRetries     = 3
)

// retryDelay computes an exponential backoff duration with jitter.
func retryDelay(attempt int) time.Duration {
	if attempt > 30 {
		return maxRetryDelay
	}
	delay := min(baseRetryDelay*time.Duration(1<<uint(attempt)), maxRetryDelay)
	jitter := time.Duration(rand.Int64N(int64(delay) / 4))
	return delay + jitter
}
