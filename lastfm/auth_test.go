package lastfm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const mobileSessionXML = `<lfm status="ok">
  <session>
    <name>testuser</name>
    <key>abc123sessionkey</key>
    <subscriber>0</subscriber>
  </session>
</lfm>`

const webAuthTokenXML = `<lfm status="ok">
  <token>webauthtoken42</token>
</lfm>`

const webAuthSessionXML = `<lfm status="ok">
  <session>
    <name>webuser</name>
    <key>websessionkey99</key>
    <subscriber>0</subscriber>
  </session>
</lfm>`

const tokenUnauthorizedXML = `<lfm status="failed">
  <error code="14">This token has not been authorised</error>
</lfm>`

func TestGetSessionKey_Success(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(mobileSessionXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	gen := NewSessionKeyGenerator(c)

	sk, err := gen.GetSessionKey(context.Background(), "testuser", MD5("password"))
	if err != nil {
		t.Fatalf("GetSessionKey: unexpected error: %v", err)
	}
	if sk != "abc123sessionkey" {
		t.Errorf("session key = %q, want %q", sk, "abc123sessionkey")
	}
}

func TestGetSessionKey_WSError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(sampleErrorXML)) // invalid API key
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	gen := NewSessionKeyGenerator(c)

	_, err := gen.GetSessionKey(context.Background(), "user", "hash")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var wsErr *WSError
	if !errors.As(err, &wsErr) {
		t.Errorf("expected *WSError in chain, got %T: %v", err, err)
	}
}

func TestGetWebAuthURL(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(webAuthTokenXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	gen := NewSessionKeyGenerator(c)

	url, err := gen.GetWebAuthURL(context.Background())
	if err != nil {
		t.Fatalf("GetWebAuthURL: unexpected error: %v", err)
	}
	if url == "" {
		t.Error("GetWebAuthURL returned empty URL")
	}
	// The token should be recorded internally.
	if _, ok := gen.webAuthTokens[url]; !ok {
		t.Error("token not stored for the returned URL")
	}
}

func TestGetWebAuthSessionKeyAndUsername_Success(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(webAuthSessionXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	gen := NewSessionKeyGenerator(c)

	sk, username, err := gen.GetWebAuthSessionKeyAndUsername(
		context.Background(), "", "sometoken",
	)
	if err != nil {
		t.Fatalf("GetWebAuthSessionKeyAndUsername: unexpected error: %v", err)
	}
	if sk != "websessionkey99" {
		t.Errorf("session key = %q, want %q", sk, "websessionkey99")
	}
	if username != "webuser" {
		t.Errorf("username = %q, want %q", username, "webuser")
	}
}

func TestGetWebAuthSessionKey_TokenUnauthorized(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(tokenUnauthorizedXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	gen := NewSessionKeyGenerator(c)

	_, err := gen.GetWebAuthSessionKey(context.Background(), "")
	// Should fail with "no token for URL" before even hitting the server.
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestAuthenticateWithPassword_StoresSessionKey(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(mobileSessionXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.AuthenticateWithPassword(context.Background(), "testuser", MD5("password"))
	if err != nil {
		t.Fatalf("AuthenticateWithPassword: unexpected error: %v", err)
	}
	if c.net.SessionKey != "abc123sessionkey" {
		t.Errorf("session key = %q, want %q", c.net.SessionKey, "abc123sessionkey")
	}
	if c.net.Username != "testuser" {
		t.Errorf("username = %q, want %q", c.net.Username, "testuser")
	}
}
