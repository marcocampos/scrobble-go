package lastfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient returns a Client wired to the given TLS test server.
// The server's own Client() is used so the self-signed cert is trusted.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c := NewLastFMClient("testapikey", "testapisecret",
		WithHTTPClient(srv.Client()),
	)
	// Point the client at the test server (TLS).
	c.net.WSHost = srv.Listener.Addr().String()
	c.net.WSPath = "/"
	return c
}

// serveXML returns an TLS httptest.Server that always responds with body.
func serveXML(body string) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
}

func TestAPIRequest_Signature(t *testing.T) {
	c := NewLastFMClient("abc", "secret")
	r := newAPIRequest(c, "artist.getInfo", map[string]string{
		"artist": "Iron Maiden",
	})
	// Signing must not panic and must produce a 32-char hex string.
	sig := r.signature()
	if len(sig) != 32 {
		t.Errorf("signature length = %d, want 32", len(sig))
	}
}

func TestAPIRequest_SignatureIsDeterministic(t *testing.T) {
	c := NewLastFMClient("abc", "secret")
	r := newAPIRequest(c, "artist.getInfo", map[string]string{
		"artist": "Iron Maiden",
	})
	sig1 := r.signature()
	sig2 := r.signature()
	if sig1 != sig2 {
		t.Errorf("signature is not deterministic: %q != %q", sig1, sig2)
	}
}

func TestAPIRequest_CacheKey_ExcludesAuthFields(t *testing.T) {
	c1 := NewLastFMClient("key1", "secret1", WithSessionKey("sk1"))
	c2 := NewLastFMClient("key2", "secret2", WithSessionKey("sk2"))

	r1 := newAPIRequest(c1, "artist.getInfo", map[string]string{"artist": "Maiden"})
	r2 := newAPIRequest(c2, "artist.getInfo", map[string]string{"artist": "Maiden"})

	// Cache keys should be identical regardless of api_key, api_secret, sk.
	if r1.cacheKey() != r2.cacheKey() {
		t.Errorf("cache keys differ for identical method+params: %q vs %q",
			r1.cacheKey(), r2.cacheKey())
	}
}

func TestAPIRequest_Execute_Success(t *testing.T) {
	srv := serveXML(sampleArtistXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	r := newAPIRequest(c, "artist.getInfo", map[string]string{"artist": "Iron Maiden"})

	doc, err := r.execute(context.Background(), false)
	if err != nil {
		t.Fatalf("execute: unexpected error: %v", err)
	}
	if name := extract(doc, "name"); name != "Iron Maiden" {
		t.Errorf("extract(name) = %q, want %q", name, "Iron Maiden")
	}
}

func TestAPIRequest_Execute_WSError(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	r := newAPIRequest(c, "artist.getInfo", map[string]string{"artist": "Iron Maiden"})

	_, err := r.execute(context.Background(), false)
	if err == nil {
		t.Fatal("expected WSError, got nil")
	}
	if _, ok := err.(*WSError); !ok {
		t.Errorf("expected *WSError, got %T: %v", err, err)
	}
}

func TestAPIRequest_Execute_CacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(sampleArtistXML))
	}))
	defer srv.Close()

	cache := NewMemoryCache()
	c := newTestClient(t, srv)
	c.cache = cache

	req := newAPIRequest(c, "artist.getInfo", map[string]string{"artist": "Iron Maiden"})

	// First call — should hit the server.
	if _, err := req.execute(context.Background(), true); err != nil {
		t.Fatalf("first execute: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 server call, got %d", callCount)
	}

	// Second call — should be served from cache.
	req2 := newAPIRequest(c, "artist.getInfo", map[string]string{"artist": "Iron Maiden"})
	if _, err := req2.execute(context.Background(), true); err != nil {
		t.Fatalf("second execute: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 server call after cache hit, got %d", callCount)
	}
}

func TestAPIRequest_Execute_NetworkError(t *testing.T) {
	// Use a server that immediately closes connections.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	r := newAPIRequest(c, "artist.getInfo", nil)

	_, err := r.execute(context.Background(), false)
	if err == nil {
		t.Fatal("expected NetworkError, got nil")
	}
	if _, ok := err.(*NetworkError); !ok {
		t.Errorf("expected *NetworkError, got %T: %v", err, err)
	}
}

func TestConvertParam(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{true, "1"},
		{false, "0"},
		{"hello", "hello"},
		{42, "42"},
		{3.14, "3.14"},
	}
	for _, tt := range tests {
		got := convertParam(tt.input)
		if got != tt.want {
			t.Errorf("convertParam(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
