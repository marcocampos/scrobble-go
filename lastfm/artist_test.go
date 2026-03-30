package lastfm

import (
	"context"
	"testing"
)

func TestArtist_GetName(t *testing.T) {
	c := NewLastFMClient("k", "s")
	a := c.GetArtist("Iron Maiden")
	if a.GetName() != "Iron Maiden" {
		t.Errorf("GetName() = %q, want %q", a.GetName(), "Iron Maiden")
	}
}

func TestArtist_String(t *testing.T) {
	c := NewLastFMClient("k", "s")
	a := c.GetArtist("Iron Maiden")
	if a.String() != "Iron Maiden" {
		t.Errorf("String() = %q, want %q", a.String(), "Iron Maiden")
	}
}

func TestArtist_GetMBID(t *testing.T) {
	srv := serveXML(artistInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	mbid, err := c.GetArtist("Iron Maiden").GetMBID(context.Background())
	if err != nil {
		t.Fatalf("GetMBID: %v", err)
	}
	if mbid != "ca891d65-d9b0-4258-89f7-e6ba29d83767" {
		t.Errorf("MBID = %q", mbid)
	}
}

func TestArtist_GetListenerCount(t *testing.T) {
	srv := serveXML(artistInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetArtist("Iron Maiden").GetListenerCount(context.Background())
	if err != nil {
		t.Fatalf("GetListenerCount: %v", err)
	}
	if count != 3456789 {
		t.Errorf("ListenerCount = %v, want 3456789", count)
	}
}

func TestArtist_GetPlaycount(t *testing.T) {
	srv := serveXML(artistInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetArtist("Iron Maiden").GetPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetPlaycount: %v", err)
	}
	if count != 123456789 {
		t.Errorf("Playcount = %v, want 123456789", count)
	}
}

func TestArtist_GetUserPlaycount(t *testing.T) {
	srv := serveXML(artistInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetUserPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetUserPlaycount: %v", err)
	}
}

func TestArtist_GetTopTracks(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetArtist("Iron Maiden").GetTopTracks(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
	if tracks[0].Item.Title != "The Trooper" {
		t.Errorf("tracks[0].Title = %q, want %q", tracks[0].Item.Title, "The Trooper")
	}
}

func TestArtist_GetTopTags(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetArtist("Iron Maiden").GetTopTags(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
	if tags[0].Item.Name != "heavy metal" {
		t.Errorf("tags[0].Name = %q, want %q", tags[0].Item.Name, "heavy metal")
	}
}

func TestArtist_GetTopTags_Limit(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetArtist("Iron Maiden").GetTopTags(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTopTags: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1 (limit applied)", len(tags))
	}
}

func TestArtist_GetTags(t *testing.T) {
	srv := serveXML(userTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetArtist("Iron Maiden").GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestArtist_AddTags(t *testing.T) {
	srv := serveXML(loveOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetArtist("Iron Maiden").AddTags(context.Background(), []string{"classic", "nwobhm"})
	if err != nil {
		t.Fatalf("AddTags: %v", err)
	}
}

func TestArtist_RemoveTag(t *testing.T) {
	srv := serveXML(loveOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetArtist("Iron Maiden").RemoveTag(context.Background(), "classic")
	if err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
}

func TestArtist_GetURL(t *testing.T) {
	c := NewLastFMClient("k", "s")
	url := c.GetArtist("Iron Maiden").GetURL(DomainEnglish)
	if url == "" {
		t.Error("GetURL should return a non-empty string")
	}
}

func TestArtist_GetMBID_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetMBID(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetListenerCount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetListenerCount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetPlaycount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetPlaycount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetUserPlaycount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetUserPlaycount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetSimilar_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetSimilar(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetTopAlbums_WithLimit(t *testing.T) {
	srv := serveXML(artistTopAlbumsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	albums, err := c.GetArtist("Iron Maiden").GetTopAlbums(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTopAlbums: %v", err)
	}
	if len(albums) != 2 {
		t.Fatalf("len(albums) = %d, want 2", len(albums))
	}
}

func TestArtist_AddTags_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetArtist("Iron Maiden").AddTags(context.Background(), []string{"metal"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_RemoveTag_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetArtist("Iron Maiden").RemoveTag(context.Background(), "metal")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetTags_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetTags(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetTopTags_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetTopTags(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetTopAlbums_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetTopAlbums(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetTopTracks_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetTopTracks(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArtist_GetInfo_MissingNode(t *testing.T) {
	srv := serveXML(`<lfm status="ok"><results></results></lfm>`)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for missing <artist> node, got nil")
	}
}

func TestArtist_GetInfo_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArtist("Iron Maiden").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
