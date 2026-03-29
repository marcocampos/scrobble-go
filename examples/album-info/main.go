// Command album-info prints detailed information about an album.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	go run ./examples/album-info -artist "Iron Maiden" -album "Piece of Mind"
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
	artistName := flag.String("artist", "Iron Maiden", "artist name")
	albumName := flag.String("album", "Piece of Mind", "album title")
	flag.Parse()

	client := lastfm.NewLastFMClient(mustEnv("LASTFM_API_KEY"), mustEnv("LASTFM_API_SECRET"),
		lastfm.WithRateLimit(),
		lastfm.WithCache(lastfm.NewMemoryCache()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	al := client.GetAlbum(*artistName, *albumName)

	info, err := al.GetInfo(ctx)
	if err != nil {
		log.Fatalf("GetInfo: %v", err)
	}

	// ── Core info ─────────────────────────────────────────────────────────────
	fmt.Printf("Album:     %s\n", info.Title)
	fmt.Printf("Artist:    %s\n", info.Artist)
	fmt.Printf("Listeners: %s\n", formatInt(info.Listeners))
	fmt.Printf("Playcount: %s\n", formatInt(info.Playcount))
	fmt.Printf("MBID:      %s\n", orNone(info.MBID))
	fmt.Printf("URL:       %s\n", info.URL)

	if url := info.Images[lastfm.SizeExtraLarge]; url != "" {
		fmt.Printf("Cover:     %s\n", url)
	}

	if len(info.TopTags) > 0 {
		names := make([]string, len(info.TopTags))
		for i, tag := range info.TopTags {
			names[i] = tag.Item.Name
		}
		fmt.Printf("Tags:      %s\n", strings.Join(names, ", "))
	}

	// ── Track listing ─────────────────────────────────────────────────────────
	if len(info.Tracks) > 0 {
		fmt.Printf("\nTrack listing (%d tracks):\n", len(info.Tracks))
		for i, t := range info.Tracks {
			fmt.Printf("  %2d. %s\n", i+1, t.Title)
		}
	}

	// ── Wiki ──────────────────────────────────────────────────────────────────
	if info.WikiSummary != "" {
		summary := stripTags(info.WikiSummary)
		if len(summary) > 300 {
			summary = summary[:300] + "..."
		}
		fmt.Printf("\nWiki summary:\n  %s\n", summary)
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

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
