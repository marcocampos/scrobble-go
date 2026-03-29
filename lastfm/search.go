package lastfm

import (
	"context"
	"fmt"
	"iter"
)

// ── Artist search ─────────────────────────────────────────────────────────────

// ArtistSearch holds a paginated search for artists by name.
type ArtistSearch struct {
	query  string
	client *Client
	page   int
}

// SearchForArtist creates an ArtistSearch. Call GetNextPage to retrieve results.
func (c *Client) SearchForArtist(query string) *ArtistSearch {
	return &ArtistSearch{query: query, client: c, page: 0}
}

// GetNextPage fetches the next page of results.
// Returns an empty slice when there are no more results.
func (s *ArtistSearch) GetNextPage(ctx context.Context) ([]*Artist, error) {
	s.page++
	return s.GetPage(ctx, s.page)
}

// GetPage fetches a specific page of results (1-indexed).
func (s *ArtistSearch) GetPage(ctx context.Context, page int) ([]*Artist, error) {
	params := map[string]string{
		"artist": s.query,
		"page":   fmt.Sprintf("%d", page),
	}
	doc, err := newAPIRequest(s.client, "artist.search", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("ArtistSearch.GetPage: %w", err)
	}

	var result []*Artist
	for _, node := range doc.findAll("artist") {
		name := extract(node, "name")
		if name != "" {
			result = append(result, newArtist(name, s.client))
		}
	}
	return result, nil
}

// All returns an iterator over every artist across all pages.
// Use it with a range-over-func loop (Go 1.23+):
//
//	for artist, err := range search.All(ctx) {
//	    if err != nil { /* handle */ }
//	    fmt.Println(artist.Name)
//	}
func (s *ArtistSearch) All(ctx context.Context) iter.Seq2[*Artist, error] {
	return func(yield func(*Artist, error) bool) {
		for page := 1; ; page++ {
			results, err := s.GetPage(ctx, page)
			if err != nil {
				yield(nil, err)
				return
			}
			if len(results) == 0 {
				return
			}
			for _, a := range results {
				if !yield(a, nil) {
					return
				}
			}
		}
	}
}

// GetTotalResultCount returns the total number of matching artists.
func (s *ArtistSearch) GetTotalResultCount(ctx context.Context) (int, error) {
	params := map[string]string{"artist": s.query, "page": "1"}
	doc, err := newAPIRequest(s.client, "artist.search", params).execute(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("ArtistSearch.GetTotalResultCount: %w", err)
	}
	return parseInt(extract(doc, "totalResults")), nil
}

// ── Album search ──────────────────────────────────────────────────────────────

// AlbumSearch holds a paginated search for albums by name.
type AlbumSearch struct {
	query  string
	client *Client
	page   int
}

// SearchForAlbum creates an AlbumSearch. Call GetNextPage to retrieve results.
func (c *Client) SearchForAlbum(query string) *AlbumSearch {
	return &AlbumSearch{query: query, client: c, page: 0}
}

// GetNextPage fetches the next page of results.
func (s *AlbumSearch) GetNextPage(ctx context.Context) ([]*Album, error) {
	s.page++
	return s.GetPage(ctx, s.page)
}

// GetPage fetches a specific page of results (1-indexed).
func (s *AlbumSearch) GetPage(ctx context.Context, page int) ([]*Album, error) {
	params := map[string]string{
		"album": s.query,
		"page":  fmt.Sprintf("%d", page),
	}
	doc, err := newAPIRequest(s.client, "album.search", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("AlbumSearch.GetPage: %w", err)
	}

	var result []*Album
	for _, node := range doc.findAll("album") {
		name := extract(node, "name")
		artist := extract(node, "artist")
		if name != "" {
			result = append(result, newAlbum(artist, name, s.client))
		}
	}
	return result, nil
}

// All returns an iterator over every album across all pages.
// Use it with a range-over-func loop (Go 1.23+):
//
//	for album, err := range search.All(ctx) {
//	    if err != nil { /* handle */ }
//	    fmt.Println(album.Title)
//	}
func (s *AlbumSearch) All(ctx context.Context) iter.Seq2[*Album, error] {
	return func(yield func(*Album, error) bool) {
		for page := 1; ; page++ {
			results, err := s.GetPage(ctx, page)
			if err != nil {
				yield(nil, err)
				return
			}
			if len(results) == 0 {
				return
			}
			for _, a := range results {
				if !yield(a, nil) {
					return
				}
			}
		}
	}
}

// GetTotalResultCount returns the total number of matching albums.
func (s *AlbumSearch) GetTotalResultCount(ctx context.Context) (int, error) {
	params := map[string]string{"album": s.query, "page": "1"}
	doc, err := newAPIRequest(s.client, "album.search", params).execute(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("AlbumSearch.GetTotalResultCount: %w", err)
	}
	return parseInt(extract(doc, "totalResults")), nil
}

// ── Track search ──────────────────────────────────────────────────────────────

// TrackSearch holds a paginated search for tracks by name and optional artist.
type TrackSearch struct {
	artist string
	track  string
	client *Client
	page   int
}

// SearchForTrack creates a TrackSearch. Pass an empty string for artist to
// search all artists. Call GetNextPage to retrieve results.
func (c *Client) SearchForTrack(artist, track string) *TrackSearch {
	return &TrackSearch{artist: artist, track: track, client: c, page: 0}
}

// GetNextPage fetches the next page of results.
func (s *TrackSearch) GetNextPage(ctx context.Context) ([]*Track, error) {
	s.page++
	return s.GetPage(ctx, s.page)
}

// GetPage fetches a specific page of results (1-indexed).
func (s *TrackSearch) GetPage(ctx context.Context, page int) ([]*Track, error) {
	params := map[string]string{
		"track": s.track,
		"page":  fmt.Sprintf("%d", page),
	}
	if s.artist != "" {
		params["artist"] = s.artist
	}
	doc, err := newAPIRequest(s.client, "track.search", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("TrackSearch.GetPage: %w", err)
	}

	var result []*Track
	for _, node := range doc.findAll("track") {
		title := extract(node, "name")
		artist := extract(node, "artist")
		if title != "" {
			result = append(result, newTrack(artist, title, s.client))
		}
	}
	return result, nil
}

// All returns an iterator over every track across all pages.
// Use it with a range-over-func loop (Go 1.23+):
//
//	for track, err := range search.All(ctx) {
//	    if err != nil { /* handle */ }
//	    fmt.Println(track.Title)
//	}
func (s *TrackSearch) All(ctx context.Context) iter.Seq2[*Track, error] {
	return func(yield func(*Track, error) bool) {
		for page := 1; ; page++ {
			results, err := s.GetPage(ctx, page)
			if err != nil {
				yield(nil, err)
				return
			}
			if len(results) == 0 {
				return
			}
			for _, t := range results {
				if !yield(t, nil) {
					return
				}
			}
		}
	}
}

// GetTotalResultCount returns the total number of matching tracks.
func (s *TrackSearch) GetTotalResultCount(ctx context.Context) (int, error) {
	params := map[string]string{"track": s.track, "page": "1"}
	if s.artist != "" {
		params["artist"] = s.artist
	}
	doc, err := newAPIRequest(s.client, "track.search", params).execute(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("TrackSearch.GetTotalResultCount: %w", err)
	}
	return parseInt(extract(doc, "totalResults")), nil
}
