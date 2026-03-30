package lastfm

import (
	"context"
	"testing"
)

func TestTrack_GetTitle(t *testing.T) {
	c := NewLastFMClient("k", "s")
	tr := c.GetTrack("Iron Maiden", "The Trooper")
	if tr.GetTitle() != "The Trooper" {
		t.Errorf("GetTitle() = %q", tr.GetTitle())
	}
}

func TestTrack_GetArtist(t *testing.T) {
	c := NewLastFMClient("k", "s")
	tr := c.GetTrack("Iron Maiden", "The Trooper")
	if tr.GetArtist().Name != "Iron Maiden" {
		t.Errorf("GetArtist().Name = %q", tr.GetArtist().Name)
	}
}

func TestTrack_String(t *testing.T) {
	c := NewLastFMClient("k", "s")
	tr := c.GetTrack("Iron Maiden", "The Trooper")
	s := tr.String()
	if s == "" {
		t.Error("String() should be non-empty")
	}
}

func TestTrack_GetInfo_WithUsername(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv, WithUsername("testuser"))
	_, err := c.GetTrack("Iron Maiden", "The Nomad").GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo (with username): %v", err)
	}
}

func TestTrack_GetMBID(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	mbid, err := c.GetTrack("Iron Maiden", "The Nomad").GetMBID(context.Background())
	if err != nil {
		t.Fatalf("GetMBID: %v", err)
	}
	if mbid != "abc-123" {
		t.Errorf("MBID = %q, want %q", mbid, "abc-123")
	}
}

func TestTrack_GetDuration(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	dur, err := c.GetTrack("Iron Maiden", "The Nomad").GetDuration(context.Background())
	if err != nil {
		t.Fatalf("GetDuration: %v", err)
	}
	// trackInfoXML has duration 613000ms → 613s
	if dur != 613 {
		t.Errorf("Duration = %d, want 613", dur)
	}
}

func TestTrack_GetListenerCount(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetTrack("Iron Maiden", "The Nomad").GetListenerCount(context.Background())
	if err != nil {
		t.Fatalf("GetListenerCount: %v", err)
	}
	if count != 500000 {
		t.Errorf("ListenerCount = %v, want 500000", count)
	}
}

func TestTrack_GetPlaycount(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetTrack("Iron Maiden", "The Nomad").GetPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetPlaycount: %v", err)
	}
	if count != 2000000 {
		t.Errorf("Playcount = %v, want 2000000", count)
	}
}

func TestTrack_GetUserPlaycount(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetTrack("Iron Maiden", "The Nomad").GetUserPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetUserPlaycount: %v", err)
	}
	if count != 7 {
		t.Errorf("UserPlaycount = %v, want 7", count)
	}
}

func TestTrack_GetSimilar(t *testing.T) {
	srv := serveXML(similarTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	similar, err := c.GetTrack("Iron Maiden", "The Trooper").GetSimilar(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetSimilar: %v", err)
	}
	if len(similar) != 1 {
		t.Fatalf("len(similar) = %d, want 1", len(similar))
	}
	if similar[0].Item.Title != "Run to the Hills" {
		t.Errorf("similar[0].Title = %q, want %q", similar[0].Item.Title, "Run to the Hills")
	}
	if similar[0].Match != 0.8 {
		t.Errorf("similar[0].Match = %v, want 0.8", similar[0].Match)
	}
}

func TestTrack_GetTopTags(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetTrack("Iron Maiden", "The Trooper").GetTopTags(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestTrack_GetTags(t *testing.T) {
	srv := serveXML(userTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetTrack("Iron Maiden", "The Trooper").GetTags(context.Background())
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestTrack_AddTags(t *testing.T) {
	srv := serveXML(loveOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetTrack("Iron Maiden", "The Trooper").AddTags(context.Background(), []string{"classic"})
	if err != nil {
		t.Fatalf("AddTags: %v", err)
	}
}

func TestTrack_RemoveTag(t *testing.T) {
	srv := serveXML(loveOKXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetTrack("Iron Maiden", "The Trooper").RemoveTag(context.Background(), "classic")
	if err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
}

func TestTrack_GetWikiSummary(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	summary, err := c.GetTrack("Iron Maiden", "The Nomad").GetWikiSummary(context.Background())
	if err != nil {
		t.Fatalf("GetWikiSummary: %v", err)
	}
	if summary == "" {
		t.Error("WikiSummary should not be empty")
	}
}

func TestTrack_GetWikiContent(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	content, err := c.GetTrack("Iron Maiden", "The Nomad").GetWikiContent(context.Background())
	if err != nil {
		t.Fatalf("GetWikiContent: %v", err)
	}
	if content == "" {
		t.Error("WikiContent should not be empty")
	}
}

func TestTrack_GetWikiPublishedDate(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTrack("Iron Maiden", "The Nomad").GetWikiPublishedDate(context.Background())
	if err != nil {
		t.Fatalf("GetWikiPublishedDate: %v", err)
	}
}

func TestTrack_GetWiki_NoWikiNode(t *testing.T) {
	srv := serveXML(noWikiTrackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	summary, err := c.GetTrack("Iron Maiden", "The Trooper").GetWikiSummary(context.Background())
	if err != nil {
		t.Fatalf("GetWikiSummary (no wiki): %v", err)
	}
	if summary != "" {
		t.Errorf("summary = %q, want empty when no wiki node", summary)
	}
}

func TestTrack_GetURL(t *testing.T) {
	c := NewLastFMClient("k", "s")
	url := c.GetTrack("Iron Maiden", "The Trooper").GetURL(DomainEnglish)
	if url == "" {
		t.Error("GetURL should return a non-empty string")
	}
}

func TestTrack_getWiki_DefaultSection(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.GetTrack("Iron Maiden", "The Nomad").getWiki(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("getWiki: %v", err)
	}
	if result != "" {
		t.Errorf("getWiki with unknown section = %q, want empty", result)
	}
}

func TestTrack_ErrorResponses(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, tr *Track) error
	}{
		{"GetMBID", func(ctx context.Context, tr *Track) error { _, err := tr.GetMBID(ctx); return err }},
		{"GetDuration", func(ctx context.Context, tr *Track) error { _, err := tr.GetDuration(ctx); return err }},
		{"GetListenerCount", func(ctx context.Context, tr *Track) error { _, err := tr.GetListenerCount(ctx); return err }},
		{"GetPlaycount", func(ctx context.Context, tr *Track) error { _, err := tr.GetPlaycount(ctx); return err }},
		{"GetUserPlaycount", func(ctx context.Context, tr *Track) error { _, err := tr.GetUserPlaycount(ctx); return err }},
		{"GetSimilar", func(ctx context.Context, tr *Track) error { _, err := tr.GetSimilar(ctx, 5); return err }},
		{"Love", func(ctx context.Context, tr *Track) error { return tr.Love(ctx) }},
		{"Unlove", func(ctx context.Context, tr *Track) error { return tr.Unlove(ctx) }},
		{"GetWikiSummary", func(ctx context.Context, tr *Track) error { _, err := tr.GetWikiSummary(ctx); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := serveXML(sampleErrorXML)
			defer srv.Close()
			c := newTestClient(t, srv)
			if err := tt.call(context.Background(), c.GetTrack("Iron Maiden", "The Trooper")); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestTrack_GetInfo_MissingNode(t *testing.T) {
	srv := serveXML(`<lfm status="ok"><results></results></lfm>`)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetTrack("Iron Maiden", "The Trooper").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for missing <track> node, got nil")
	}
}
