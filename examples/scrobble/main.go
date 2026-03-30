// Command scrobble demonstrates how to authenticate with Last.fm using a
// username and password, and how to submit a scrobble and a now-playing
// notification.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	export LASTFM_USERNAME=...
//	export LASTFM_PASSWORD=...
//	go run ./examples/scrobble
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/marcocampos/scrobble-go/lastfm"
)

func main() {
	apiKey := mustEnv("LASTFM_API_KEY")
	apiSecret := mustEnv("LASTFM_API_SECRET")
	username := mustEnv("LASTFM_USERNAME")
	password := mustEnv("LASTFM_PASSWORD")

	// Hash the password as required by the Last.fm API.
	passwordHash := lastfm.MD5(password)

	// Create an authenticated client. WithRateLimit ensures we stay within
	// the API's 5 req/s guideline.
	client := lastfm.NewLastFMClient(apiKey, apiSecret,
		lastfm.WithRateLimit(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	fmt.Printf("Authenticating as %q...\n", username)
	if err := client.AuthenticateWithPassword(ctx, username, passwordHash); err != nil {
		cancel()
		log.Fatalf("authentication failed: %v", err)
	}
	fmt.Println("Authenticated.")

	// ── Now-playing notification ──────────────────────────────────────────────
	fmt.Println("Sending now-playing notification...")
	err := client.UpdateNowPlaying(ctx, lastfm.NowPlayingParams{
		Artist:   "Iron Maiden",
		Title:    "The Trooper",
		Album:    "Piece of Mind",
		Duration: 248,
	})
	if err != nil {
		cancel()
		log.Fatalf("UpdateNowPlaying: %v", err)
	}
	fmt.Println("Now-playing sent.")

	// ── Scrobble ──────────────────────────────────────────────────────────────
	// Timestamp must be within the last two weeks for Last.fm to accept it.
	ts := time.Now().Add(-3 * time.Minute).Unix()
	fmt.Printf("Scrobbling track (timestamp %d)...\n", ts)
	err = client.Scrobble(ctx, lastfm.ScrobbleParams{
		Artist:    "Iron Maiden",
		Title:     "The Trooper",
		Album:     "Piece of Mind",
		Timestamp: ts,
	})
	if err != nil {
		cancel()
		log.Fatalf("Scrobble: %v", err)
	}
	fmt.Println("Scrobble accepted.")

	// ── Read back recent tracks to confirm ────────────────────────────────────
	fmt.Println("Fetching recent tracks...")
	tracks, err := client.GetUser(username).GetRecentTracks(ctx, 3, 0)
	if err != nil {
		cancel()
		log.Fatalf("GetRecentTracks: %v", err)
	}
	fmt.Printf("Last %d scrobbles:\n", len(tracks))
	for _, t := range tracks {
		fmt.Printf("  %s – %s\n", t.Track.Artist.Name, t.Track.Title)
	}
	cancel()
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
