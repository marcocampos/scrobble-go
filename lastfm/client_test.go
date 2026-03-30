package lastfm

import (
	"net/http"
	"testing"
	"time"
)

func TestWithUsername(t *testing.T) {
	c := NewLastFMClient("k", "s", WithUsername("marco"))
	if c.net.Username != "marco" {
		t.Errorf("Username = %q, want %q", c.net.Username, "marco")
	}
}

func TestWithPasswordHash(t *testing.T) {
	c := NewLastFMClient("k", "s", WithPasswordHash("abc123"))
	if c.net.PasswordHash != "abc123" {
		t.Errorf("PasswordHash = %q, want %q", c.net.PasswordHash, "abc123")
	}
}

func TestWithCache(t *testing.T) {
	cache := NewMemoryCache()
	c := NewLastFMClient("k", "s", WithCache(cache))
	if c.cache != cache {
		t.Error("cache should be the exact instance passed to WithCache")
	}
}

func TestWithRateLimit(t *testing.T) {
	c := NewLastFMClient("k", "s", WithRateLimit())
	if !c.rateLimit {
		t.Error("rateLimit should be true")
	}
}

func TestWithHTTPClient(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	c := NewLastFMClient("k", "s", WithHTTPClient(hc))
	if c.httpClient != hc {
		t.Error("httpClient should be the custom client")
	}
}

func TestDelayCall_EnforcesGap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
	c := NewLastFMClient("k", "s", WithRateLimit())
	c.delayCall() // prime lastCall
	start := time.Now()
	c.delayCall() // should block for ~200ms
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Errorf("delayCall gap = %v, want ≥ 150ms", elapsed)
	}
}

func TestDelayCall_NoRateLimit(t *testing.T) {
	c := NewLastFMClient("k", "s") // no WithRateLimit
	// delayCall always enforces the gap when called directly; callers
	// in the request path guard it with c.rateLimit. Call once to
	// verify it does not panic on a zero-value lastCall.
	c.delayCall()
}
