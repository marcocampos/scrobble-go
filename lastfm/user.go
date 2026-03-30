package lastfm

import (
	"context"
	"fmt"
)

// User represents a Last.fm user account.
type User struct {
	Name   string
	client *Client
}

// newUser returns a User bound to the given client.
func newUser(name string, c *Client) *User {
	return &User{Name: name, client: c}
}

// GetName returns the username.
func (u *User) GetName() string { return u.Name }

// String implements fmt.Stringer.
func (u *User) String() string { return u.Name }

func (u *User) baseParams() map[string]string {
	return map[string]string{"user": u.Name}
}

// UserInfo holds the metadata returned by user.getInfo.
type UserInfo struct {
	Name       string
	RealName   string
	URL        string
	Country    string
	Age        int
	Gender     string
	Subscriber bool
	Playcount  int
	Playlists  int
	Bootstrap  int
	Images     map[ImageSize]string
	Registered string // Unix timestamp string
}

// GetInfo returns profile information about the user.
func (u *User) GetInfo(ctx context.Context) (*UserInfo, error) {
	doc, err := newAPIRequest(u.client, "user.getInfo", u.baseParams()).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetInfo: %w", err)
	}

	node := doc.find("user")
	if node == nil {
		return nil, &MalformedResponseError{
			NetworkName:     u.client.net.Name,
			UnderlyingError: fmt.Errorf("no <user> in user.getInfo response"),
		}
	}

	info := &UserInfo{
		Name:      extract(node, "name"),
		RealName:  extract(node, "realname"),
		URL:       extract(node, "url"),
		Country:   extract(node, "country"),
		Age:       parseInt(extract(node, "age")),
		Gender:    extract(node, "gender"),
		Playcount: parseInt(extract(node, "playcount")),
		Playlists: parseInt(extract(node, "playlists")),
		Images:    extractImages(node),
	}
	if reg := node.find("registered"); reg != nil {
		info.Registered = reg.attr("unixtime")
	}
	if sub := extract(node, "subscriber"); sub == "1" {
		info.Subscriber = true
	}
	return info, nil
}

// GetRecentTracks returns the user's recently played tracks.
// Pass limit ≤ 0 for the API default (50). Pass page ≤ 0 for page 1.
func (u *User) GetRecentTracks(ctx context.Context, limit, page int) ([]PlayedTrack, error) {
	params := u.baseParams()
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	doc, err := newAPIRequest(u.client, "user.getRecentTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetRecentTracks: %w", err)
	}

	var result []PlayedTrack
	for _, node := range doc.findAll("track") {
		// Skip the now-playing entry (has nowplaying="true" attribute)
		if node.attr("nowplaying") == "true" {
			continue
		}
		title := extract(node, "name")
		artistName := ""
		if artistNode := node.find("artist"); artistNode != nil {
			artistName = extract(artistNode, "name")
			if artistName == "" {
				artistName = artistNode.text()
			}
		}
		album := extract(node, "album")
		ts := ""
		dt := ""
		if dateNode := node.find("date"); dateNode != nil {
			ts = dateNode.attr("uts")
			dt = dateNode.text()
		}
		result = append(result, PlayedTrack{
			Track:        newTrack(artistName, title, u.client),
			Album:        album,
			PlaybackDate: dt,
			Timestamp:    ts,
		})
	}
	return result, nil
}

// GetNowPlaying returns the track the user is currently playing, or nil if
// the user is not playing anything.
func (u *User) GetNowPlaying(ctx context.Context) (*Track, error) {
	params := u.baseParams()
	params["limit"] = "1"
	doc, err := newAPIRequest(u.client, "user.getRecentTracks", params).execute(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("User.GetNowPlaying: %w", err)
	}
	for _, node := range doc.findAll("track") {
		if node.attr("nowplaying") == "true" {
			title := extract(node, "name")
			artistName := ""
			if artistNode := node.find("artist"); artistNode != nil {
				artistName = extract(artistNode, "name")
				if artistName == "" {
					artistName = artistNode.text()
				}
			}
			return newTrack(artistName, title, u.client), nil
		}
	}
	return nil, nil
}

// GetLovedTracks returns tracks the user has marked as loved.
func (u *User) GetLovedTracks(ctx context.Context, limit, page int) ([]LovedTrack, error) {
	params := u.baseParams()
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	doc, err := newAPIRequest(u.client, "user.getLovedTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetLovedTracks: %w", err)
	}

	var result []LovedTrack
	for _, node := range doc.findAll("track") {
		title := extract(node, "name")
		artistName := ""
		if artistNode := node.find("artist"); artistNode != nil {
			artistName = extract(artistNode, "name")
		}
		dt := ""
		ts := ""
		if dateNode := node.find("date"); dateNode != nil {
			ts = dateNode.attr("uts")
			dt = dateNode.text()
		}
		result = append(result, LovedTrack{
			Track:     newTrack(artistName, title, u.client),
			Date:      dt,
			Timestamp: ts,
		})
	}
	return result, nil
}

// GetTopArtists returns the user's most played artists.
func (u *User) GetTopArtists(ctx context.Context, period string, limit int) ([]TopItem[*Artist], error) {
	params := u.baseParams()
	if period != "" {
		params["period"] = period
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(u.client, "user.getTopArtists", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetTopArtists: %w", err)
	}
	return extractTopArtists(doc, u.client), nil
}

// GetTopAlbums returns the user's most played albums.
func (u *User) GetTopAlbums(ctx context.Context, period string, limit int) ([]TopItem[*Album], error) {
	params := u.baseParams()
	if period != "" {
		params["period"] = period
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(u.client, "user.getTopAlbums", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetTopAlbums: %w", err)
	}
	return extractTopAlbums(doc, u.client), nil
}

// GetTopTracks returns the user's most played tracks.
func (u *User) GetTopTracks(ctx context.Context, period string, limit int) ([]TopItem[*Track], error) {
	params := u.baseParams()
	if period != "" {
		params["period"] = period
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	doc, err := newAPIRequest(u.client, "user.getTopTracks", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetTopTracks: %w", err)
	}
	return extractTopTracks(doc, u.client), nil
}

// GetTopTags returns the user's most used tags.
func (u *User) GetTopTags(ctx context.Context, limit int) ([]TopItem[*Tag], error) {
	return getTopTagsRequest(ctx, u.client, "user", u.baseParams(), limit)
}

// GetPlaycount returns the total number of scrobbles for the user.
func (u *User) GetPlaycount(ctx context.Context) (int, error) {
	doc, err := newAPIRequest(u.client, "user.getInfo", u.baseParams()).execute(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("User.GetPlaycount: %w", err)
	}
	return parseInt(extract(doc, "playcount")), nil
}

// GetWeeklyChartDates returns the available weekly chart date ranges for this user.
func (u *User) GetWeeklyChartDates(ctx context.Context) ([]ChartDateRange, error) {
	doc, err := newAPIRequest(u.client, "user.getWeeklyChartList", u.baseParams()).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetWeeklyChartDates: %w", err)
	}
	var result []ChartDateRange
	for _, node := range doc.findAll("chart") {
		result = append(result, ChartDateRange{
			From: node.attr("from"),
			To:   node.attr("to"),
		})
	}
	return result, nil
}

// GetWeeklyArtistCharts returns the weekly artist chart for a given date range.
// Pass empty strings to get the most recent chart.
func (u *User) GetWeeklyArtistCharts(ctx context.Context, from, to string) ([]TopItem[*Artist], error) {
	params := u.baseParams()
	if from != "" && to != "" {
		params["from"] = from
		params["to"] = to
	}
	doc, err := newAPIRequest(u.client, "user.getWeeklyArtistChart", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetWeeklyArtistCharts: %w", err)
	}
	return extractTopArtists(doc, u.client), nil
}

// GetWeeklyTrackCharts returns the weekly track chart for a given date range.
func (u *User) GetWeeklyTrackCharts(ctx context.Context, from, to string) ([]TopItem[*Track], error) {
	params := u.baseParams()
	if from != "" && to != "" {
		params["from"] = from
		params["to"] = to
	}
	doc, err := newAPIRequest(u.client, "user.getWeeklyTrackChart", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetWeeklyTrackCharts: %w", err)
	}
	return extractTopTracks(doc, u.client), nil
}

// GetWeeklyAlbumCharts returns the weekly album chart for a given date range.
func (u *User) GetWeeklyAlbumCharts(ctx context.Context, from, to string) ([]TopItem[*Album], error) {
	params := u.baseParams()
	if from != "" && to != "" {
		params["from"] = from
		params["to"] = to
	}
	doc, err := newAPIRequest(u.client, "user.getWeeklyAlbumChart", params).execute(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("User.GetWeeklyAlbumCharts: %w", err)
	}
	return extractTopAlbums(doc, u.client), nil
}

// GetURL returns the Last.fm profile URL for this user.
func (u *User) GetURL(domain Domain) string {
	return entityURL(u.client, urlUser, domain, u.Name)
}
