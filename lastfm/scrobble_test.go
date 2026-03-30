package lastfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const scrobbleOKXML = `<lfm status="ok">
  <scrobbles accepted="1" ignored="0">
    <scrobble>
      <track corrected="0">The Nomad</track>
      <artist corrected="0">Iron Maiden</artist>
      <album corrected="0">Dance of Death</album>
      <albumArtist corrected="0">Iron Maiden</albumArtist>
      <timestamp>1609459200</timestamp>
      <ignoredMessage code="0"></ignoredMessage>
    </scrobble>
  </scrobbles>
</lfm>`

const nowPlayingOKXML = `<lfm status="ok">
  <nowplaying>
    <track corrected="0">The Nomad</track>
    <artist corrected="0">Iron Maiden</artist>
    <album corrected="0">Dance of Death</album>
    <albumArtist corrected="0">Iron Maiden</albumArtist>
    <ignoredMessage code="0"></ignoredMessage>
  </nowplaying>
</lfm>`

// captureServer returns a TLS server that records the last POST body and
// responds with the given XML.
func captureServer(responseXML string) (*httptest.Server, *url.Values) {
	captured := &url.Values{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err == nil {
			*captured = r.Form
		}
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(responseXML))
	}))
	return srv, captured
}

func TestScrobble_SingleTrack(t *testing.T) {
	srv, captured := captureServer(scrobbleOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.Scrobble(context.Background(), ScrobbleParams{
		Artist:    "Iron Maiden",
		Title:     "The Nomad",
		Timestamp: 1609459200,
		Album:     "Dance of Death",
	})
	if err != nil {
		t.Fatalf("Scrobble: unexpected error: %v", err)
	}

	if captured.Get("method") != "track.scrobble" {
		t.Errorf("method = %q, want %q", captured.Get("method"), "track.scrobble")
	}
	if captured.Get("artist[0]") != "Iron Maiden" {
		t.Errorf("artist[0] = %q, want %q", captured.Get("artist[0]"), "Iron Maiden")
	}
	if captured.Get("track[0]") != "The Nomad" {
		t.Errorf("track[0] = %q, want %q", captured.Get("track[0]"), "The Nomad")
	}
	if captured.Get("timestamp[0]") != "1609459200" {
		t.Errorf("timestamp[0] = %q, want %q", captured.Get("timestamp[0]"), "1609459200")
	}
	if captured.Get("album[0]") != "Dance of Death" {
		t.Errorf("album[0] = %q, want %q", captured.Get("album[0]"), "Dance of Death")
	}
}

func TestScrobbleMany_BatchSplit(t *testing.T) {
	callCount := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(scrobbleOKXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)

	// Build 55 tracks — should result in two API calls (50 + 5).
	tracks := make([]ScrobbleParams, 55)
	for i := range tracks {
		tracks[i] = ScrobbleParams{
			Artist:    "Artist",
			Title:     "Track",
			Timestamp: int64(1609459200 + i),
		}
	}

	if err := c.ScrobbleMany(context.Background(), tracks); err != nil {
		t.Fatalf("ScrobbleMany: unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for 55 tracks, got %d", callCount)
	}
}

func TestScrobble_ChosenByUser(t *testing.T) {
	srv, captured := captureServer(scrobbleOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	chosen := false
	err := c.Scrobble(context.Background(), ScrobbleParams{
		Artist:       "Artist",
		Title:        "Track",
		Timestamp:    1609459200,
		ChosenByUser: &chosen,
	})
	if err != nil {
		t.Fatalf("Scrobble: %v", err)
	}
	if captured.Get("chosenByUser[0]") != "0" {
		t.Errorf("chosenByUser[0] = %q, want %q", captured.Get("chosenByUser[0]"), "0")
	}
}

func TestUpdateNowPlaying(t *testing.T) {
	srv, captured := captureServer(nowPlayingOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.UpdateNowPlaying(context.Background(), NowPlayingParams{
		Artist:   "Iron Maiden",
		Title:    "The Nomad",
		Album:    "Dance of Death",
		Duration: 613,
	})
	if err != nil {
		t.Fatalf("UpdateNowPlaying: unexpected error: %v", err)
	}

	if captured.Get("method") != "track.updateNowPlaying" {
		t.Errorf("method = %q, want %q", captured.Get("method"), "track.updateNowPlaying")
	}
	if captured.Get("artist") != "Iron Maiden" {
		t.Errorf("artist = %q, want %q", captured.Get("artist"), "Iron Maiden")
	}
	if captured.Get("duration") != "613" {
		t.Errorf("duration = %q, want %q", captured.Get("duration"), "613")
	}
}

func TestUpdateNowPlaying_OptionalFieldsOmitted(t *testing.T) {
	srv, captured := captureServer(nowPlayingOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_ = c.UpdateNowPlaying(context.Background(), NowPlayingParams{
		Artist: "Artist",
		Title:  "Track",
	})

	// Optional fields must not be present when zero.
	for _, key := range []string{"album", "albumArtist", "duration", "trackNumber", "mbid", "context"} {
		if v := captured.Get(key); v != "" {
			t.Errorf("optional field %q should be absent, got %q", key, v)
		}
	}
}

func TestScrobble_AllOptionalFields(t *testing.T) {
	srv, captured := captureServer(scrobbleOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	chosen := true
	err := c.Scrobble(context.Background(), ScrobbleParams{
		Artist:       "Iron Maiden",
		Title:        "The Nomad",
		Timestamp:    1609459200,
		Album:        "Dance of Death",
		AlbumArtist:  "Iron Maiden",
		TrackNumber:  6,
		Duration:     613,
		StreamID:     "stream1",
		Context:      "ctx",
		MBID:         "abc-123",
		ChosenByUser: &chosen,
	})
	if err != nil {
		t.Fatalf("Scrobble: %v", err)
	}

	checks := map[string]string{
		"albumArtist[0]":  "Iron Maiden",
		"trackNumber[0]":  "6",
		"duration[0]":     "613",
		"streamID[0]":     "stream1",
		"context[0]":      "ctx",
		"mbid[0]":         "abc-123",
		"chosenByUser[0]": "1",
	}
	for key, want := range checks {
		if got := captured.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestUpdateNowPlaying_AllOptionalFields(t *testing.T) {
	srv, captured := captureServer(nowPlayingOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.UpdateNowPlaying(context.Background(), NowPlayingParams{
		Artist:      "Iron Maiden",
		Title:       "The Nomad",
		AlbumArtist: "Iron Maiden",
		TrackNumber: 6,
		MBID:        "abc-123",
		Context:     "ctx",
	})
	if err != nil {
		t.Fatalf("UpdateNowPlaying: %v", err)
	}

	checks := map[string]string{
		"albumArtist": "Iron Maiden",
		"trackNumber": "6",
		"mbid":        "abc-123",
		"context":     "ctx",
	}
	for key, want := range checks {
		if got := captured.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestUpdateNowPlaying_WSError(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.UpdateNowPlaying(context.Background(), NowPlayingParams{
		Artist: "Iron Maiden",
		Title:  "The Nomad",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "UpdateNowPlaying") {
		t.Errorf("error should mention UpdateNowPlaying, got: %v", err)
	}
}

func TestScrobble_WSError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(sampleErrorXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.Scrobble(context.Background(), ScrobbleParams{
		Artist:    "X",
		Title:     "Y",
		Timestamp: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ScrobbleMany") {
		t.Errorf("error should mention ScrobbleMany, got: %v", err)
	}
}
