package lastfm

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const (
	// delayTime is the minimum gap between API calls mandated by Last.fm ToS §4.4.
	delayTime = 200 * time.Millisecond

	version = "0.2.0"
)

// network holds the static configuration for a Last.fm-compatible service.
type network struct {
	Name         string
	Homepage     string
	WSHost       string // e.g. "ws.audioscrobbler.com"
	WSPath       string // e.g. "/2.0/"
	APIKey       string
	APISecret    string
	SessionKey   string
	Username     string
	PasswordHash string
	DomainNames  map[int]string
	URLs         map[string]string
}

// Client is the entry point for all Last.fm API interactions.
// Create one with NewLastFMClient or NewLibreFMClient.
type Client struct {
	net         *network
	httpClient  *http.Client
	cache       CacheBackend
	rateLimit   bool
	maxAttempts int

	mu       sync.RWMutex
	lastCall time.Time
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithSessionKey sets a pre-existing session key, skipping automatic auth.
func WithSessionKey(sk string) Option {
	return func(c *Client) { c.net.SessionKey = sk }
}

// WithUsername sets the username used for mobile (password) authentication.
func WithUsername(username string) Option {
	return func(c *Client) { c.net.Username = username }
}

// WithPasswordHash sets the MD5 password hash used for mobile authentication.
// Use MD5("yourpassword") to produce the value.
func WithPasswordHash(hash string) Option {
	return func(c *Client) { c.net.PasswordHash = hash }
}

// WithCache enables response caching with the provided backend.
func WithCache(backend CacheBackend) Option {
	return func(c *Client) { c.cache = backend }
}

// WithRateLimit enables automatic rate limiting (≥200 ms between calls),
// as required by the Last.fm API Terms of Service §4.4.
func WithRateLimit() Option {
	return func(c *Client) { c.rateLimit = true }
}

// WithHTTPClient replaces the default HTTP client (useful for proxy or
// transport customisation, and for testing).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithRetry enables automatic retries on transient errors (network failures
// and HTTP 502/503/504). The default is 3 attempts with exponential backoff
// starting at 100 ms and ±25% jitter. Pass a value ≥ 1 to override.
func WithRetry(maxAttempts ...int) Option {
	attempts := 3
	if len(maxAttempts) > 0 && maxAttempts[0] >= 1 {
		attempts = maxAttempts[0]
	}
	return func(c *Client) { c.maxAttempts = attempts }
}

// newClient creates a Client for the given network, applying all options.
func newClient(net *network, opts []Option) *Client {
	c := &Client{
		net:        net,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// NewLastFMClient returns a Client configured for Last.fm.
func NewLastFMClient(apiKey, apiSecret string, opts ...Option) *Client {
	net := &network{
		Name:      "Last.fm",
		Homepage:  "https://www.last.fm",
		WSHost:    "ws.audioscrobbler.com",
		WSPath:    "/2.0/",
		APIKey:    apiKey,
		APISecret: apiSecret,
		DomainNames: map[int]string{
			DomainEnglish:    "www.last.fm",
			DomainGerman:     "www.last.fm/de",
			DomainSpanish:    "www.last.fm/es",
			DomainFrench:     "www.last.fm/fr",
			DomainItalian:    "www.last.fm/it",
			DomainPolish:     "www.last.fm/pl",
			DomainPortuguese: "www.last.fm/pt",
			DomainSwedish:    "www.last.fm/sv",
			DomainTurkish:    "www.last.fm/tr",
			DomainRussian:    "www.last.fm/ru",
			DomainJapanese:   "www.last.fm/ja",
			DomainChinese:    "www.last.fm/zh",
		},
		URLs: map[string]string{
			"album":   "music/%s/%s",
			"artist":  "music/%s",
			"country": "place/%s",
			"tag":     "tag/%s",
			"track":   "music/%s/_/%s",
			"user":    "user/%s",
		},
	}
	return newClient(net, opts)
}

// NewLibreFMClient returns a Client configured for Libre.fm.
func NewLibreFMClient(apiKey, apiSecret string, opts ...Option) *Client {
	net := &network{
		Name:      "Libre.fm",
		Homepage:  "https://libre.fm",
		WSHost:    "libre.fm",
		WSPath:    "/2.0/",
		APIKey:    apiKey,
		APISecret: apiSecret,
		DomainNames: map[int]string{
			DomainEnglish:    "libre.fm",
			DomainGerman:     "libre.fm",
			DomainSpanish:    "libre.fm",
			DomainFrench:     "libre.fm",
			DomainItalian:    "libre.fm",
			DomainPolish:     "libre.fm",
			DomainPortuguese: "libre.fm",
			DomainSwedish:    "libre.fm",
			DomainTurkish:    "libre.fm",
			DomainRussian:    "libre.fm",
			DomainJapanese:   "libre.fm",
			DomainChinese:    "libre.fm",
		},
		URLs: map[string]string{
			"album":   "artist/%s/album/%s",
			"artist":  "artist/%s",
			"country": "place/%s",
			"tag":     "tag/%s",
			"track":   "music/%s/_/%s",
			"user":    "user/%s",
		},
	}
	return newClient(net, opts)
}

// delayCall sleeps until at least delayTime has elapsed since the last call,
// honouring the provided context. Returns ctx.Err() if the context is
// cancelled before the delay expires.
//
// It uses a reservation scheme: under the lock it advances lastCall to the
// next allowed slot, then waits outside the lock. This ensures concurrent
// callers are spaced correctly.
func (c *Client) delayCall(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	now := time.Now()
	next := c.lastCall.Add(delayTime)

	var wait time.Duration
	if now.Before(next) {
		wait = next.Sub(now)
		c.lastCall = next
	} else {
		c.lastCall = now
	}
	c.mu.Unlock()

	if wait <= 0 {
		return nil
	}

	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
