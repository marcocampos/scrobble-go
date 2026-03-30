package lastfm

import (
	"context"
	"fmt"
)

// Album represents a music album on Last.fm.
type Album struct {
	Artist   *Artist
	Title    string
	username string
	client   *Client
}

// newAlbum returns an Album bound to the given client.
func newAlbum(artist, title string, c *Client) *Album {
	return &Album{
		Artist:   newArtist(artist, c),
		Title:    title,
		username: c.net.Username,
		client:   c,
	}
}

// GetTitle returns the album title.
func (a *Album) GetTitle() string { return a.Title }

// GetArtist returns the album's artist.
func (a *Album) GetArtist() *Artist { return a.Artist }

// String implements fmt.Stringer.
func (a *Album) String() string { return a.Artist.Name + " - " + a.Title }

func (a *Album) baseParams() map[string]string {
	return map[string]string{
		"artist": a.Artist.Name,
		"album":  a.Title,
	}
}

// AlbumInfo holds the metadata returned by album.getInfo.
type AlbumInfo struct {
	Title         string
	Artist        string
	MBID          string
	URL           string
	Listeners     int
	Playcount     int
	UserPlaycount int
	Images        map[int]string
	TopTags       []TopItem[*Tag]
	Tracks        []*Track
	WikiSummary   string
	WikiContent   string
	WikiPublished string
}

// GetInfo returns detailed information about the album.
// If the client has an authenticated username, the response also includes
// that user's personal play count (UserPlaycount).
func (a *Album) GetInfo(ctx context.Context) (*AlbumInfo, error) {
	params := a.baseParams()
	if username := a.client.net.Username; username != "" {
		params["username"] = username
	}
	doc, err := newAPIRequest(a.client, "album.getInfo", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Album.GetInfo: %w", err)
	}

	node := doc.find("album")
	if node == nil {
		return nil, &MalformedResponseError{
			NetworkName:     a.client.net.Name,
			UnderlyingError: fmt.Errorf("no <album> in album.getInfo response"),
		}
	}

	info := &AlbumInfo{
		Title:         extract(node, "name"),
		Artist:        extract(node, "artist"),
		MBID:          extract(node, "mbid"),
		URL:           extract(node, "url"),
		Listeners:     parseInt(extract(node, "listeners")),
		Playcount:     parseInt(extract(node, "playcount")),
		UserPlaycount: parseInt(extract(node, "userplaycount")),
		Images:        extractImages(node),
	}

	if tagsNode := node.find("tags"); tagsNode != nil {
		for _, tagNode := range tagsNode.findAll("tag") {
			name := extract(tagNode, "name")
			info.TopTags = append(info.TopTags, TopItem[*Tag]{
				Item:   newTag(name, a.client),
				Weight: 0,
			})
		}
	}

	if tracksNode := node.find("tracks"); tracksNode != nil {
		for _, trackNode := range tracksNode.findAll("track") {
			title := extract(trackNode, "name")
			info.Tracks = append(info.Tracks, newTrack(a.Artist.Name, title, a.client))
		}
	}

	if wikiNode := node.find("wiki"); wikiNode != nil {
		info.WikiPublished = extract(wikiNode, "published")
		info.WikiSummary = extract(wikiNode, "summary")
		info.WikiContent = extract(wikiNode, "content")
	}

	return info, nil
}

// GetMBID returns the album's MusicBrainz ID.
func (a *Album) GetMBID(ctx context.Context) (string, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("Album.GetMBID: %w", err)
	}
	return info.MBID, nil
}

// GetListenerCount returns the number of Last.fm listeners.
func (a *Album) GetListenerCount(ctx context.Context) (float64, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Album.GetListenerCount: %w", err)
	}
	return float64(info.Listeners), nil
}

// GetPlaycount returns the total play count on Last.fm.
func (a *Album) GetPlaycount(ctx context.Context) (float64, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Album.GetPlaycount: %w", err)
	}
	return float64(info.Playcount), nil
}

// GetUserPlaycount returns the play count for the authenticated user.
func (a *Album) GetUserPlaycount(ctx context.Context) (float64, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Album.GetUserPlaycount: %w", err)
	}
	return float64(info.UserPlaycount), nil
}

// GetCoverImage returns the URL of the album cover at the given size.
// Use one of the Size* constants.
func (a *Album) GetCoverImage(ctx context.Context, size int) (string, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("Album.GetCoverImage: %w", err)
	}
	return info.Images[size], nil
}

// GetTracks returns the tracks listed in this album.
func (a *Album) GetTracks(ctx context.Context) ([]*Track, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("Album.GetTracks: %w", err)
	}
	return info.Tracks, nil
}

// GetTopTags returns the most frequently applied tags for this album.
func (a *Album) GetTopTags(ctx context.Context, limit int) ([]TopItem[*Tag], error) {
	return getTopTagsRequest(ctx, a.client, "album", a.baseParams(), limit)
}

// GetTags returns the tags the authenticated user has applied to this album.
func (a *Album) GetTags(ctx context.Context) ([]*Tag, error) {
	return getUserTagsRequest(ctx, a.client, "album", a.baseParams())
}

// AddTags applies one or more tags to this album for the authenticated user.
func (a *Album) AddTags(ctx context.Context, tags []string) error {
	return addTagsRequest(ctx, a.client, "album", a.baseParams(), tags)
}

// RemoveTag removes a tag from this album for the authenticated user.
func (a *Album) RemoveTag(ctx context.Context, tag string) error {
	return removeTagRequest(ctx, a.client, "album", a.baseParams(), tag)
}

// GetWikiSummary returns the short wiki description for the album.
func (a *Album) GetWikiSummary(ctx context.Context) (string, error) {
	return a.getWiki(ctx, "summary")
}

// GetWikiContent returns the full wiki content for the album.
func (a *Album) GetWikiContent(ctx context.Context) (string, error) {
	return a.getWiki(ctx, "content")
}

// GetWikiPublishedDate returns the published date of the album wiki.
func (a *Album) GetWikiPublishedDate(ctx context.Context) (string, error) {
	return a.getWiki(ctx, "published")
}

func (a *Album) getWiki(ctx context.Context, section string) (string, error) {
	info, err := a.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("Album.GetWiki: %w", err)
	}
	switch section {
	case "summary":
		return info.WikiSummary, nil
	case "content":
		return info.WikiContent, nil
	case "published":
		return info.WikiPublished, nil
	default:
		return "", nil
	}
}

// GetURL returns the Last.fm page URL for this album in the given domain/language.
func (a *Album) GetURL(domain int) string {
	return entityURL(a.client, urlAlbum, domain, a.Artist.Name, a.Title)
}
