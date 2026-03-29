// Command search queries Last.fm for artists, albums, or tracks.
//
// Usage:
//
//	export LASTFM_API_KEY=...
//	export LASTFM_API_SECRET=...
//
//	go run ./examples/search -type artist -query "Iron Maiden"
//	go run ./examples/search -type album  -query "Piece of Mind"
//	go run ./examples/search -type track  -query "The Trooper"
//	go run ./examples/search -type track  -query "The Trooper" -artist "Iron Maiden"
//
// Pass -pages N to retrieve more than one page of results.
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
	searchType := flag.String("type", "track", "search type: artist, album, or track")
	query := flag.String("query", "", "search query (required)")
	artist := flag.String("artist", "", "narrow track search to this artist")
	pages := flag.Int("pages", 1, "number of result pages to fetch")
	flag.Parse()

	if *query == "" {
		log.Fatal("usage: search -type [artist|album|track] -query <text> [-artist <name>] [-pages N]")
	}

	client := lastfm.NewLastFMClient(mustEnv("LASTFM_API_KEY"), mustEnv("LASTFM_API_SECRET"),
		lastfm.WithRateLimit(),
		lastfm.WithCache(lastfm.NewMemoryCache()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Searching for %s: %q\n\n", *searchType, *query)

	switch strings.ToLower(*searchType) {
	case "artist":
		searchArtists(ctx, client, *query, *pages)
	case "album":
		searchAlbums(ctx, client, *query, *pages)
	case "track":
		searchTracks(ctx, client, *artist, *query, *pages)
	default:
		log.Fatalf("unknown search type %q — must be artist, album, or track", *searchType)
	}
}

func searchArtists(ctx context.Context, client *lastfm.Client, query string, pages int) {
	s := client.SearchForArtist(query)
	total := 0
	for p := 1; p <= pages; p++ {
		results, err := s.GetPage(ctx, p)
		if err != nil {
			log.Fatalf("page %d: %v", p, err)
		}
		if len(results) == 0 {
			fmt.Println("No more results.")
			break
		}
		for _, a := range results {
			fmt.Printf("  %s\n", a.Name)
			total++
		}
	}
	fmt.Printf("\n%d artist(s) shown.\n", total)
}

func searchAlbums(ctx context.Context, client *lastfm.Client, query string, pages int) {
	s := client.SearchForAlbum(query)
	total := 0
	for p := 1; p <= pages; p++ {
		results, err := s.GetPage(ctx, p)
		if err != nil {
			log.Fatalf("page %d: %v", p, err)
		}
		if len(results) == 0 {
			fmt.Println("No more results.")
			break
		}
		for _, al := range results {
			fmt.Printf("  %-40s by %s\n", al.Title, al.Artist.Name)
			total++
		}
	}
	fmt.Printf("\n%d album(s) shown.\n", total)
}

func searchTracks(ctx context.Context, client *lastfm.Client, artist, query string, pages int) {
	s := client.SearchForTrack(artist, query)
	total := 0
	for p := 1; p <= pages; p++ {
		results, err := s.GetPage(ctx, p)
		if err != nil {
			log.Fatalf("page %d: %v", p, err)
		}
		if len(results) == 0 {
			fmt.Println("No more results.")
			break
		}
		for _, t := range results {
			fmt.Printf("  %-35s by %s\n", t.Title, t.Artist.Name)
			total++
		}
	}
	fmt.Printf("\n%d track(s) shown.\n", total)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
