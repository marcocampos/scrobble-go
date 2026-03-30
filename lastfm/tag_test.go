package lastfm

import (
	"context"
	"testing"
)

func TestTag_GetName(t *testing.T) {
	c := NewLastFMClient("k", "s")
	tag := c.GetTag("heavy metal")
	if tag.GetName() != "heavy metal" {
		t.Errorf("GetName() = %q", tag.GetName())
	}
}

func TestTag_String(t *testing.T) {
	c := NewLastFMClient("k", "s")
	tag := c.GetTag("heavy metal")
	if tag.String() != "heavy metal" {
		t.Errorf("String() = %q", tag.String())
	}
}

func TestTag_GetInfo(t *testing.T) {
	srv := serveXML(tagInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	info, err := c.GetTag("heavy metal").GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Name != "heavy metal" {
		t.Errorf("Name = %q, want %q", info.Name, "heavy metal")
	}
	if info.Reach != 12345 {
		t.Errorf("Reach = %d, want 12345", info.Reach)
	}
	if info.Taggings != 67890 {
		t.Errorf("Taggings = %d, want 67890", info.Taggings)
	}
	if info.WikiSummary == "" {
		t.Error("WikiSummary should not be empty")
	}
	if info.WikiContent == "" {
		t.Error("WikiContent should not be empty")
	}
}

func TestTag_GetInfo_MissingNode(t *testing.T) {
	srv := serveXML(`<lfm status="ok"><results></results></lfm>`)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTag("heavy metal").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for missing <tag> node, got nil")
	}
}

func TestTag_GetTopArtists(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artists, err := c.GetTag("heavy metal").GetTopArtists(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopArtists: %v", err)
	}
	if len(artists) != 2 {
		t.Fatalf("len(artists) = %d, want 2", len(artists))
	}
}

func TestTag_GetTopTracks(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetTag("heavy metal").GetTopTracks(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
}

func TestTag_GetInfo_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTag("heavy metal").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTag_GetTopArtists_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTag("heavy metal").GetTopArtists(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTag_GetTopTracks_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTag("heavy metal").GetTopTracks(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTag_GetTopAlbums(t *testing.T) {
	srv := serveXML(topAlbumsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	albums, err := c.GetTag("heavy metal").GetTopAlbums(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopAlbums: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("len(albums) = %d, want 1", len(albums))
	}
}

func TestTag_GetTopAlbums_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTag("heavy metal").GetTopAlbums(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
