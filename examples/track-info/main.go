// Command track-info prints detailed information about a track.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	go run ./examples/track-info -artist "Iron Maiden" -track "The Trooper"
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
	artist := flag.String("artist", "Iron Maiden", "artist name")
	track := flag.String("track", "The Trooper", "track title")
	flag.Parse()

	client := lastfm.NewLastFMClient(mustEnv("LASTFM_API_KEY"), mustEnv("LASTFM_API_SECRET"),
		lastfm.WithRateLimit(),
		lastfm.WithCache(lastfm.NewMemoryCache()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t := client.GetTrack(*artist, *track)

	// ── Core info ─────────────────────────────────────────────────────────────
	info, err := t.GetInfo(ctx)
	if err != nil {
		log.Fatalf("GetInfo: %v", err)
	}

	fmt.Printf("Track:     %s\n", info.Title)
	fmt.Printf("Artist:    %s\n", info.Artist)
	fmt.Printf("Album:     %s\n", orNone(info.Album))
	fmt.Printf("Duration:  %s\n", formatDuration(info.Duration))
	fmt.Printf("Listeners: %s\n", formatInt(info.Listeners))
	fmt.Printf("Playcount: %s\n", formatInt(info.Playcount))
	fmt.Printf("MBID:      %s\n", orNone(info.MBID))
	fmt.Printf("URL:       %s\n", info.URL)

	if info.WikiSummary != "" {
		// Strip HTML tags for plain-text output.
		summary := stripTags(info.WikiSummary)
		if len(summary) > 300 {
			summary = summary[:300] + "..."
		}
		fmt.Printf("\nWiki summary:\n  %s\n", summary)
	}

	// ── Top tags ──────────────────────────────────────────────────────────────
	tags, err := t.GetTopTags(ctx, 5)
	if err != nil {
		log.Printf("GetTopTags: %v", err)
	} else if len(tags) > 0 {
		names := make([]string, len(tags))
		for i, tag := range tags {
			names[i] = tag.Item.Name
		}
		fmt.Printf("\nTop tags:  %s\n", strings.Join(names, ", "))
	}

	// ── Similar tracks ────────────────────────────────────────────────────────
	similar, err := t.GetSimilar(ctx, 5)
	if err != nil {
		log.Printf("GetSimilar: %v", err)
	} else if len(similar) > 0 {
		fmt.Println("\nSimilar tracks:")
		for _, s := range similar {
			fmt.Printf("  %.0f%%  %s – %s\n",
				s.Match*100, s.Item.Artist.Name, s.Item.Title)
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
	// Simple thousands separator.
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

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "(unknown)"
	}
	m, s := seconds/60, seconds%60
	return fmt.Sprintf("%d:%02d", m, s)
}

// stripTags removes HTML tags from s.
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
