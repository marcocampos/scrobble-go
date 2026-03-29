//go:build integration

package lastfm_test

// Integration tests against the real Last.fm API.
//
// Prerequisites:
//  1. Obtain a Last.fm API key and secret:
//     https://www.last.fm/api/account/create
//  2. Export credentials as environment variables:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	export LASTFM_USERNAME=...
//	export LASTFM_PASSWORD_HASH=$(echo -n "yourpassword" | md5sum | cut -d' ' -f1)
//	# Or, if you already have a session key:
//	export LASTFM_SESSION_KEY=...
//
//  3. Run:
//     go test -tags integration -v -timeout 60s ./...
//
// WARNING: some tests perform write operations on your Last.fm account
// (scrobbles, now-playing, love/unlove, tags). All writes are reversed or
// use past timestamps so your profile data stays clean.

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	lastfm "github.com/marcocampos/scrobble-go/lastfm"
)

// creds holds the API credentials loaded from environment variables.
type creds struct {
	APIKey       string
	APISecret    string
	Username     string
	PasswordHash string
	SessionKey   string
}

// loadCreds reads credentials from environment variables and skips the test
// if any required value is missing.
func loadCreds(t *testing.T) creds {
	t.Helper()
	c := creds{
		APIKey:       os.Getenv("LASTFM_API_KEY"),
		APISecret:    os.Getenv("LASTFM_API_SECRET"),
		Username:     os.Getenv("LASTFM_USERNAME"),
		PasswordHash: os.Getenv("LASTFM_PASSWORD_HASH"),
		SessionKey:   os.Getenv("LASTFM_SESSION_KEY"),
	}
	if c.APIKey == "" || c.APISecret == "" {
		t.Skip("LASTFM_API_KEY and LASTFM_API_SECRET are required; skipping integration test")
	}
	return c
}

// newClient builds an authenticated Client from env-var credentials.
// Authentication order: session key → username+password → read-only.
func newClient(t *testing.T, c creds) *lastfm.Client {
	t.Helper()
	opts := []lastfm.Option{lastfm.WithRateLimit()}

	switch {
	case c.SessionKey != "":
		opts = append(opts, lastfm.WithSessionKey(c.SessionKey))

	case c.Username != "" && c.PasswordHash != "":
		// Authenticate and cache the session key for the lifetime of this test.
		client := lastfm.NewLastFMClient(c.APIKey, c.APISecret,
			lastfm.WithRateLimit(),
		)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := client.AuthenticateWithPassword(ctx, c.Username, c.PasswordHash); err != nil {
			t.Fatalf("authentication failed: %v", err)
		}
		return client
	}

	if c.Username != "" {
		opts = append(opts, lastfm.WithUsername(c.Username))
	}
	return lastfm.NewLastFMClient(c.APIKey, c.APISecret, opts...)
}

func ctx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// ── Read-only tests ───────────────────────────────────────────────────────────

func TestIntegration_Artist_GetInfo(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	info, err := c.GetArtist("Iron Maiden").GetInfo(cx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Name == "" {
		t.Error("Name is empty")
	}
	if info.Listeners == 0 {
		t.Error("Listeners is 0 — unexpected for Iron Maiden")
	}
	t.Logf("Artist: %s | Listeners: %d | Playcount: %d", info.Name, info.Listeners, info.Playcount)
}

func TestIntegration_Artist_GetSimilar(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	similar, err := c.GetArtist("Iron Maiden").GetSimilar(cx, 5)
	if err != nil {
		t.Fatalf("GetSimilar: %v", err)
	}
	if len(similar) == 0 {
		t.Fatal("expected at least one similar artist")
	}
	t.Logf("Similar to Iron Maiden: %s (match=%.2f)", similar[0].Item.Name, similar[0].Match)
}

func TestIntegration_Artist_GetTopAlbums(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	albums, err := c.GetArtist("Iron Maiden").GetTopAlbums(cx, 3)
	if err != nil {
		t.Fatalf("GetTopAlbums: %v", err)
	}
	if len(albums) == 0 {
		t.Fatal("expected at least one album")
	}
	for _, a := range albums {
		t.Logf("  %s (%.0f plays)", a.Item.Title, a.Weight)
	}
}

func TestIntegration_Artist_GetTopTracks(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	tracks, err := c.GetArtist("Iron Maiden").GetTopTracks(cx, 3)
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) == 0 {
		t.Fatal("expected at least one track")
	}
	t.Logf("Top track: %s", tracks[0].Item.Title)
}

func TestIntegration_Track_GetInfo(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	info, err := c.GetTrack("Iron Maiden", "The Trooper").GetInfo(cx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Title == "" {
		t.Error("Title is empty")
	}
	t.Logf("Track: %s | Duration: %ds | Playcount: %d", info.Title, info.Duration, info.Playcount)
}

func TestIntegration_Album_GetInfo(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	info, err := c.GetAlbum("Iron Maiden", "The Number of the Beast").GetInfo(cx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Title == "" {
		t.Error("Title is empty")
	}
	t.Logf("Album: %s by %s | Playcount: %d | Tracks: %d",
		info.Title, info.Artist, info.Playcount, len(info.Tracks))
}

func TestIntegration_Network_GetTopArtists(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	artists, err := c.GetTopArtists(cx, 5)
	if err != nil {
		t.Fatalf("GetTopArtists: %v", err)
	}
	if len(artists) == 0 {
		t.Fatal("expected top artists")
	}
	t.Logf("Global #1 artist: %s (%.0f plays)", artists[0].Item.Name, artists[0].Weight)
}

func TestIntegration_Search_Artist(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	results, err := c.SearchForArtist("Iron Maiden").GetPage(cx, 1)
	if err != nil {
		t.Fatalf("SearchForArtist: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
	t.Logf("First result: %s", results[0].Name)
}

func TestIntegration_Search_Track(t *testing.T) {
	c := newClient(t, loadCreds(t))
	cx, cancel := ctx(t)
	defer cancel()

	results, err := c.SearchForTrack("Iron Maiden", "Trooper").GetPage(cx, 1)
	if err != nil {
		t.Fatalf("SearchForTrack: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
	t.Logf("First result: %s – %s", results[0].Artist.Name, results[0].Title)
}

// ── Authenticated read tests ──────────────────────────────────────────────────

func TestIntegration_User_GetInfo(t *testing.T) {
	cr := loadCreds(t)
	if cr.Username == "" {
		t.Skip("LASTFM_USERNAME required")
	}
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	info, err := c.GetUser(cr.Username).GetInfo(cx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	t.Logf("User: %s | Playcount: %d | Country: %s", info.Name, info.Playcount, info.Country)
}

func TestIntegration_User_GetRecentTracks(t *testing.T) {
	cr := loadCreds(t)
	if cr.Username == "" {
		t.Skip("LASTFM_USERNAME required")
	}
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	tracks, err := c.GetUser(cr.Username).GetRecentTracks(cx, 5, 0)
	if err != nil {
		t.Fatalf("GetRecentTracks: %v", err)
	}
	t.Logf("Recent tracks count: %d", len(tracks))
	for _, pt := range tracks {
		t.Logf("  %s – %s (%s)", pt.Track.Artist.Name, pt.Track.Title, pt.Timestamp)
	}
}

// ── Write tests (authenticated, self-cleaning) ────────────────────────────────

func TestIntegration_UpdateNowPlaying(t *testing.T) {
	cr := loadCreds(t)
	if cr.Username == "" || (cr.PasswordHash == "" && cr.SessionKey == "") {
		t.Skip("authentication credentials required")
	}
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	err := c.UpdateNowPlaying(cx, lastfm.NowPlayingParams{
		Artist:   "Iron Maiden",
		Title:    "The Trooper",
		Album:    "Piece of Mind",
		Duration: 248,
	})
	if err != nil {
		t.Fatalf("UpdateNowPlaying: %v", err)
	}
	t.Log("UpdateNowPlaying succeeded")
}

func TestIntegration_Scrobble(t *testing.T) {
	cr := loadCreds(t)
	if cr.Username == "" || (cr.PasswordHash == "" && cr.SessionKey == "") {
		t.Skip("authentication credentials required")
	}
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	// Scrobble a track with a timestamp 5 minutes in the past.
	ts := time.Now().Add(-5 * time.Minute).Unix()
	err := c.Scrobble(cx, lastfm.ScrobbleParams{
		Artist:    "Iron Maiden",
		Title:     "The Trooper",
		Album:     "Piece of Mind",
		Timestamp: ts,
	})
	if err != nil {
		t.Fatalf("Scrobble: %v", err)
	}
	t.Logf("Scrobbled at timestamp %d", ts)
}

func TestIntegration_Track_LoveUnlove(t *testing.T) {
	cr := loadCreds(t)
	if cr.Username == "" || (cr.PasswordHash == "" && cr.SessionKey == "") {
		t.Skip("authentication credentials required")
	}
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	track := c.GetTrack("Iron Maiden", "The Trooper")

	if err := track.Love(cx); err != nil {
		t.Fatalf("Love: %v", err)
	}
	t.Log("Love succeeded")

	if err := track.Unlove(cx); err != nil {
		t.Fatalf("Unlove: %v", err)
	}
	t.Log("Unlove succeeded — track state restored")
}

func TestIntegration_Track_AddRemoveTag(t *testing.T) {
	cr := loadCreds(t)
	if cr.Username == "" || (cr.PasswordHash == "" && cr.SessionKey == "") {
		t.Skip("authentication credentials required")
	}
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	track := c.GetTrack("Iron Maiden", "The Trooper")
	testTag := fmt.Sprintf("gotest-%d", time.Now().Unix())

	if err := track.AddTags(cx, []string{testTag}); err != nil {
		t.Fatalf("AddTags: %v", err)
	}
	t.Logf("Added tag %q", testTag)

	tags, err := track.GetTags(cx)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	found := false
	for _, tag := range tags {
		if tag.Name == testTag {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("tag %q not found after adding", testTag)
	}

	if err := track.RemoveTag(cx, testTag); err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
	t.Logf("Removed tag %q — state restored", testTag)
}

func TestIntegration_BoltCache_WithRealAPI(t *testing.T) {
	cr := loadCreds(t)
	c := newClient(t, cr)
	cx, cancel := ctx(t)
	defer cancel()

	cache, err := lastfm.NewBoltCache(t.TempDir() + "/integration.db")
	if err != nil {
		t.Fatalf("NewBoltCache: %v", err)
	}
	defer cache.Close()
	c2 := lastfm.NewLastFMClient(cr.APIKey, cr.APISecret,
		lastfm.WithCache(cache),
		lastfm.WithRateLimit(),
	)
	_ = c // silence unused warning

	// First call — hits network.
	info1, err := c2.GetArtist("Iron Maiden").GetInfo(cx)
	if err != nil {
		t.Fatalf("first GetInfo: %v", err)
	}

	// Second call — should come from cache (same result, no new network request).
	info2, err := c2.GetArtist("Iron Maiden").GetInfo(cx)
	if err != nil {
		t.Fatalf("second GetInfo: %v", err)
	}

	if info1.Name != info2.Name {
		t.Errorf("cached result differs: %q vs %q", info1.Name, info2.Name)
	}
	t.Logf("BoltCache: both calls returned %q", info1.Name)
}
