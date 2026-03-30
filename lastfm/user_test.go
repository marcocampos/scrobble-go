package lastfm

import (
	"context"
	"testing"
)

func TestUser_GetName(t *testing.T) {
	c := NewLastFMClient("k", "s")
	u := c.GetUser("testuser")
	if u.GetName() != "testuser" {
		t.Errorf("GetName() = %q", u.GetName())
	}
}

func TestUser_String(t *testing.T) {
	c := NewLastFMClient("k", "s")
	u := c.GetUser("testuser")
	if u.String() != "testuser" {
		t.Errorf("String() = %q", u.String())
	}
}

func TestUser_GetLovedTracks(t *testing.T) {
	srv := serveXML(lovedTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetUser("testuser").GetLovedTracks(context.Background(), 10, 1)
	if err != nil {
		t.Fatalf("GetLovedTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
	if tracks[0].Track.Title != "The Trooper" {
		t.Errorf("Track.Title = %q, want %q", tracks[0].Track.Title, "The Trooper")
	}
	if tracks[0].Timestamp != "1609459200" {
		t.Errorf("Timestamp = %q, want %q", tracks[0].Timestamp, "1609459200")
	}
}

func TestUser_GetTopArtists(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artists, err := c.GetUser("testuser").GetTopArtists(context.Background(), PeriodOverall, 5)
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

func TestUser_GetTopAlbums(t *testing.T) {
	srv := serveXML(topAlbumsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	albums, err := c.GetUser("testuser").GetTopAlbums(context.Background(), Period7Days, 5)
	if err != nil {
		t.Fatalf("GetTopAlbums: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("len(albums) = %d, want 1", len(albums))
	}
	if albums[0].Item.Title != "Piece of Mind" {
		t.Errorf("albums[0].Title = %q", albums[0].Item.Title)
	}
}

func TestUser_GetTopTracks(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetUser("testuser").GetTopTracks(context.Background(), Period1Month, 5)
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
}

func TestUser_GetTopTags(t *testing.T) {
	srv := serveXML(topTagsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tags, err := c.GetUser("testuser").GetTopTags(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestUser_GetPlaycount(t *testing.T) {
	srv := serveXML(userInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.GetUser("testuser").GetPlaycount(context.Background())
	if err != nil {
		t.Fatalf("GetPlaycount: %v", err)
	}
	if count != 12345 {
		t.Errorf("Playcount = %d, want 12345", count)
	}
}

func TestUser_GetWeeklyChartDates(t *testing.T) {
	srv := serveXML(weeklyChartListXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	dates, err := c.GetUser("testuser").GetWeeklyChartDates(context.Background())
	if err != nil {
		t.Fatalf("GetWeeklyChartDates: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("len(dates) = %d, want 2", len(dates))
	}
	if dates[0].From != "1609459200" {
		t.Errorf("dates[0].From = %q", dates[0].From)
	}
}

func TestUser_GetWeeklyArtistCharts(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	artists, err := c.GetUser("testuser").GetWeeklyArtistCharts(context.Background(), "1609459200", "1610064000")
	if err != nil {
		t.Fatalf("GetWeeklyArtistCharts: %v", err)
	}
	if len(artists) != 2 {
		t.Fatalf("len(artists) = %d, want 2", len(artists))
	}
}

func TestUser_GetWeeklyArtistCharts_Empty(t *testing.T) {
	srv := serveXML(topArtistsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	// Empty from/to should use most recent chart.
	_, err := c.GetUser("testuser").GetWeeklyArtistCharts(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetWeeklyArtistCharts (empty range): %v", err)
	}
}

func TestUser_GetWeeklyTrackCharts(t *testing.T) {
	srv := serveXML(topTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetUser("testuser").GetWeeklyTrackCharts(context.Background(), "1609459200", "1610064000")
	if err != nil {
		t.Fatalf("GetWeeklyTrackCharts: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
}

func TestUser_GetWeeklyAlbumCharts(t *testing.T) {
	srv := serveXML(topAlbumsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	albums, err := c.GetUser("testuser").GetWeeklyAlbumCharts(context.Background(), "1609459200", "1610064000")
	if err != nil {
		t.Fatalf("GetWeeklyAlbumCharts: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("len(albums) = %d, want 1", len(albums))
	}
}

func TestUser_GetURL(t *testing.T) {
	c := NewLastFMClient("k", "s")
	url := c.GetUser("testuser").GetURL(DomainEnglish)
	if url == "" {
		t.Error("GetURL should return a non-empty string")
	}
}

func TestUser_GetNowPlaying_None(t *testing.T) {
	// recentTracksXML has a now-playing track; use a response without one.
	srv := serveXML(`<lfm status="ok">
	  <recenttracks user="testuser">
	    <track>
	      <artist>Metallica</artist>
	      <name>Enter Sandman</name>
	      <date uts="1609459200">01 Jan 2021</date>
	    </track>
	  </recenttracks>
	</lfm>`)
	defer srv.Close()

	c := newTestClient(t, srv)
	track, err := c.GetUser("testuser").GetNowPlaying(context.Background())
	if err != nil {
		t.Fatalf("GetNowPlaying: %v", err)
	}
	if track != nil {
		t.Errorf("expected nil (not playing), got %v", track)
	}
}

func TestUser_GetInfo_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetRecentTracks_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetRecentTracks(context.Background(), 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetNowPlaying_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetNowPlaying(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetLovedTracks_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetLovedTracks(context.Background(), 10, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetTopArtists_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetTopArtists(context.Background(), PeriodOverall, 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetTopAlbums_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetTopAlbums(context.Background(), Period7Days, 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetTopTracks_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetTopTracks(context.Background(), Period1Month, 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetPlaycount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetPlaycount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetWeeklyChartDates_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetWeeklyChartDates(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetWeeklyArtistCharts_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetWeeklyArtistCharts(context.Background(), "1609459200", "1610064000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetWeeklyTrackCharts_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetWeeklyTrackCharts(context.Background(), "1609459200", "1610064000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetWeeklyAlbumCharts_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetWeeklyAlbumCharts(context.Background(), "1609459200", "1610064000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUser_GetInfo_MissingNode(t *testing.T) {
	srv := serveXML(`<lfm status="ok"><results></results></lfm>`)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetUser("testuser").GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for missing <user> node, got nil")
	}
}
