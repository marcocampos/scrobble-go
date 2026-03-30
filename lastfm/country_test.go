package lastfm

import (
	"context"
	"testing"
)

func TestCountry_GetName(t *testing.T) {
	c := NewLastFMClient("k", "s")
	co := c.GetCountry("Germany")
	if co.GetName() != "Germany" {
		t.Errorf("GetName() = %q", co.GetName())
	}
}

func TestCountry_String(t *testing.T) {
	c := NewLastFMClient("k", "s")
	co := c.GetCountry("Germany")
	if co.String() != "Germany" {
		t.Errorf("String() = %q", co.String())
	}
}

func TestCountry_GetTopArtists(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artists, err := c.GetCountry("Germany").GetTopArtists(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopArtists: %v", err)
	}
	if len(artists) != 2 {
		t.Fatalf("len(artists) = %d, want 2", len(artists))
	}
	if artists[0].Item.Name != "Iron Maiden" {
		t.Errorf("artists[0].Name = %q", artists[0].Item.Name)
	}
}

func TestCountry_GetTopTracks(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetCountry("Germany").GetTopTracks(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
}

func TestCountry_ErrorResponses(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, co *Country) error
	}{
		{"GetTopArtists", func(ctx context.Context, co *Country) error { _, err := co.GetTopArtists(ctx, 5); return err }},
		{"GetTopTracks", func(ctx context.Context, co *Country) error { _, err := co.GetTopTracks(ctx, 5); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := serveXML(sampleErrorXML)
			defer srv.Close()
			c := newTestClient(t, srv)
			if err := tt.call(context.Background(), c.GetCountry("Germany")); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestCountry_GetURL(t *testing.T) {
	c := NewLastFMClient("k", "s")
	url := c.GetCountry("Germany").GetURL(DomainEnglish)
	if url == "" {
		t.Error("GetURL should return a non-empty string")
	}
}
