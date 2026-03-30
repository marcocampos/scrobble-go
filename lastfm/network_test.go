package lastfm

import (
	"context"
	"testing"
)

func TestClient_GetTag(t *testing.T) {
	c := NewLastFMClient("k", "s")
	tag := c.GetTag("heavy metal")
	if tag.Name != "heavy metal" {
		t.Errorf("Name = %q", tag.Name)
	}
}

func TestClient_GetCountry(t *testing.T) {
	c := NewLastFMClient("k", "s")
	co := c.GetCountry("Germany")
	if co.Name != "Germany" {
		t.Errorf("Name = %q", co.Name)
	}
}

func TestClient_GetAuthenticatedUser(t *testing.T) {
	c := NewLastFMClient("k", "s", WithUsername("testuser"))
	u := c.GetAuthenticatedUser()
	if u.Name != "testuser" {
		t.Errorf("Name = %q, want %q", u.Name, "testuser")
	}
}

func TestClient_GetTrackByMBID(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	track, err := c.GetTrackByMBID(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("GetTrackByMBID: %v", err)
	}
	if track == nil {
		t.Fatal("expected track, got nil")
	}
}

func TestClient_GetArtistByMBID(t *testing.T) {
	srv := serveXML(artistInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artist, err := c.GetArtistByMBID(context.Background(), "ca891d65-d9b0-4258-89f7-e6ba29d83767")
	if err != nil {
		t.Fatalf("GetArtistByMBID: %v", err)
	}
	if artist.Name != "Iron Maiden" {
		t.Errorf("Name = %q, want %q", artist.Name, "Iron Maiden")
	}
}

func TestClient_GetAlbumByMBID(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	album, err := c.GetAlbumByMBID(context.Background(), "xyz-456")
	if err != nil {
		t.Fatalf("GetAlbumByMBID: %v", err)
	}
	if album == nil {
		t.Fatal("expected album, got nil")
	}
}

func TestClient_GetTopArtists(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artists, err := c.GetTopArtists(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopArtists: %v", err)
	}
	if len(artists) != 2 {
		t.Fatalf("len(artists) = %d, want 2", len(artists))
	}
}

func TestClient_GetTopArtists_NoLimit(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTopArtists(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopArtists (no limit): %v", err)
	}
}

func TestClient_GetTopTracks(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetTopTracks(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
}

func TestClient_GetTopTags(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetTopTags(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTopTags: %v", err)
	}
	// limit = 1 should truncate the 2 results to 1
	if len(tags) != 1 {
		t.Fatalf("len(tags) = %d, want 1 (limit applied)", len(tags))
	}
}

func TestClient_GetTopTags_NoLimit(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetTopTags(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopTags (no limit): %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestClient_GetGeoTopArtists(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artists, err := c.GetGeoTopArtists(context.Background(), "Germany", 5)
	if err != nil {
		t.Fatalf("GetGeoTopArtists: %v", err)
	}
	if len(artists) != 2 {
		t.Fatalf("len(artists) = %d, want 2", len(artists))
	}
}

func TestClient_GetGeoTopTracks(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetGeoTopTracks(context.Background(), "Germany", "", 5)
	if err != nil {
		t.Fatalf("GetGeoTopTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
}

func TestClient_ErrorResponses(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, c *Client) error
	}{
		{"GetTrackByMBID", func(ctx context.Context, c *Client) error { _, err := c.GetTrackByMBID(ctx, "abc-123"); return err }},
		{"GetArtistByMBID", func(ctx context.Context, c *Client) error { _, err := c.GetArtistByMBID(ctx, "abc-123"); return err }},
		{"GetAlbumByMBID", func(ctx context.Context, c *Client) error { _, err := c.GetAlbumByMBID(ctx, "abc-123"); return err }},
		{"GetTopArtists", func(ctx context.Context, c *Client) error { _, err := c.GetTopArtists(ctx, 5); return err }},
		{"GetTopTracks", func(ctx context.Context, c *Client) error { _, err := c.GetTopTracks(ctx, 5); return err }},
		{"GetTopTags", func(ctx context.Context, c *Client) error { _, err := c.GetTopTags(ctx, 5); return err }},
		{"GetGeoTopArtists", func(ctx context.Context, c *Client) error {
			_, err := c.GetGeoTopArtists(ctx, "Germany", 5)
			return err
		}},
		{"GetGeoTopTracks", func(ctx context.Context, c *Client) error {
			_, err := c.GetGeoTopTracks(ctx, "Germany", "", 5)
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := serveXML(sampleErrorXML)
			defer srv.Close()
			c := newTestClient(t, srv)
			if err := tt.call(context.Background(), c); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestClient_GetGeoTopTracks_WithLocation(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetGeoTopTracks(context.Background(), "Germany", "Berlin", 5)
	if err != nil {
		t.Fatalf("GetGeoTopTracks (with location): %v", err)
	}
}
