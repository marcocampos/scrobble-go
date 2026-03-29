package lastfm

import (
	"context"
	"fmt"
)

// Tag represents a Last.fm tag (genre, mood, etc.).
type Tag struct {
	Name   string
	client *Client
}

// newTag returns a Tag bound to the given client.
func newTag(name string, c *Client) *Tag {
	return &Tag{Name: name, client: c}
}

// GetName returns the tag name.
func (t *Tag) GetName() string { return t.Name }

// String implements fmt.Stringer.
func (t *Tag) String() string { return t.Name }

// TagInfo holds metadata about a tag.
type TagInfo struct {
	Name        string
	Reach       int
	Taggings    int
	URL         string
	WikiSummary string
	WikiContent string
}

// GetInfo returns metadata about the tag.
func (t *Tag) GetInfo(ctx context.Context) (*TagInfo, error) {
	doc, err := newAPIRequest(t.client, "tag.getInfo",
		map[string]string{"tag": t.Name},
	).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Tag.GetInfo: %w", err)
	}

	tagNode := doc.find("tag")
	if tagNode == nil {
		return nil, &MalformedResponseError{
			NetworkName:     t.client.net.Name,
			UnderlyingError: fmt.Errorf("no <tag> in tag.getInfo response"),
		}
	}

	info := &TagInfo{
		Name:     extract(tagNode, "name"),
		Reach:    parseInt(extract(tagNode, "reach")),
		Taggings: parseInt(extract(tagNode, "taggings")),
		URL:      extract(tagNode, "url"),
	}
	if wiki := tagNode.find("wiki"); wiki != nil {
		info.WikiSummary = extract(wiki, "summary")
		info.WikiContent = extract(wiki, "content")
	}
	return info, nil
}

// GetTopArtists returns the most-used artists for this tag.
func (t *Tag) GetTopArtists(ctx context.Context, limit int) ([]TopItem[*Artist], error) {
	params := map[string]string{"tag": t.Name}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(t.client, "tag.getTopArtists", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Tag.GetTopArtists: %w", err)
	}
	return extractTopArtists(doc, t.client), nil
}

// GetTopTracks returns the most-used tracks for this tag.
func (t *Tag) GetTopTracks(ctx context.Context, limit int) ([]TopItem[*Track], error) {
	params := map[string]string{"tag": t.Name}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(t.client, "tag.getTopTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Tag.GetTopTracks: %w", err)
	}
	return extractTopTracks(doc, t.client), nil
}

// GetTopAlbums returns the most-used albums for this tag.
func (t *Tag) GetTopAlbums(ctx context.Context, limit int) ([]TopItem[*Album], error) {
	params := map[string]string{"tag": t.Name}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(t.client, "tag.getTopAlbums", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Tag.GetTopAlbums: %w", err)
	}
	return extractTopAlbums(doc, t.client), nil
}
