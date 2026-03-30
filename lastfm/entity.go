package lastfm

import (
	"context"
	"fmt"
)

// extractImages builds a map of size constant → image URL from an XML node
// containing <image size="...">URL</image> elements.
func extractImages(n *xmlNode) map[ImageSize]string {
	sizeMap := map[string]ImageSize{
		"small":      SizeSmall,
		"medium":     SizeMedium,
		"large":      SizeLarge,
		"extralarge": SizeExtraLarge,
		"mega":       SizeMega,
	}
	images := make(map[ImageSize]string)
	for _, imgNode := range n.findAll("image") {
		sizeStr := imgNode.attr("size")
		if idx, ok := sizeMap[sizeStr]; ok {
			images[idx] = imgNode.text()
		}
	}
	return images
}

// extractTopArtists extracts a []TopItem[*Artist] from an XML response that
// contains <artist><name>…</name><playcount>…</playcount></artist> elements.
func extractTopArtists(doc *xmlNode, c *Client) []TopItem[*Artist] {
	var result []TopItem[*Artist]
	for _, node := range doc.findAll("artist") {
		name := extract(node, "name")
		weight := parseNumber(extract(node, "playcount"))
		result = append(result, TopItem[*Artist]{
			Item:   newArtist(name, c),
			Weight: weight,
		})
	}
	return result
}

// extractTopTracks extracts a []TopItem[*Track] from an XML response.
func extractTopTracks(doc *xmlNode, c *Client) []TopItem[*Track] {
	var result []TopItem[*Track]
	for _, node := range doc.findAll("track") {
		title := extract(node, "name")
		// In many responses the artist name is in a nested <artist><name>
		artistName := ""
		if artistNode := node.find("artist"); artistNode != nil {
			artistName = extract(artistNode, "name")
		}
		if artistName == "" {
			artistName = extract(node, "name", 1)
		}
		weight := parseNumber(extract(node, "playcount"))
		result = append(result, TopItem[*Track]{
			Item:   newTrack(artistName, title, c),
			Weight: weight,
		})
	}
	return result
}

// extractTopAlbums extracts a []TopItem[*Album] from an XML response.
func extractTopAlbums(doc *xmlNode, c *Client) []TopItem[*Album] {
	var result []TopItem[*Album]
	for _, node := range doc.findAll("album") {
		title := extract(node, "name")
		artistName := ""
		if artistNode := node.find("artist"); artistNode != nil {
			artistName = extract(artistNode, "name")
		}
		if artistName == "" {
			artistName = extract(node, "name", 1)
		}
		weight := parseNumber(extract(node, "playcount"))
		result = append(result, TopItem[*Album]{
			Item:   newAlbum(artistName, title, c),
			Weight: weight,
		})
	}
	return result
}

// extractTopTags extracts a []TopItem[*Tag] from an XML response.
func extractTopTags(doc *xmlNode, c *Client) []TopItem[*Tag] {
	var result []TopItem[*Tag]
	for _, node := range doc.findAll("tag") {
		name := extract(node, "name")
		weight := parseNumber(extract(node, "count"))
		result = append(result, TopItem[*Tag]{
			Item:   newTag(name, c),
			Weight: weight,
		})
	}
	return result
}

// entityURL returns the Last.fm URL for an entity given the client's URL
// template for that entity type.
func entityURL(c *Client, entityType entityKind, domain Domain, args ...any) string {
	tmpl, ok := c.net.URLs[urlKey(entityType)]
	if !ok {
		return ""
	}
	domainHost, ok := c.net.DomainNames[domain]
	if !ok {
		domainHost = c.net.DomainNames[DomainEnglish]
	}
	return fmt.Sprintf("https://"+domainHost+"/"+tmpl, args...)
}

// entityKind identifies an entity type for URL generation.
type entityKind int

// URL type constants for entityURL.
const (
	urlAlbum   entityKind = 0
	urlArtist  entityKind = 1
	urlCountry entityKind = 2
	urlTag     entityKind = 3
	urlTrack   entityKind = 4
	urlUser    entityKind = 5
)

func urlKey(entityType entityKind) string {
	switch entityType {
	case urlAlbum:
		return "album"
	case urlArtist:
		return "artist"
	case urlCountry:
		return "country"
	case urlTag:
		return "tag"
	case urlTrack:
		return "track"
	case urlUser:
		return "user"
	}
	return ""
}

// addTagsRequest sends a <ws_prefix>.addTags request.
func addTagsRequest(ctx context.Context, c *Client, wsPrefix string, baseParams map[string]string, tags []string) error {
	for _, tag := range tags {
		params := copyMap(baseParams)
		params["tags"] = tag
		r := newAPIRequest(c, wsPrefix+".addTags", params)
		if _, err := r.execute(ctx, false); err != nil {
			return fmt.Errorf("AddTags: %w", err)
		}
	}
	return nil
}

// removeTagRequest sends a <ws_prefix>.removeTag request.
func removeTagRequest(ctx context.Context, c *Client, wsPrefix string, baseParams map[string]string, tag string) error {
	params := copyMap(baseParams)
	params["tag"] = tag
	r := newAPIRequest(c, wsPrefix+".removeTag", params)
	if _, err := r.execute(ctx, false); err != nil {
		return fmt.Errorf("RemoveTag: %w", err)
	}
	return nil
}

// getUserTagsRequest sends a <ws_prefix>.getTags request.
func getUserTagsRequest(ctx context.Context, c *Client, wsPrefix string, baseParams map[string]string) ([]*Tag, error) {
	r := newAPIRequest(c, wsPrefix+".getTags", baseParams)
	doc, err := r.execute(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("GetTags: %w", err)
	}
	var tags []*Tag
	for _, name := range extractAll(doc, "name") {
		tags = append(tags, newTag(name, c))
	}
	return tags, nil
}

// getTopTagsRequest sends a <ws_prefix>.getTopTags request.
func getTopTagsRequest(ctx context.Context, c *Client, wsPrefix string, baseParams map[string]string, limit int) ([]TopItem[*Tag], error) {
	r := newAPIRequest(c, wsPrefix+".getTopTags", baseParams)
	doc, err := r.execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("GetTopTags: %w", err)
	}
	seq := extractTopTags(doc, c)
	if limit > 0 && len(seq) > limit {
		seq = seq[:limit]
	}
	return seq, nil
}

// copyMap returns a shallow copy of m.
func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
