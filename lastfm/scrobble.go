package lastfm

import (
	"context"
	"fmt"
	"strconv"
)

// maxScrobbleBatch is the maximum number of tracks per scrobble.batch call,
// as defined by the Last.fm API.
const maxScrobbleBatch = 50

// Scrobble records a single track play to the authenticated user's profile.
// Artist, Title, and Timestamp are required; all other fields in p are optional.
func (c *Client) Scrobble(ctx context.Context, p ScrobbleParams) error {
	return c.ScrobbleMany(ctx, []ScrobbleParams{p})
}

// ScrobbleMany records up to many track plays in a single batch.
// The slice is automatically split into chunks of 50 (the API limit).
func (c *Client) ScrobbleMany(ctx context.Context, tracks []ScrobbleParams) error {
	for len(tracks) > 0 {
		batch := tracks
		if len(batch) > maxScrobbleBatch {
			batch = tracks[:maxScrobbleBatch]
		}
		tracks = tracks[len(batch):]

		if err := c.scrobbleBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// scrobbleBatch sends one API call for up to 50 tracks.
func (c *Client) scrobbleBatch(ctx context.Context, tracks []ScrobbleParams) error {
	params := make(map[string]string, len(tracks)*5)

	for i, t := range tracks {
		idx := strconv.Itoa(i)
		params["artist["+idx+"]"] = t.Artist
		params["track["+idx+"]"] = t.Title
		params["timestamp["+idx+"]"] = strconv.FormatInt(t.Timestamp, 10)

		if t.Album != "" {
			params["album["+idx+"]"] = t.Album
		}
		if t.AlbumArtist != "" {
			params["albumArtist["+idx+"]"] = t.AlbumArtist
		}
		if t.TrackNumber != 0 {
			params["trackNumber["+idx+"]"] = strconv.Itoa(t.TrackNumber)
		}
		if t.Duration != 0 {
			params["duration["+idx+"]"] = strconv.Itoa(t.Duration)
		}
		if t.StreamID != "" {
			params["streamID["+idx+"]"] = t.StreamID
		}
		if t.Context != "" {
			params["context["+idx+"]"] = t.Context
		}
		if t.MBID != "" {
			params["mbid["+idx+"]"] = t.MBID
		}
		if t.ChosenByUser != nil {
			if *t.ChosenByUser {
				params["chosenByUser["+idx+"]"] = "1"
			} else {
				params["chosenByUser["+idx+"]"] = "0"
			}
		}
	}

	r := newAPIRequest(c, "track.scrobble", params)
	_, err := r.execute(ctx, false)
	if err != nil {
		return fmt.Errorf("ScrobbleMany: %w", err)
	}
	return nil
}

// UpdateNowPlaying notifies Last.fm that the authenticated user has started
// playing a track. Only Artist and Title are required.
func (c *Client) UpdateNowPlaying(ctx context.Context, p NowPlayingParams) error {
	params := map[string]string{
		"artist": p.Artist,
		"track":  p.Title,
	}
	if p.Album != "" {
		params["album"] = p.Album
	}
	if p.AlbumArtist != "" {
		params["albumArtist"] = p.AlbumArtist
	}
	if p.Duration != 0 {
		params["duration"] = strconv.Itoa(p.Duration)
	}
	if p.TrackNumber != 0 {
		params["trackNumber"] = strconv.Itoa(p.TrackNumber)
	}
	if p.MBID != "" {
		params["mbid"] = p.MBID
	}
	if p.Context != "" {
		params["context"] = p.Context
	}

	r := newAPIRequest(c, "track.updateNowPlaying", params)
	_, err := r.execute(ctx, false)
	if err != nil {
		return fmt.Errorf("UpdateNowPlaying: %w", err)
	}
	return nil
}

// NowPlayingParams holds the parameters for an UpdateNowPlaying call.
// Artist and Title are required; all others are optional.
type NowPlayingParams struct {
	Artist      string
	Title       string
	Album       string
	AlbumArtist string
	Duration    int
	TrackNumber int
	MBID        string
	Context     string
}
