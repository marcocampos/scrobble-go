package lastfm

import (
	"context"
	"fmt"
)

// Track represents a music track on Last.fm.
type Track struct {
	Artist   *Artist
	Title    string
	username string
	client   *Client
}

// newTrack returns a Track bound to the given client.
func newTrack(artist, title string, c *Client) *Track {
	return &Track{
		Artist:   newArtist(artist, c),
		Title:    title,
		username: c.net.Username,
		client:   c,
	}
}

// GetTitle returns the track title.
func (t *Track) GetTitle() string { return t.Title }

// GetArtist returns the track's artist.
func (t *Track) GetArtist() *Artist { return t.Artist }

// String implements fmt.Stringer.
func (t *Track) String() string { return t.Artist.Name + " - " + t.Title }

func (t *Track) baseParams() map[string]string {
	return map[string]string{
		"artist": t.Artist.Name,
		"track":  t.Title,
	}
}

// TrackInfo holds the metadata returned by track.getInfo.
type TrackInfo struct {
	Title         string
	Artist        string
	Album         string
	MBID          string
	URL           string
	Listeners     int
	Playcount     int
	UserPlaycount int
	Duration      int // seconds
	Images        map[int]string
	TopTags       []TopItem[*Tag]
	WikiSummary   string
	WikiContent   string
	WikiPublished string
}

// GetInfo returns detailed information about the track.
func (t *Track) GetInfo(ctx context.Context) (*TrackInfo, error) {
	params := t.baseParams()
	if t.username != "" {
		params["username"] = t.username
	}
	doc, err := newAPIRequest(t.client, "track.getInfo", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Track.GetInfo: %w", err)
	}

	node := doc.find("track")
	if node == nil {
		return nil, &MalformedResponseError{
			NetworkName:     t.client.net.Name,
			UnderlyingError: fmt.Errorf("no <track> in track.getInfo response"),
		}
	}

	info := &TrackInfo{
		Title:         extract(node, "name"),
		MBID:          extract(node, "mbid"),
		URL:           extract(node, "url"),
		Listeners:     parseInt(extract(node, "listeners")),
		Playcount:     parseInt(extract(node, "playcount")),
		UserPlaycount: parseInt(extract(node, "userplaycount")),
		Duration:      parseInt(extract(node, "duration")) / 1000, // API returns ms
		Images:        extractImages(node),
	}

	if artistNode := node.find("artist"); artistNode != nil {
		info.Artist = extract(artistNode, "name")
	}
	if albumNode := node.find("album"); albumNode != nil {
		info.Album = extract(albumNode, "title")
	}

	if tagsNode := node.find("toptags"); tagsNode != nil {
		for _, tagNode := range tagsNode.findAll("tag") {
			name := extract(tagNode, "name")
			count := parseNumber(extract(tagNode, "count"))
			info.TopTags = append(info.TopTags, TopItem[*Tag]{
				Item:   newTag(name, t.client),
				Weight: count,
			})
		}
	}

	if wikiNode := node.find("wiki"); wikiNode != nil {
		info.WikiPublished = extract(wikiNode, "published")
		info.WikiSummary = extract(wikiNode, "summary")
		info.WikiContent = extract(wikiNode, "content")
	}

	return info, nil
}

// GetMBID returns the track's MusicBrainz ID.
func (t *Track) GetMBID(ctx context.Context) (string, error) {
	info, err := t.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("Track.GetMBID: %w", err)
	}
	return info.MBID, nil
}

// GetDuration returns the track duration in seconds.
func (t *Track) GetDuration(ctx context.Context) (int, error) {
	info, err := t.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Track.GetDuration: %w", err)
	}
	return info.Duration, nil
}

// GetListenerCount returns the number of Last.fm listeners.
func (t *Track) GetListenerCount(ctx context.Context) (float64, error) {
	info, err := t.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Track.GetListenerCount: %w", err)
	}
	return float64(info.Listeners), nil
}

// GetPlaycount returns the total play count on Last.fm.
func (t *Track) GetPlaycount(ctx context.Context) (float64, error) {
	info, err := t.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Track.GetPlaycount: %w", err)
	}
	return float64(info.Playcount), nil
}

// GetUserPlaycount returns the play count for the authenticated user.
func (t *Track) GetUserPlaycount(ctx context.Context) (float64, error) {
	info, err := t.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("Track.GetUserPlaycount: %w", err)
	}
	return float64(info.UserPlaycount), nil
}

// GetSimilar returns tracks similar to this one.
func (t *Track) GetSimilar(ctx context.Context, limit int) ([]SimilarItem[*Track], error) {
	params := t.baseParams()
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(t.client, "track.getSimilar", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Track.GetSimilar: %w", err)
	}

	var result []SimilarItem[*Track]
	for _, node := range doc.findAll("track") {
		title := extract(node, "name")
		artistName := ""
		if artistNode := node.find("artist"); artistNode != nil {
			artistName = extract(artistNode, "name")
		}
		match := parseNumber(extract(node, "match"))
		result = append(result, SimilarItem[*Track]{
			Item:  newTrack(artistName, title, t.client),
			Match: match,
		})
	}
	return result, nil
}

// GetTopTags returns the most frequently applied tags for this track.
func (t *Track) GetTopTags(ctx context.Context, limit int) ([]TopItem[*Tag], error) {
	return getTopTagsRequest(ctx, t.client, "track", t.baseParams(), limit)
}

// GetTags returns the tags the authenticated user has applied to this track.
func (t *Track) GetTags(ctx context.Context) ([]*Tag, error) {
	return getUserTagsRequest(ctx, t.client, "track", t.baseParams())
}

// AddTags applies one or more tags to this track for the authenticated user.
func (t *Track) AddTags(ctx context.Context, tags []string) error {
	return addTagsRequest(ctx, t.client, "track", t.baseParams(), tags)
}

// RemoveTag removes a tag from this track for the authenticated user.
func (t *Track) RemoveTag(ctx context.Context, tag string) error {
	return removeTagRequest(ctx, t.client, "track", t.baseParams(), tag)
}

// Love marks this track as loved for the authenticated user.
func (t *Track) Love(ctx context.Context) error {
	r := newAPIRequest(t.client, "track.love", t.baseParams())
	if _, err := r.execute(ctx, false); err != nil {
		return fmt.Errorf("Track.Love: %w", err)
	}
	return nil
}

// Unlove removes the love mark from this track for the authenticated user.
func (t *Track) Unlove(ctx context.Context) error {
	r := newAPIRequest(t.client, "track.unlove", t.baseParams())
	if _, err := r.execute(ctx, false); err != nil {
		return fmt.Errorf("Track.Unlove: %w", err)
	}
	return nil
}

// GetWikiSummary returns the short wiki description for the track.
func (t *Track) GetWikiSummary(ctx context.Context) (string, error) {
	return t.getWiki(ctx, "summary")
}

// GetWikiContent returns the full wiki content for the track.
func (t *Track) GetWikiContent(ctx context.Context) (string, error) {
	return t.getWiki(ctx, "content")
}

// GetWikiPublishedDate returns the published date of the track wiki.
func (t *Track) GetWikiPublishedDate(ctx context.Context) (string, error) {
	return t.getWiki(ctx, "published")
}

func (t *Track) getWiki(ctx context.Context, section string) (string, error) {
	info, err := t.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("Track.GetWiki: %w", err)
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

// GetURL returns the Last.fm page URL for this track in the given domain/language.
func (t *Track) GetURL(domain int) string {
	return entityURL(t.client, urlTrack, domain, t.Artist.Name, t.Title)
}
