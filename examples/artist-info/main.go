// Command artist-info prints detailed information about an artist.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//	go run ./examples/artist-info -artist "Iron Maiden"
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
	flag.Parse()

	client := lastfm.NewLastFMClient(mustEnv("LASTFM_API_KEY"), mustEnv("LASTFM_API_SECRET"),
		lastfm.WithRateLimit(),
		lastfm.WithCache(lastfm.NewMemoryCache()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a := client.GetArtist(*artistName)

	// ── Core info ─────────────────────────────────────────────────────────────
	info, err := a.GetInfo(ctx)
	if err != nil {
		cancel()
		log.Fatalf("GetInfo: %v", err)
	}

	fmt.Printf("Artist:    %s\n", info.Name)
	fmt.Printf("Listeners: %s\n", formatInt(info.Listeners))
	fmt.Printf("Playcount: %s\n", formatInt(info.Playcount))
	fmt.Printf("MBID:      %s\n", orNone(info.MBID))
	fmt.Printf("URL:       %s\n", info.URL)

	if len(info.TopTags) > 0 {
		names := make([]string, len(info.TopTags))
		for i, tag := range info.TopTags {
			names[i] = tag.Item.Name
		}
		fmt.Printf("Tags:      %s\n", strings.Join(names, ", "))
	}

	if info.BioSummary != "" {
		summary := stripTags(info.BioSummary)
		if len(summary) > 300 {
			summary = summary[:300] + "..."
		}
		fmt.Printf("\nBio:\n  %s\n", summary)
	}

	// ── Similar artists ───────────────────────────────────────────────────────
	similar, err := a.GetSimilar(ctx, 5)
	if err != nil {
		log.Printf("GetSimilar: %v", err)
	} else if len(similar) > 0 {
		fmt.Println("\nSimilar artists:")
		for _, s := range similar {
			fmt.Printf("  %.0f%%  %s\n", s.Match*100, s.Item.Name)
		}
	}

	// ── Top albums ────────────────────────────────────────────────────────────
	albums, err := a.GetTopAlbums(ctx, 5)
	if err != nil {
		log.Printf("GetTopAlbums: %v", err)
	} else if len(albums) > 0 {
		fmt.Println("\nTop albums:")
		for _, al := range albums {
			fmt.Printf("  %-40s %s plays\n", al.Item.Title, formatInt(int64(al.Weight)))
		}
	}

	// ── Top tracks ────────────────────────────────────────────────────────────
	tracks, err := a.GetTopTracks(ctx, 5)
	if err != nil {
		log.Printf("GetTopTracks: %v", err)
	} else if len(tracks) > 0 {
		fmt.Println("\nTop tracks:")
		for _, t := range tracks {
			fmt.Printf("  %-40s %s plays\n", t.Item.Title, formatInt(int64(t.Weight)))
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

func formatInt(n int64) string {
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
