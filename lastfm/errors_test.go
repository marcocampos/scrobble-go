package lastfm

import (
	"errors"
	"strings"
	"testing"
)

func TestWSError_Error(t *testing.T) {
	e := &WSError{Status: StatusInvalidAPIKey, Details: "Invalid API key"}
	msg := e.Error()
	if !strings.Contains(msg, "10") {
		t.Errorf("Error() = %q, want it to contain status code", msg)
	}
	if !strings.Contains(msg, "Invalid API key") {
		t.Errorf("Error() = %q, want it to contain details", msg)
	}
}

func TestNetworkError_Error(t *testing.T) {
	underlying := errors.New("connection refused")
	e := &NetworkError{NetworkName: "Last.fm", UnderlyingError: underlying}
	msg := e.Error()
	if !strings.Contains(msg, "Last.fm") {
		t.Errorf("Error() = %q, want it to contain network name", msg)
	}
	if !strings.Contains(msg, "connection refused") {
		t.Errorf("Error() = %q, want it to contain underlying error", msg)
	}
}

func TestNetworkError_Unwrap(t *testing.T) {
	underlying := errors.New("timeout")
	e := &NetworkError{NetworkName: "Last.fm", UnderlyingError: underlying}
	if !errors.Is(e, underlying) {
		t.Error("errors.Is should find the underlying error via Unwrap")
	}
}

func TestMalformedResponseError_Error(t *testing.T) {
	underlying := errors.New("unexpected EOF")
	e := &MalformedResponseError{NetworkName: "Last.fm", UnderlyingError: underlying}
	msg := e.Error()
	if !strings.Contains(msg, "Last.fm") {
		t.Errorf("Error() = %q, want it to contain network name", msg)
	}
	if !strings.Contains(msg, "unexpected EOF") {
		t.Errorf("Error() = %q, want it to contain underlying error", msg)
	}
}

func TestMalformedResponseError_Unwrap(t *testing.T) {
	underlying := errors.New("bad xml")
	e := &MalformedResponseError{NetworkName: "Last.fm", UnderlyingError: underlying}
	if !errors.Is(e, underlying) {
		t.Error("errors.Is should find the underlying error via Unwrap")
	}
}
