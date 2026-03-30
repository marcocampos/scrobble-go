package lastfm

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

const (
	retryBaseDelay = 100 * time.Millisecond
	retryMaxDelay  = 30 * time.Second
	retryJitter    = 0.25 // ±25% jitter applied to each delay
)

// withRetry calls fn up to maxAttempts times, retrying on transient errors
// (NetworkError and HTTP 502/503/504). It uses exponential backoff starting at
// retryBaseDelay with ±retryJitter jitter. Non-retriable errors (WSError,
// context cancellation) are returned immediately without further attempts.
func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
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

		timer := time.NewTimer(retryDelay(attempt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
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
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return true
	}
	var wsErr *WSError
	if errors.As(err, &wsErr) {
		switch wsErr.Status {
		case "502", "503", "504":
			return true
		}
	}
	return false
}

// retryDelay returns the backoff duration for a given attempt index (0-based),
// applying exponential growth and ±retryJitter random jitter.
// The base delay is capped at retryMaxDelay to prevent overflow for large attempt values.
func retryDelay(attempt int) time.Duration {
	base := retryBaseDelay * (1 << uint(attempt)) // 100ms, 200ms, 400ms, …
	if base <= 0 || base > retryMaxDelay {
		base = retryMaxDelay
	}
	jitter := float64(base) * retryJitter * (2*rand.Float64() - 1)
	return base + time.Duration(jitter)
}
