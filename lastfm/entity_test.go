package lastfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── Artist ───────────────────────────────────────────────────────────────────

const artistInfoXML = `<lfm status="ok">
  <artist>
    <name>Iron Maiden</name>
    <mbid>ca891d65-d9b0-4258-89f7-e6ba29d83767</mbid>
    <url>https://www.last.fm/music/Iron+Maiden</url>
    <image size="small">https://img.last.fm/small.jpg</image>
    <image size="extralarge">https://img.last.fm/xl.jpg</image>
    <stats>
      <listeners>3456789</listeners>
      <playcount>123456789</playcount>
      <userplaycount>42</userplaycount>
    </stats>
    <tags>
      <tag><name>heavy metal</name><count>100</count></tag>
      <tag><name>metal</name><count>90</count></tag>
    </tags>
    <bio>
      <summary>Iron Maiden summary.</summary>
      <content>Iron Maiden full bio.</content>
    </bio>
  </artist>
</lfm>`

const artistSimilarXML = `<lfm status="ok">
  <similarartists artist="Iron Maiden">
    <artist>
      <name>Judas Priest</name>
      <match>0.9</match>
    </artist>
    <artist>
      <name>Black Sabbath</name>
      <match>0.85</match>
    </artist>
  </similarartists>
</lfm>`

const artistTopAlbumsXML = `<lfm status="ok">
  <topalbums artist="Iron Maiden">
    <album rank="1">
      <name>The Number of the Beast</name>
      <playcount>5000000</playcount>
      <artist><name>Iron Maiden</name></artist>
    </album>
    <album rank="2">
      <name>Powerslave</name>
      <playcount>4000000</playcount>
      <artist><name>Iron Maiden</name></artist>
    </album>
  </topalbums>
</lfm>`

func TestArtist_GetInfo(t *testing.T) {
	srv := serveXML(artistInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	a := c.GetArtist("Iron Maiden")
	info, err := a.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Name != "Iron Maiden" {
		t.Errorf("Name = %q, want %q", info.Name, "Iron Maiden")
	}
	if info.MBID != "ca891d65-d9b0-4258-89f7-e6ba29d83767" {
		t.Errorf("MBID = %q", info.MBID)
	}
	if info.Listeners != 3456789 {
		t.Errorf("Listeners = %d, want 3456789", info.Listeners)
	}
	if info.Playcount != 123456789 {
		t.Errorf("Playcount = %d, want 123456789", info.Playcount)
	}
	if info.UserPlaycount != 42 {
		t.Errorf("UserPlaycount = %d, want 42", info.UserPlaycount)
	}
	if len(info.TopTags) != 2 {
		t.Errorf("TopTags = %d, want 2", len(info.TopTags))
	}
	if info.TopTags[0].Item.Name != "heavy metal" {
		t.Errorf("TopTags[0] = %q, want %q", info.TopTags[0].Item.Name, "heavy metal")
	}
	if info.BioSummary == "" {
		t.Error("BioSummary should not be empty")
	}
	if info.Images[SizeSmall] != "https://img.last.fm/small.jpg" {
		t.Errorf("Image[SizeSmall] = %q", info.Images[SizeSmall])
	}
}

func TestArtist_GetSimilar(t *testing.T) {
	srv := serveXML(artistSimilarXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	similar, err := c.GetArtist("Iron Maiden").GetSimilar(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetSimilar: %v", err)
	}
	if len(similar) != 2 {
		t.Fatalf("len(similar) = %d, want 2", len(similar))
	}
	if similar[0].Item.Name != "Judas Priest" {
		t.Errorf("similar[0] = %q, want %q", similar[0].Item.Name, "Judas Priest")
	}
	if similar[0].Match != 0.9 {
		t.Errorf("similar[0].Match = %v, want 0.9", similar[0].Match)
	}
}

func TestArtist_GetTopAlbums(t *testing.T) {
	srv := serveXML(artistTopAlbumsXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	albums, err := c.GetArtist("Iron Maiden").GetTopAlbums(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopAlbums: %v", err)
	}
	if len(albums) != 2 {
		t.Fatalf("len(albums) = %d, want 2", len(albums))
	}
	if albums[0].Item.Title != "The Number of the Beast" {
		t.Errorf("albums[0].Title = %q", albums[0].Item.Title)
	}
	if albums[0].Weight != 5000000 {
		t.Errorf("albums[0].Weight = %v, want 5000000", albums[0].Weight)
	}
}

// ── Track ────────────────────────────────────────────────────────────────────

const trackInfoXML = `<lfm status="ok">
  <track>
    <name>The Nomad</name>
    <mbid>abc-123</mbid>
    <url>https://www.last.fm/music/Iron+Maiden/_/The+Nomad</url>
    <duration>613000</duration>
    <listeners>500000</listeners>
    <playcount>2000000</playcount>
    <userplaycount>7</userplaycount>
    <artist>
      <name>Iron Maiden</name>
    </artist>
    <album>
      <title>Dance of Death</title>
    </album>
    <toptags>
      <tag><name>heavy metal</name><count>50</count></tag>
    </toptags>
    <wiki>
      <summary>Track summary here.</summary>
      <content>Full track content here.</content>
    </wiki>
  </track>
</lfm>`

const loveOKXML = `<lfm status="ok"></lfm>`

func TestTrack_GetInfo(t *testing.T) {
	srv := serveXML(trackInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	info, err := c.GetTrack("Iron Maiden", "The Nomad").GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Title != "The Nomad" {
		t.Errorf("Title = %q, want %q", info.Title, "The Nomad")
	}
	if info.Artist != "Iron Maiden" {
		t.Errorf("Artist = %q, want %q", info.Artist, "Iron Maiden")
	}
	if info.Album != "Dance of Death" {
		t.Errorf("Album = %q, want %q", info.Album, "Dance of Death")
	}
	if info.Duration != 613 {
		t.Errorf("Duration = %d, want 613", info.Duration)
	}
	if info.UserPlaycount != 7 {
		t.Errorf("UserPlaycount = %d, want 7", info.UserPlaycount)
	}
	if info.WikiSummary == "" {
		t.Error("WikiSummary should not be empty")
	}
}

func TestTrack_Love(t *testing.T) {
	var method string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err == nil {
			method = r.Form.Get("method")
		}
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(loveOKXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetTrack("Iron Maiden", "The Nomad").Love(context.Background())
	if err != nil {
		t.Fatalf("Love: %v", err)
	}
	if method != "track.love" {
		t.Errorf("method = %q, want %q", method, "track.love")
	}
}

func TestTrack_Unlove(t *testing.T) {
	var method string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err == nil {
			method = r.Form.Get("method")
		}
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = w.Write([]byte(loveOKXML))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetTrack("Iron Maiden", "The Nomad").Unlove(context.Background())
	if err != nil {
		t.Fatalf("Unlove: %v", err)
	}
	if method != "track.unlove" {
		t.Errorf("method = %q, want %q", method, "track.unlove")
	}
}

// ── Album ────────────────────────────────────────────────────────────────────

const albumInfoXML = `<lfm status="ok">
  <album>
    <name>Dance of Death</name>
    <artist>Iron Maiden</artist>
    <mbid>xyz-456</mbid>
    <url>https://www.last.fm/music/Iron+Maiden/Dance+of+Death</url>
    <image size="large">https://img.last.fm/large.jpg</image>
    <listeners>300000</listeners>
    <playcount>1500000</playcount>
    <userplaycount>15</userplaycount>
    <tags>
      <tag><name>heavy metal</name></tag>
    </tags>
    <tracks>
      <track><name>Wildest Dreams</name></track>
      <track><name>The Nomad</name></track>
    </tracks>
    <wiki>
      <summary>Dance of Death summary.</summary>
    </wiki>
  </album>
</lfm>`

func TestAlbum_GetInfo(t *testing.T) {
	srv := serveXML(albumInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	info, err := c.GetAlbum("Iron Maiden", "Dance of Death").GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Title != "Dance of Death" {
		t.Errorf("Title = %q", info.Title)
	}
	if info.Artist != "Iron Maiden" {
		t.Errorf("Artist = %q", info.Artist)
	}
	if info.Playcount != 1500000 {
		t.Errorf("Playcount = %d, want 1500000", info.Playcount)
	}
	if info.UserPlaycount != 15 {
		t.Errorf("UserPlaycount = %d, want 15", info.UserPlaycount)
	}
	if len(info.Tracks) != 2 {
		t.Errorf("Tracks = %d, want 2", len(info.Tracks))
	}
	if info.WikiSummary == "" {
		t.Error("WikiSummary should not be empty")
	}
}

// ── User ─────────────────────────────────────────────────────────────────────

const userInfoXML = `<lfm status="ok">
  <user>
    <name>testuser</name>
    <realname>Test User</realname>
    <url>https://www.last.fm/user/testuser</url>
    <country>United Kingdom</country>
    <age>30</age>
    <gender>m</gender>
    <subscriber>0</subscriber>
    <playcount>12345</playcount>
    <playlists>2</playlists>
    <image size="medium">https://img.last.fm/avatar.jpg</image>
    <registered unixtime="1000000000">1 Sep 2001</registered>
  </user>
</lfm>`

const recentTracksXML = `<lfm status="ok">
  <recenttracks user="testuser">
    <track nowplaying="true">
      <artist>Iron Maiden</artist>
      <name>The Nomad</name>
      <album>Dance of Death</album>
    </track>
    <track>
      <artist>Metallica</artist>
      <name>Enter Sandman</name>
      <album>Metallica</album>
      <date uts="1609459200">01 Jan 2021</date>
    </track>
  </recenttracks>
</lfm>`

func TestUser_GetInfo(t *testing.T) {
	srv := serveXML(userInfoXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	info, err := c.GetUser("testuser").GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Name != "testuser" {
		t.Errorf("Name = %q", info.Name)
	}
	if info.RealName != "Test User" {
		t.Errorf("RealName = %q", info.RealName)
	}
	if info.Playcount != 12345 {
		t.Errorf("Playcount = %d, want 12345", info.Playcount)
	}
	if info.Country != "United Kingdom" {
		t.Errorf("Country = %q", info.Country)
	}
}

func TestUser_GetRecentTracks(t *testing.T) {
	srv := serveXML(recentTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	tracks, err := c.GetUser("testuser").GetRecentTracks(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("GetRecentTracks: %v", err)
	}
	// The now-playing entry should be skipped.
	if len(tracks) != 1 {
		t.Fatalf("len(tracks) = %d, want 1", len(tracks))
	}
	if tracks[0].Track.Title != "Enter Sandman" {
		t.Errorf("Track = %q, want %q", tracks[0].Track.Title, "Enter Sandman")
	}
	if tracks[0].Timestamp != "1609459200" {
		t.Errorf("Timestamp = %q, want %q", tracks[0].Timestamp, "1609459200")
	}
}

func TestUser_GetNowPlaying(t *testing.T) {
	srv := serveXML(recentTracksXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	track, err := c.GetUser("testuser").GetNowPlaying(context.Background())
	if err != nil {
		t.Fatalf("GetNowPlaying: %v", err)
	}
	if track == nil {
		t.Fatal("expected now-playing track, got nil")
	}
	if track.Title != "The Nomad" {
		t.Errorf("Title = %q, want %q", track.Title, "The Nomad")
	}
}

// ── Client factory + network ──────────────────────────────────────────────────

func TestClient_GetArtist(t *testing.T) {
	c := NewLastFMClient("key", "secret")
	a := c.GetArtist("Iron Maiden")
	if a.Name != "Iron Maiden" {
		t.Errorf("Name = %q", a.Name)
	}
}

func TestClient_GetTrack(t *testing.T) {
	c := NewLastFMClient("key", "secret")
	tr := c.GetTrack("Iron Maiden", "The Nomad")
	if tr.Title != "The Nomad" {
		t.Errorf("Title = %q", tr.Title)
	}
	if tr.Artist.Name != "Iron Maiden" {
		t.Errorf("Artist = %q", tr.Artist.Name)
	}
}

func TestClient_NewLibreFMClient(t *testing.T) {
	c := NewLibreFMClient("key", "secret")
	if c.net.Name != "Libre.fm" {
		t.Errorf("Name = %q, want %q", c.net.Name, "Libre.fm")
	}
	if c.net.WSHost != "libre.fm" {
		t.Errorf("WSHost = %q, want %q", c.net.WSHost, "libre.fm")
	}
}
