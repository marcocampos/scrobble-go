package lastfm

import (
	"context"
	"fmt"
)

// ── Factory methods ─────────────────────────────────────────────────────────

// GetArtist returns an Artist object for the given name.
func (c *Client) GetArtist(name string) *Artist { return newArtist(name, c) }

// GetTrack returns a Track object for the given artist and title.
func (c *Client) GetTrack(artist, title string) *Track { return newTrack(artist, title, c) }

// GetAlbum returns an Album object for the given artist and title.
func (c *Client) GetAlbum(artist, title string) *Album { return newAlbum(artist, title, c) }

// GetTag returns a Tag object for the given name.
func (c *Client) GetTag(name string) *Tag { return newTag(name, c) }

// GetUser returns a User object for the given username.
func (c *Client) GetUser(username string) *User { return newUser(username, c) }

// GetCountry returns a Country object for the given country name.
func (c *Client) GetCountry(name string) *Country { return newCountry(name, c) }

// GetAuthenticatedUser returns the User for the currently authenticated account.
func (c *Client) GetAuthenticatedUser() *User {
	c.mu.RLock()
	username := c.net.Username
	c.mu.RUnlock()
	return newUser(username, c)
}

// ── MBID lookups ─────────────────────────────────────────────────────────────

// GetTrackByMBID looks up a track by its MusicBrainz ID.
func (c *Client) GetTrackByMBID(ctx context.Context, mbid string) (*Track, error) {
	doc, err := newAPIRequest(c, "track.getInfo", map[string]string{"mbid": mbid}).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetTrackByMBID: %w", err)
	}
	// name[0] = track name, name[1] = artist name in the response
	title := extract(doc, "name")
	artist := extract(doc, "name", 1)
	return newTrack(artist, title, c), nil
}

// GetArtistByMBID looks up an artist by its MusicBrainz ID.
func (c *Client) GetArtistByMBID(ctx context.Context, mbid string) (*Artist, error) {
	doc, err := newAPIRequest(c, "artist.getInfo", map[string]string{"mbid": mbid}).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetArtistByMBID: %w", err)
	}
	return newArtist(extract(doc, "name"), c), nil
}

// GetAlbumByMBID looks up an album by its MusicBrainz ID.
func (c *Client) GetAlbumByMBID(ctx context.Context, mbid string) (*Album, error) {
	doc, err := newAPIRequest(c, "album.getInfo", map[string]string{"mbid": mbid}).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetAlbumByMBID: %w", err)
	}
	return newAlbum(extract(doc, "artist"), extract(doc, "name"), c), nil
}

// ── Chart methods ─────────────────────────────────────────────────────────────

// GetTopArtists returns the globally most played artists.
func (c *Client) GetTopArtists(ctx context.Context, limit int) ([]TopItem[*Artist], error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(c, "chart.getTopArtists", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetTopArtists: %w", err)
	}
	return extractTopArtists(doc, c), nil
}

// GetTopTracks returns the globally most played tracks.
func (c *Client) GetTopTracks(ctx context.Context, limit int) ([]TopItem[*Track], error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(c, "chart.getTopTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetTopTracks: %w", err)
	}
	return extractTopTracks(doc, c), nil
}

// GetTopTags returns the most used tags globally.
func (c *Client) GetTopTags(ctx context.Context, limit int) ([]TopItem[*Tag], error) {
	doc, err := newAPIRequest(c, "tag.getTopTags", nil).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetTopTags: %w", err)
	}
	seq := extractTopTags(doc, c)
	if limit > 0 && len(seq) > limit {
		seq = seq[:limit]
	}
	return seq, nil
}

// GetGeoTopArtists returns the most popular artists in a country.
// country should be an ISO 3166-1 country name (e.g. "United Kingdom").
func (c *Client) GetGeoTopArtists(ctx context.Context, country string, limit int) ([]TopItem[*Artist], error) {
	params := map[string]string{"country": country}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(c, "geo.getTopArtists", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetGeoTopArtists: %w", err)
	}
	return extractTopArtists(doc, c), nil
}

// GetGeoTopTracks returns the most popular tracks in a country.
// location is optional (a metro area within the country); pass "" to omit it.
func (c *Client) GetGeoTopTracks(ctx context.Context, country, location string, limit int) ([]TopItem[*Track], error) {
	params := map[string]string{"country": country}
	if location != "" {
		params["location"] = location
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(c, "geo.getTopTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetGeoTopTracks: %w", err)
	}
	return extractTopTracks(doc, c), nil
}
