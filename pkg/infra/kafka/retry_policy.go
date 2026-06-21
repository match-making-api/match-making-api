package kafka

import (
	"os"
	"strconv"
	"time"
)

// RetryPolicy controls consumer retries before routing to DLQ.
type RetryPolicy struct {
	MaxRetries int
	Backoff    time.Duration
}

// DefaultRetryPolicy reads KAFKA_MAX_RETRIES (default 3) and KAFKA_RETRY_BACKOFF_MS (default 500).
func DefaultRetryPolicy() RetryPolicy {
	maxRetries := 3
	if v := os.Getenv("KAFKA_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxRetries = n
		}
	}

	backoff := 500 * time.Millisecond
	if v := os.Getenv("KAFKA_RETRY_BACKOFF_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			backoff = time.Duration(n) * time.Millisecond
		}
	}

	return RetryPolicy{MaxRetries: maxRetries, Backoff: backoff}
}

// ShouldRetry returns true if another attempt is allowed.
func (p RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < p.MaxRetries
}
