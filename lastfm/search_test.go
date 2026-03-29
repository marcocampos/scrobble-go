package lastfm

import (
	"context"
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
