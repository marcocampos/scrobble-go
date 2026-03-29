package lastfm

// TopItem pairs an item (Artist, Album, Track, or Tag) with a play/use weight.
type TopItem[T any] struct {
	Item   T
	Weight float64
}

// SimilarItem pairs an item (Artist or Track) with a similarity score 0–1.
type SimilarItem[T any] struct {
	Item  T
	Match float64
}

// PlayedTrack is a scrobble entry returned by user.getRecentTracks.
type PlayedTrack struct {
	Track        *Track
	Album        string
	PlaybackDate string // human-readable date string from the API
	Timestamp    string // Unix timestamp string; empty for now-playing entries
}

// LovedTrack is an entry returned by user.getLovedTracks.
type LovedTrack struct {
	Track     *Track
	Date      string
	Timestamp string
}

// LibraryItem is an artist entry in a user's library.
type LibraryItem struct {
	Artist    *Artist
	Playcount int
	Tagcount  int
}

// ChartDateRange holds the from/to timestamps for a weekly chart period.
type ChartDateRange struct {
	From string
	To   string
}

// ScrobbleParams holds the parameters for a single scrobble.
// Artist, Title, and Timestamp are required; all others are optional.
type ScrobbleParams struct {
	Artist       string
	Title        string
	Timestamp    int64
	Album        string
	AlbumArtist  string
	TrackNumber  int
	Duration     int
	StreamID     string
	Context      string
	MBID         string
	ChosenByUser *bool // nil means unset (API assumes true)
}
