package lastfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const artistSearchXML = `<lfm status="ok">
  <results for="iron maiden">
    <opensearch:totalResults>5</opensearch:totalResults>
    <opensearch:startIndex>0</opensearch:startIndex>
    <opensearch:itemsPerPage>3</opensearch:itemsPerPage>
    <artistmatches>
      <artist>
        <name>Iron Maiden</name>
        <listeners>3456789</listeners>
      </artist>
      <artist>
        <name>Iron Maiden Tribute</name>
        <listeners>1000</listeners>
      </artist>
    </artistmatches>
  </results>
</lfm>`

const trackSearchXML = `<lfm status="ok">
  <results for="the nomad">
    <opensearch:totalResults>3</opensearch:totalResults>
    <trackmatches>
      <track>
        <name>The Nomad</name>
        <artist>Iron Maiden</artist>
        <listeners>500000</listeners>
      </track>
      <track>
        <name>Nomad</name>
        <artist>Khaled</artist>
        <listeners>100000</listeners>
      </track>
    </trackmatches>
  </results>
</lfm>`

const albumSearchXML = `<lfm status="ok">
  <results for="dance of death">
    <opensearch:totalResults>2</opensearch:totalResults>
    <albummatches>
      <album>
        <name>Dance of Death</name>
        <artist>Iron Maiden</artist>
      </album>
    </albummatches>
  </results>
</lfm>`

func TestArtistSearch_GetNextPage(t *testing.T) {
	srv := serveXML(artistSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	search := c.SearchForArtist("iron maiden")
	results, err := search.GetNextPage(context.Background())
	if err != nil {
		t.Fatalf("GetNextPage: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Name != "Iron Maiden" {
		t.Errorf("results[0] = %q, want %q", results[0].Name, "Iron Maiden")
	}
}

func TestArtistSearch_PageAdvances(t *testing.T) {
	srv := serveXML(artistSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	s := c.SearchForArtist("iron")
	if s.page != 0 {
		t.Errorf("initial page = %d, want 0", s.page)
	}
	_, _ = s.GetNextPage(context.Background())
	if s.page != 1 {
		t.Errorf("after first GetNextPage, page = %d, want 1", s.page)
	}
	_, _ = s.GetNextPage(context.Background())
	if s.page != 2 {
		t.Errorf("after second GetNextPage, page = %d, want 2", s.page)
	}
}

func TestTrackSearch_GetNextPage(t *testing.T) {
	srv := serveXML(trackSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	results, err := c.SearchForTrack("Iron Maiden", "The Nomad").GetNextPage(context.Background())
	if err != nil {
		t.Fatalf("GetNextPage: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Title != "The Nomad" {
		t.Errorf("results[0].Title = %q", results[0].Title)
	}
	if results[0].Artist.Name != "Iron Maiden" {
		t.Errorf("results[0].Artist = %q", results[0].Artist.Name)
	}
}

func TestAlbumSearch_GetNextPage(t *testing.T) {
	srv := serveXML(albumSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	results, err := c.SearchForAlbum("dance of death").GetNextPage(context.Background())
	if err != nil {
		t.Fatalf("GetNextPage: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Title != "Dance of Death" {
		t.Errorf("Title = %q", results[0].Title)
	}
}

// ── All() iterator tests ──────────────────────────────────────────────────────

const emptyArtistSearchXML = `<lfm status="ok">
  <results for="nothing">
    <opensearch:totalResults>0</opensearch:totalResults>
    <artistmatches></artistmatches>
  </results>
</lfm>`

const emptyAlbumSearchXML = `<lfm status="ok">
  <results for="nothing">
    <opensearch:totalResults>0</opensearch:totalResults>
    <albummatches></albummatches>
  </results>
</lfm>`

const emptyTrackSearchXML = `<lfm status="ok">
  <results for="nothing">
    <opensearch:totalResults>0</opensearch:totalResults>
    <trackmatches></trackmatches>
  </results>
</lfm>`

// servePages returns a TLS server that responds with pages[0] on the first
// request, pages[1] on the second, and so on. The last entry is repeated for
// any additional requests. Panics if called with no pages.
func servePages(pages ...string) *httptest.Server {
	if len(pages) == 0 {
		panic("servePages: at least one page response must be provided")
	}
	calls := 0
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		idx := calls
		if idx >= len(pages) {
			idx = len(pages) - 1
		}
		calls++
		_, _ = w.Write([]byte(pages[idx]))
	}))
}

func TestArtistSearch_All_YieldsAllResults(t *testing.T) {
	// Page 1 returns 2 artists; page 2 returns empty → iterator stops.
	srv := servePages(artistSearchXML, emptyArtistSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	var names []string
	for artist, err := range c.SearchForArtist("iron maiden").All(context.Background()) {
		if err != nil {
			t.Fatalf("All: unexpected error: %v", err)
		}
		names = append(names, artist.Name)
	}
	if len(names) != 2 {
		t.Errorf("got %d results, want 2", len(names))
	}
}

func TestAlbumSearch_All_YieldsAllResults(t *testing.T) {
	srv := servePages(albumSearchXML, emptyAlbumSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	var titles []string
	for album, err := range c.SearchForAlbum("dance of death").All(context.Background()) {
		if err != nil {
			t.Fatalf("All: unexpected error: %v", err)
		}
		titles = append(titles, album.Title)
	}
	if len(titles) != 1 {
		t.Errorf("got %d results, want 1", len(titles))
	}
}

func TestTrackSearch_All_YieldsAllResults(t *testing.T) {
	srv := servePages(trackSearchXML, emptyTrackSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	var titles []string
	for track, err := range c.SearchForTrack("", "nomad").All(context.Background()) {
		if err != nil {
			t.Fatalf("All: unexpected error: %v", err)
		}
		titles = append(titles, track.Title)
	}
	if len(titles) != 2 {
		t.Errorf("got %d results, want 2", len(titles))
	}
}

func TestAlbumSearch_All_StopsOnEarlyReturn(t *testing.T) {
	srv := servePages(albumSearchXML, emptyAlbumSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count := 0
	for _, err := range c.SearchForAlbum("dance of death").All(context.Background()) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		break // stop after first result
	}
	if count != 1 {
		t.Errorf("expected early exit after 1 result, got %d", count)
	}
}

func TestAlbumSearch_All_PropagatesError(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	for _, err := range c.SearchForAlbum("dance of death").All(context.Background()) {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}
	t.Fatal("expected at least one iteration with an error")
}

func TestTrackSearch_All_StopsOnEarlyReturn(t *testing.T) {
	srv := servePages(trackSearchXML, emptyTrackSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count := 0
	for _, err := range c.SearchForTrack("", "nomad").All(context.Background()) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		break // stop after first result
	}
	if count != 1 {
		t.Errorf("expected early exit after 1 result, got %d", count)
	}
}

func TestTrackSearch_All_PropagatesError(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	for _, err := range c.SearchForTrack("", "nomad").All(context.Background()) {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}
	t.Fatal("expected at least one iteration with an error")
}

func TestArtistSearch_All_StopsOnEarlyReturn(t *testing.T) {
	srv := servePages(artistSearchXML, emptyArtistSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count := 0
	for _, err := range c.SearchForArtist("iron maiden").All(context.Background()) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		break // stop after first result
	}
	if count != 1 {
		t.Errorf("expected early exit after 1 result, got %d", count)
	}
}

func TestArtistSearch_All_PropagatesError(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	for _, err := range c.SearchForArtist("iron maiden").All(context.Background()) {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}
	t.Fatal("expected at least one iteration with an error")
}

func TestArtistSearch_GetTotalResultCount(t *testing.T) {
	srv := serveXML(artistSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.SearchForArtist("iron maiden").GetTotalResultCount(context.Background())
	if err != nil {
		t.Fatalf("GetTotalResultCount: %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestAlbumSearch_GetTotalResultCount(t *testing.T) {
	srv := serveXML(albumSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.SearchForAlbum("dance of death").GetTotalResultCount(context.Background())
	if err != nil {
		t.Fatalf("GetTotalResultCount: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestTrackSearch_GetTotalResultCount(t *testing.T) {
	srv := serveXML(trackSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	count, err := c.SearchForTrack("Iron Maiden", "The Nomad").GetTotalResultCount(context.Background())
	if err != nil {
		t.Fatalf("GetTotalResultCount: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestArtistSearch_GetTotalResultCount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.SearchForArtist("iron maiden").GetTotalResultCount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAlbumSearch_GetTotalResultCount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.SearchForAlbum("dance of death").GetTotalResultCount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTrackSearch_GetTotalResultCount_Error(t *testing.T) {
	srv := serveXML(sampleErrorXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.SearchForTrack("Iron Maiden", "The Nomad").GetTotalResultCount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTrackSearch_EmptyArtist(t *testing.T) {
	srv := serveXML(trackSearchXML)
	defer srv.Close()

	c := newTestClient(t, srv)
	// Empty artist should still work (searches all artists).
	results, err := c.SearchForTrack("", "The Nomad").GetNextPage(context.Background())
	if err != nil {
		t.Fatalf("GetNextPage: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results, got none")
	}
}
