package lastfm

import (
	"context"
	"testing"
)

func TestAlbum_GetTitle(t *testing.T) {
	c := NewLastFMClient("k", "s")
	a := c.GetAlbum("Iron Maiden", "Piece of Mind")
	if a.GetTitle() != "Piece of Mind" {
		t.Errorf("GetTitle() = %q", a.GetTitle())
	}
}

func TestAlbum_GetArtist(t *testing.T) {
	c := NewLastFMClient("k", "s")
	a := c.GetAlbum("Iron Maiden", "Piece of Mind")
	if a.GetArtist().Name != "Iron Maiden" {
		t.Errorf("GetArtist().Name = %q", a.GetArtist().Name)
	}
}

func TestAlbum_String(t *testing.T) {
	c := NewLastFMClient("k", "s")
	a := c.GetAlbum("Iron Maiden", "Piece of Mind")
	s := a.String()
	if s == "" {
		t.Error("String() should be non-empty")
	}
}

func TestAlbum_GetInfo_WithUsername(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv, WithUsername("testuser"))
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo (with username): %v", err)
	}
}

func TestAlbum_GetMBID(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	mbid, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetMBID(context.Background())
	if err != nil {
		t.Fatalf("GetMBID: %v", err)
	}
	if mbid != "xyz-456" {
		t.Errorf("MBID = %q, want %q", mbid, "xyz-456")
	}
}

func TestAlbum_GetListenerCount(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetListenerCount(context.Background())
	if err != nil {
		t.Fatalf("GetListenerCount: %v", err)
	}
	if count != 300000 {
		t.Errorf("ListenerCount = %v, want 300000", count)
	}
}

func TestAlbum_GetPlaycount(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetPlaycount: %v", err)
	}
	if count != 1500000 {
		t.Errorf("Playcount = %v, want 1500000", count)
	}
}

func TestAlbum_GetUserPlaycount(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetUserPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetUserPlaycount: %v", err)
	}
}

func TestAlbum_GetCoverImage(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	url, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetCoverImage(context.Background(), SizeLarge)
	if err != nil {
		t.Fatalf("GetCoverImage: %v", err)
	}
	if url != "https://img.last.fm/large.jpg" {
		t.Errorf("CoverImage = %q, want %q", url, "https://img.last.fm/large.jpg")
	}
}

func TestAlbum_GetTracks(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetTracks(context.Background())
	if err != nil {
		t.Fatalf("GetTracks: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("len(tracks) = %d, want 2", len(tracks))
	}
	if tracks[0].Title != "Wildest Dreams" {
		t.Errorf("tracks[0].Title = %q, want %q", tracks[0].Title, "Wildest Dreams")
	}
}

func TestAlbum_GetTopTags(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetTopTags(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestAlbum_GetTags(t *testing.T) {
	srv := serveXML(userTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestAlbum_AddTags(t *testing.T) {
	srv := serveXML(loveOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetAlbum("Iron Maiden", "Dance of Death").AddTags(context.Background(), []string{"classic"})
	if err != nil {
		t.Fatalf("AddTags: %v", err)
	}
}

func TestAlbum_RemoveTag(t *testing.T) {
	srv := serveXML(loveOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetAlbum("Iron Maiden", "Dance of Death").RemoveTag(context.Background(), "classic")
	if err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
}

func TestAlbum_GetWikiSummary(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	summary, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetWikiSummary(context.Background())
	if err != nil {
		t.Fatalf("GetWikiSummary: %v", err)
	}
	if summary == "" {
		t.Error("WikiSummary should not be empty")
	}
}

func TestAlbum_GetWikiContent(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetWikiContent(context.Background())
	if err != nil {
		t.Fatalf("GetWikiContent: %v", err)
	}
}

func TestAlbum_GetWikiPublishedDate(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetWikiPublishedDate(context.Background())
	if err != nil {
		t.Fatalf("GetWikiPublishedDate: %v", err)
	}
}

func TestAlbum_GetWiki_NoWikiNode(t *testing.T) {
	srv := serveXML(noWikiAlbumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	summary, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetWikiSummary(context.Background())
	if err != nil {
		t.Fatalf("GetWikiSummary (no wiki): %v", err)
	}
	if summary != "" {
		t.Errorf("summary = %q, want empty when no wiki node", summary)
	}
}

func TestAlbum_GetURL(t *testing.T) {
	c := NewLastFMClient("k", "s")
	url := c.GetAlbum("Iron Maiden", "Piece of Mind").GetURL(DomainEnglish)
	if url == "" {
		t.Error("GetURL should return a non-empty string")
	}
}

func TestAlbum_getWiki_DefaultSection(t *testing.T) {
	// The unexported getWiki has a default switch case that returns "".
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.GetAlbum("Iron Maiden", "Dance of Death").getWiki(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("getWiki: %v", err)
	}
	if result != "" {
		t.Errorf("getWiki with unknown section = %q, want empty", result)
	}
}

func TestAlbum_GetMBID_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetMBID(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetListenerCount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetListenerCount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetPlaycount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetPlaycount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetUserPlaycount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetUserPlaycount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetCoverImage_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetCoverImage(context.Background(), SizeLarge)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetTracks_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetTracks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetWikiSummary_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetWikiSummary(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbum_GetInfo_MissingNode(t *testing.T) {
	srv := serveXML(`<lfm status="ok"><results></results></lfm>`)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for missing <album> node, got nil")
	}
}
