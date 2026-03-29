package lastfm

import (
	"context"
	"crypto/md5" //nolint:gosec // required by the Last.fm API
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// apiRequest represents a single Last.fm web service call.
type apiRequest struct {
	client *Client
	method string
	params map[string]string
}

// newAPIRequest creates a request, attaches auth params, and signs it if a
// session key is present.
func newAPIRequest(c *Client, method string, params map[string]string) *apiRequest {
	r := &apiRequest{
		client: c,
		method: method,
		params: make(map[string]string, len(params)+4),
	}
	for k, v := range params {
		r.params[k] = convertParam(v)
	}
	r.params["api_key"] = c.net.APIKey
	r.params["method"] = method

	if c.net.SessionKey != "" {
		r.params["sk"] = c.net.SessionKey
		r.sign()
	}
	return r
}

// convertParam normalises a parameter value to a string.
// Boolean true/false become "1"/"0" per the Last.fm API convention.
func convertParam(v any) string {
	switch val := v.(type) {
	case bool:
		if val {
			return "1"
		}
		return "0"
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// sign appends an "api_sig" parameter if one is not already present.
func (r *apiRequest) sign() {
	if _, ok := r.params["api_sig"]; !ok {
		r.params["api_sig"] = r.signature()
	}
}

// signature builds the MD5 hex signature as required by the Last.fm API:
// sort all param names, concatenate name+value pairs, append the API secret,
// then MD5 the result.
func (r *apiRequest) signature() string {
	keys := make([]string, 0, len(r.params))
	for k := range r.params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(r.params[k])
	}
	b.WriteString(r.client.net.APISecret)

	h := md5.Sum([]byte(b.String())) //nolint:gosec
	return fmt.Sprintf("%x", h)
}

// cacheKey returns a stable SHA-1 hex key derived from the request params,
// excluding auth-only fields (api_sig, api_key, sk).
func (r *apiRequest) cacheKey() string {
	keys := make([]string, 0, len(r.params))
	for k := range r.params {
		if k != "api_sig" && k != "api_key" && k != "sk" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(r.params[k])
	}

	// Use MD5 for the cache key (not security-sensitive).
	h := md5.Sum([]byte(b.String())) //nolint:gosec
	return fmt.Sprintf("%x", h)
}

// execute performs the request and returns the parsed XML response.
// If cacheable is true and a cache backend is configured, the response is
// read from (and written to) the cache.
func (r *apiRequest) execute(ctx context.Context, cacheable bool) (*xmlNode, error) {
	slog.Debug("last.fm API call", "method", r.method)

	if cacheable && r.client.cache != nil {
		key := r.cacheKey()
		if cached, ok := r.client.cache.Get(key); ok {
			slog.Debug("cache hit", "method", r.method)
			return parseXMLResponse(cached)
		}
	}

	body, err := r.download(ctx)
	if err != nil {
		return nil, err
	}

	if cacheable && r.client.cache != nil {
		r.client.cache.Set(r.cacheKey(), body)
	}

	return parseXMLResponse(body)
}

// download makes the actual HTTP POST request and returns the raw body.
func (r *apiRequest) download(ctx context.Context) (string, error) {
	if r.client.rateLimit {
		r.client.delayCall()
	}

	formData := make(url.Values, len(r.params))
	for k, v := range r.params {
		formData.Set(k, v)
	}

	endpoint := fmt.Sprintf("https://%s%s", r.client.net.WSHost, r.client.net.WSPath)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", &NetworkError{
			NetworkName:     r.client.net.Name,
			UnderlyingError: err,
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("User-Agent", "scrobble.go/"+version)

	resp, err := r.client.httpClient.Do(req)
	if err != nil {
		return "", &NetworkError{
			NetworkName:     r.client.net.Name,
			UnderlyingError: err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusInternalServerError ||
		resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return "", &WSError{
			Status:      fmt.Sprintf("%d", resp.StatusCode),
			Details:     fmt.Sprintf("API connection failed with HTTP %d", resp.StatusCode),
			networkName: r.client.net.Name,
		}
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &NetworkError{
			NetworkName:     r.client.net.Name,
			UnderlyingError: err,
		}
	}
	body := string(raw)

	if err := checkAPIErrors(body, r.client.net.Name); err != nil {
		return "", err
	}
	return body, nil
}

// checkAPIErrors parses the XML response and returns a WSError if the API
// reports a failure, or a MalformedResponseError if the XML is invalid.
func checkAPIErrors(body, networkName string) error {
	doc, err := parseXMLResponse(body)
	if err != nil {
		return &MalformedResponseError{
			NetworkName:     networkName,
			UnderlyingError: err,
		}
	}

	slog.Debug("API response", "xml", body)

	if doc.attr("status") == "ok" {
		return nil
	}

	errEl := doc.find("error")
	if errEl == nil {
		return &MalformedResponseError{
			NetworkName:     networkName,
			UnderlyingError: fmt.Errorf("status=%q but no <error> element found", doc.attr("status")),
		}
	}
	return &WSError{
		Status:      errEl.attr("code"),
		Details:     strings.TrimSpace(errEl.Content),
		networkName: networkName,
	}
}
