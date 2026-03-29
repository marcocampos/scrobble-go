package lastfm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ── withRetry ─────────────────────────────────────────────────────────────────

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 3, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), 3, func() error {
		calls++
		if calls < 3 {
			return &NetworkError{NetworkName: "test", UnderlyingError: errors.New("transient")}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ExhaustsAttempts(t *testing.T) {
	calls := 0
	netErr := &NetworkError{NetworkName: "test", UnderlyingError: errors.New("down")}
	err := withRetry(context.Background(), 3, func() error {
		calls++
		return netErr
	})
	if err == nil {
		t.Fatal("expected error after exhausting attempts, got nil")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_NonRetriableErrorStopsImmediately(t *testing.T) {
	calls := 0
	wsErr := &WSError{Status: "6", Details: "invalid params"}
	err := withRetry(context.Background(), 3, func() error {
		calls++
		return wsErr
	})
	if !errors.Is(err, wsErr) {
		t.Fatalf("expected WSError, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call for non-retriable error, got %d", calls)
	}
}

func TestWithRetry_ContextCancelledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	calls := 0
	err := withRetry(ctx, 5, func() error {
		calls++
		return &NetworkError{NetworkName: "test", UnderlyingError: errors.New("transient")}
	})
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
	if calls > 2 {
		t.Errorf("expected context to cancel early, but got %d calls", calls)
	}
}

// ── isRetriable ───────────────────────────────────────────────────────────────

func TestIsRetriable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"NetworkError", &NetworkError{NetworkName: "x", UnderlyingError: errors.New("x")}, true},
		{"WSError 502", &WSError{Status: "502"}, true},
		{"WSError 503", &WSError{Status: "503"}, true},
		{"WSError 504", &WSError{Status: "504"}, true},
		{"WSError 6 (invalid params)", &WSError{Status: "6"}, false},
		{"generic error", errors.New("something"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetriable(tt.err); got != tt.want {
				t.Errorf("isRetriable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ── retryDelay ────────────────────────────────────────────────────────────────

func TestRetryDelay_ExponentialGrowth(t *testing.T) {
	d0 := retryDelay(0)
	d1 := retryDelay(1)
	d2 := retryDelay(2)

	// With jitter each value is within ±25% of the base.
	// Base values: 100ms, 200ms, 400ms.
	check := func(name string, got, base time.Duration) {
		t.Helper()
		lo := time.Duration(float64(base) * 0.75)
		hi := time.Duration(float64(base) * 1.25)
		if got < lo || got > hi {
			t.Errorf("%s: delay %v out of expected range [%v, %v]", name, got, lo, hi)
		}
	}
	check("attempt 0", d0, 100*time.Millisecond)
	check("attempt 1", d1, 200*time.Millisecond)
	check("attempt 2", d2, 400*time.Millisecond)
}

// ── WithRetry option ──────────────────────────────────────────────────────────

func TestWithRetryOption_DefaultThreeAttempts(t *testing.T) {
	c := NewLastFMClient("k", "s", WithRetry())
	if c.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", c.maxRetries)
	}
}

func TestWithRetryOption_CustomAttempts(t *testing.T) {
	c := NewLastFMClient("k", "s", WithRetry(5))
	if c.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", c.maxRetries)
	}
}

func TestWithRetryOption_InvalidValueUsesDefault(t *testing.T) {
	c := NewLastFMClient("k", "s", WithRetry(0))
	if c.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3 for invalid input", c.maxRetries)
	}
}
