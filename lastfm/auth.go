package lastfm

import (
	"context"
	"fmt"
	"sync"
)

// SessionKeyGenerator handles the two flows for obtaining a Last.fm session key:
//
//  1. Mobile auth  — username + MD5(password), no browser required.
//  2. Web auth     — redirect the user to a Last.fm URL; poll until they approve.
//
// A session key is valid indefinitely unless the user revokes it.
// Once obtained, pass it to NewLastFMClient via WithSessionKey.
type SessionKeyGenerator struct {
	client        *Client
	mu            sync.RWMutex
	webAuthTokens map[string]string // url → token
}

// NewSessionKeyGenerator returns a SessionKeyGenerator for the given client.
func NewSessionKeyGenerator(c *Client) *SessionKeyGenerator {
	return &SessionKeyGenerator{
		client:        c,
		webAuthTokens: make(map[string]string),
	}
}

// GetSessionKey authenticates with a username and MD5 password hash and
// returns a session key.
//
// Use MD5("yourpassword") to hash the password before passing it here.
func (g *SessionKeyGenerator) GetSessionKey(ctx context.Context, username, passwordHash string) (string, error) {
	params := map[string]string{
		"username":  username,
		"authToken": MD5(username + passwordHash),
	}

	r := newAPIRequest(g.client, "auth.getMobileSession", params)
	r.sign() // sign even without a session key

	doc, err := r.execute(ctx, false)
	if err != nil {
		return "", fmt.Errorf("GetSessionKey: %w", err)
	}

	key := extract(doc, "key")
	if key == "" {
		return "", &MalformedResponseError{
			NetworkName:     g.client.net.Name,
			UnderlyingError: fmt.Errorf("no <key> element in auth.getMobileSession response"),
		}
	}
	return key, nil
}

// GetWebAuthURL obtains a one-time token from Last.fm and returns the URL
// to which the user should be redirected to authorise the application.
// After authorisation, call GetWebAuthSessionKey with the same URL.
func (g *SessionKeyGenerator) GetWebAuthURL(ctx context.Context) (string, error) {
	token, err := g.getWebAuthToken(ctx)
	if err != nil {
		return "", err
	}

	authURL := fmt.Sprintf(
		"%s/api/auth/?api_key=%s&token=%s",
		g.client.net.Homepage,
		g.client.net.APIKey,
		token,
	)
	g.mu.Lock()
	g.webAuthTokens[authURL] = token
	g.mu.Unlock()
	return authURL, nil
}

// GetWebAuthSessionKey retrieves the session key after the user has authorised
// the application at the URL returned by GetWebAuthURL.
// Poll this method (with a short sleep between attempts) until it succeeds or
// returns a non-token-unauthorised error.
func (g *SessionKeyGenerator) GetWebAuthSessionKey(ctx context.Context, authURL string) (string, error) {
	sk, _, err := g.GetWebAuthSessionKeyAndUsername(ctx, authURL, "")
	return sk, err
}

// GetWebAuthSessionKeyAndUsername is like GetWebAuthSessionKey but also returns
// the authenticated username.
func (g *SessionKeyGenerator) GetWebAuthSessionKeyAndUsername(ctx context.Context, authURL, token string) (sessionKey, username string, err error) {
	g.mu.RLock()
	t, ok := g.webAuthTokens[authURL]
	g.mu.RUnlock()
	if ok {
		token = t
	}
	if token == "" {
		return "", "", fmt.Errorf("GetWebAuthSessionKeyAndUsername: no token for URL %q", authURL)
	}

	r := newAPIRequest(g.client, "auth.getSession", map[string]string{"token": token})
	r.sign()

	doc, err := r.execute(ctx, false)
	if err != nil {
		return "", "", fmt.Errorf("GetWebAuthSessionKeyAndUsername: %w", err)
	}

	sessionKey = extract(doc, "key")
	username = extract(doc, "name")
	if sessionKey == "" {
		return "", "", &MalformedResponseError{
			NetworkName:     g.client.net.Name,
			UnderlyingError: fmt.Errorf("no <key> element in auth.getSession response"),
		}
	}
	return sessionKey, username, nil
}

// AuthenticateWithPassword is a convenience wrapper that runs mobile auth and
// stores the resulting session key on the client.
func (c *Client) AuthenticateWithPassword(ctx context.Context, username, passwordHash string) error {
	gen := NewSessionKeyGenerator(c)
	sk, err := gen.GetSessionKey(ctx, username, passwordHash)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.net.SessionKey = sk
	c.net.Username = username
	c.mu.Unlock()
	return nil
}

// getWebAuthToken fetches a short-lived token from auth.getToken.
func (g *SessionKeyGenerator) getWebAuthToken(ctx context.Context) (string, error) {
	r := newAPIRequest(g.client, "auth.getToken", nil)
	r.sign()

	doc, err := r.execute(ctx, false)
	if err != nil {
		return "", fmt.Errorf("getWebAuthToken: %w", err)
	}

	token := extract(doc, "token")
	if token == "" {
		return "", &MalformedResponseError{
			NetworkName:     g.client.net.Name,
			UnderlyingError: fmt.Errorf("no <token> element in auth.getToken response"),
		}
	}
	return token, nil
}
