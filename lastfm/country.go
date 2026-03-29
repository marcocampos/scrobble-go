package lastfm

import (
	"context"
	"fmt"
)

// Country represents a geographic country on Last.fm.
type Country struct {
	Name   string
	client *Client
}

// newCountry returns a Country bound to the given client.
func newCountry(name string, c *Client) *Country {
	return &Country{Name: name, client: c}
}

// GetName returns the country name.
func (co *Country) GetName() string { return co.Name }

// String implements fmt.Stringer.
func (co *Country) String() string { return co.Name }

// GetTopArtists returns the most popular artists in this country.
func (co *Country) GetTopArtists(ctx context.Context, limit int) ([]TopItem[*Artist], error) {
	params := map[string]string{"country": co.Name}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(co.client, "geo.getTopArtists", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Country.GetTopArtists: %w", err)
	}
	return extractTopArtists(doc, co.client), nil
}

// GetTopTracks returns the most popular tracks in this country.
func (co *Country) GetTopTracks(ctx context.Context, limit int) ([]TopItem[*Track], error) {
	params := map[string]string{"country": co.Name}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(co.client, "geo.getTopTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Country.GetTopTracks: %w", err)
	}
	return extractTopTracks(doc, co.client), nil
}

// GetURL returns the Last.fm page URL for this country.
func (co *Country) GetURL(domain int) string {
	return entityURL(co.client, urlCountry, domain, co.Name)
}
