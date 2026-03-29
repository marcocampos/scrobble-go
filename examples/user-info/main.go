// Command user-info prints a Last.fm user's profile and listening history.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	go run ./examples/user-info -user "someusername"
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/marcocampos/scrobble-go/lastfm"
)

func main() {
	username := flag.String("user", "", "Last.fm username (required)")
	flag.Parse()

	if *username == "" {
		log.Fatal("usage: user-info -user <username>")
	}

	client := lastfm.NewLastFMClient(mustEnv("LASTFM_API_KEY"), mustEnv("LASTFM_API_SECRET"),
		lastfm.WithRateLimit(),
		lastfm.WithCache(lastfm.NewMemoryCache()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	u := client.GetUser(*username)

	// ── Profile ───────────────────────────────────────────────────────────────
	info, err := u.GetInfo(ctx)
	if err != nil {
		log.Fatalf("GetInfo: %v", err)
	}

	fmt.Printf("Username:  %s\n", info.Name)
	if info.RealName != "" {
		fmt.Printf("Real name: %s\n", info.RealName)
	}
	fmt.Printf("Country:   %s\n", orNone(info.Country))
	fmt.Printf("Scrobbles: %s\n", formatInt(info.Playcount))
	fmt.Printf("URL:       %s\n", info.URL)

	// ── Now playing ───────────────────────────────────────────────────────────
	np, err := u.GetNowPlaying(ctx)
	if err != nil {
		log.Printf("GetNowPlaying: %v", err)
	} else if np != nil {
		fmt.Printf("\nNow playing: %s – %s\n", np.Artist.Name, np.Title)
	}

	// ── Recent tracks ─────────────────────────────────────────────────────────
	recent, err := u.GetRecentTracks(ctx, 10, 0)
	if err != nil {
		log.Printf("GetRecentTracks: %v", err)
	} else if len(recent) > 0 {
		fmt.Printf("\nRecent scrobbles:\n")
		for _, pt := range recent {
			fmt.Printf("  %-35s %s – %s\n",
				pt.Track.Artist.Name+" – "+pt.Track.Title,
				"", pt.PlaybackDate)
		}
	}

	// ── Top artists (overall) ─────────────────────────────────────────────────
	topArtists, err := u.GetTopArtists(ctx, lastfm.PeriodOverall, 5)
	if err != nil {
		log.Printf("GetTopArtists: %v", err)
	} else if len(topArtists) > 0 {
		fmt.Println("\nTop artists (all time):")
		for _, a := range topArtists {
			fmt.Printf("  %-35s %s plays\n", a.Item.Name, formatInt(int(a.Weight)))
		}
	}

	// ── Top tracks (last 7 days) ──────────────────────────────────────────────
	topTracks, err := u.GetTopTracks(ctx, lastfm.Period7Days, 5)
	if err != nil {
		log.Printf("GetTopTracks: %v", err)
	} else if len(topTracks) > 0 {
		fmt.Println("\nTop tracks (last 7 days):")
		for _, t := range topTracks {
			fmt.Printf("  %-35s %s plays\n",
				t.Item.Artist.Name+" – "+t.Item.Title,
				formatInt(int(t.Weight)))
		}
	}

	// ── Loved tracks ─────────────────────────────────────────────────────────
	loved, err := u.GetLovedTracks(ctx, 5, 0)
	if err != nil {
		log.Printf("GetLovedTracks: %v", err)
	} else if len(loved) > 0 {
		fmt.Println("\nRecently loved:")
		for _, lt := range loved {
			fmt.Printf("  %s – %s\n", lt.Track.Artist.Name, lt.Track.Title)
		}
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}
