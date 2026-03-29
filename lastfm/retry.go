package lastfm

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"
)

const (
	retryBaseDelay = 100 * time.Millisecond
	retryJitter    = 0.25 // ±25% jitter applied to each delay
)

// withRetry calls fn up to maxAttempts times, retrying on transient errors
// (NetworkError and HTTP 502/503/504). It uses exponential backoff starting at
// retryBaseDelay with ±retryJitter jitter. Non-retriable errors (WSError,
// context cancellation) are returned immediately without further attempts.
func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
	var err error
	for attempt := range maxAttempts {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetriable(err) {
			return err
		}

		if attempt == maxAttempts-1 {
			break
		}

		delay := retryDelay(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

// isRetriable reports whether err should trigger a retry.
// NetworkErrors and server-side HTTP errors (502/503/504) are retriable;
// API-level WSErrors and malformed responses are not.
func isRetriable(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*NetworkError); ok {
		return true
	}
	if wsErr, ok := err.(*WSError); ok {
		switch wsErr.Status {
		case
			http.StatusText(http.StatusBadGateway),
			http.StatusText(http.StatusServiceUnavailable),
			http.StatusText(http.StatusGatewayTimeout),
			"502", "503", "504":
			return true
		}
	}
	return false
}

// retryDelay returns the backoff duration for a given attempt index (0-based),
// applying exponential growth and ±retryJitter random jitter.
func retryDelay(attempt int) time.Duration {
	base := retryBaseDelay * (1 << uint(attempt)) // 100ms, 200ms, 400ms, …
	jitter := float64(base) * retryJitter * (2*rand.Float64() - 1)
	return base + time.Duration(jitter)
}
