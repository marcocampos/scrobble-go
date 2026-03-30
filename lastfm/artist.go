package lastfm

import (
	"context"
	"fmt"
)

// Artist represents a musical artist on Last.fm.
type Artist struct {
	Name   string
	client *Client
}

// newArtist returns an Artist bound to the given client.
func newArtist(name string, c *Client) *Artist {
	return &Artist{Name: name, client: c}
}

// GetName returns the artist's name.
func (a *Artist) GetName() string { return a.Name }

// String implements fmt.Stringer.
func (a *Artist) String() string { return a.Name }

func (a *Artist) baseParams() map[string]string {
	return map[string]string{"artist": a.Name}
}

// ArtistInfo holds the metadata returned by artist.getInfo.
type ArtistInfo struct {
	Name          string
	MBID          string
	URL           string
	Listeners     int64
	Playcount     int64
	UserPlaycount int64
	Images        map[ImageSize]string
	TopTags       []TopItem[*Tag]
	BioSummary    string
	BioContent    string
	BioPublished  string
}

// GetInfo returns detailed information about the artist.
// If the client has an authenticated username, the response also includes
// that user's personal play count (UserPlaycount).
func (a *Artist) GetInfo(ctx context.Context) (*ArtistInfo, error) {
	params := a.baseParams()
	a.client.mu.RLock()
	username := a.client.net.Username
	a.client.mu.RUnlock()
	if username != "" {
		params["username"] = username
	}
	doc, err := newAPIRequest(a.client, "artist.getInfo", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Artist.GetInfo: %w", err)
	}

	node := doc.find("artist")
	if node == nil {
		return nil, &MalformedResponseError{
			NetworkName:     a.client.net.Name,
			UnderlyingError: fmt.Errorf("no <artist> in artist.getInfo response"),
		}
	}

	info := &ArtistInfo{
		Name:          extract(node, "name"),
		MBID:          extract(node, "mbid"),
		URL:           extract(node, "url"),
		Listeners:     parseInt64(extract(node, "listeners")),
		Playcount:     parseInt64(extract(node, "playcount")),
		UserPlaycount: parseInt64(extract(node, "userplaycount")),
		Images:        extractImages(node),
	}

	if tagsNode := node.find("tags"); tagsNode != nil {
		for _, tagNode := range tagsNode.findAll("tag") {
			name := extract(tagNode, "name")
			count := parseNumber(extract(tagNode, "count"))
			info.TopTags = append(info.TopTags, TopItem[*Tag]{
				Item:   newTag(name, a.client),
				Weight: count,
			})
		}
	}

	if bioNode := node.find("bio"); bioNode != nil {
		info.BioPublished = extract(bioNode, "published")
		info.BioSummary = extract(bioNode, "summary")
		info.BioContent = extract(bioNode, "content")
	}

	return info, nil
}

// GetMBID returns the artist's MusicBrainz ID.
func (a *Artist) GetMBID(ctx context.Context) (string, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("Artist.GetMBID: %w", err)
	}
	return info.MBID, nil
}

// GetListenerCount returns the number of Last.fm listeners.
func (a *Artist) GetListenerCount(ctx context.Context) (int64, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Artist.GetListenerCount: %w", err)
	}
	return info.Listeners, nil
}

// GetPlaycount returns the total play count on Last.fm.
func (a *Artist) GetPlaycount(ctx context.Context) (int64, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Artist.GetPlaycount: %w", err)
	}
	return info.Playcount, nil
}

// GetUserPlaycount returns the play count for the authenticated user.
func (a *Artist) GetUserPlaycount(ctx context.Context) (int64, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Artist.GetUserPlaycount: %w", err)
	}
	return info.UserPlaycount, nil
}

// GetSimilar returns artists similar to this one.
// Pass limit ≤ 0 for the default (100).
func (a *Artist) GetSimilar(ctx context.Context, limit int) ([]SimilarItem[*Artist], error) {
	params := a.baseParams()
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(a.client, "artist.getSimilar", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Artist.GetSimilar: %w", err)
	}

	var result []SimilarItem[*Artist]
	for _, node := range doc.findAll("artist") {
		name := extract(node, "name")
		match := parseNumber(extract(node, "match"))
		result = append(result, SimilarItem[*Artist]{
			Item:  newArtist(name, a.client),
			Match: match,
		})
	}
	return result, nil
}

// GetTopTags returns the most frequently applied tags for this artist.
func (a *Artist) GetTopTags(ctx context.Context, limit int) ([]TopItem[*Tag], error) {
	return getTopTagsRequest(ctx, a.client, "artist", a.baseParams(), limit)
}

// GetTags returns the tags the authenticated user has applied to this artist.
func (a *Artist) GetTags(ctx context.Context) ([]*Tag, error) {
	return getUserTagsRequest(ctx, a.client, "artist", a.baseParams())
}

// AddTags applies one or more tags to this artist for the authenticated user.
func (a *Artist) AddTags(ctx context.Context, tags []string) error {
	return addTagsRequest(ctx, a.client, "artist", a.baseParams(), tags)
}

// RemoveTag removes a tag from this artist for the authenticated user.
func (a *Artist) RemoveTag(ctx context.Context, tag string) error {
	return removeTagRequest(ctx, a.client, "artist", a.baseParams(), tag)
}

// GetTopAlbums returns the most played albums for this artist.
func (a *Artist) GetTopAlbums(ctx context.Context, limit int) ([]TopItem[*Album], error) {
	params := a.baseParams()
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(a.client, "artist.getTopAlbums", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Artist.GetTopAlbums: %w", err)
	}
	return extractTopAlbums(doc, a.client), nil
}

// GetTopTracks returns the most played tracks for this artist.
func (a *Artist) GetTopTracks(ctx context.Context, limit int) ([]TopItem[*Track], error) {
	params := a.baseParams()
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(a.client, "artist.getTopTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Artist.GetTopTracks: %w", err)
	}
	return extractTopTracks(doc, a.client), nil
}

// GetURL returns the Last.fm page URL for this artist in the given domain/language.
func (a *Artist) GetURL(domain Domain) string {
	return entityURL(a.client, urlArtist, domain, a.Name)
}
