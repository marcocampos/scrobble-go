# scrobble-go

> **AI-Assisted Project** — This project is designed and implemented with the assistance of AI and large language models (LLMs). All code, documentation, and configurations have been developed collaboratively with AI tools.

A Go client for the [Last.fm API](https://www.last.fm/api) (and compatible networks such as [Libre.fm](https://libre.fm)).

A port of [pylast](https://github.com/pylast/pylast) to Go.

## Installation

```sh
go get github.com/marcocampos/scrobble-go/lastfm
```

Requires Go 1.21 or later.

## Quick start

```go
import "github.com/marcocampos/scrobble-go/lastfm"

client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithSessionKey(sessionKey), // see Authentication below
    lastfm.WithRateLimit(),            // recommended
)

track, err := client.GetTrack("Iron Maiden", "The Trooper")
info, err := track.GetInfo(ctx)
fmt.Println(info.Title, "-", info.Artist, "|", info.Playcount, "plays")
```

## Authentication

Last.fm requires authentication for any operation that writes to a user's
profile (scrobbling, loving tracks, adding tags, etc.). Read-only operations
only need an API key and secret.

Obtain a key and secret from <https://www.last.fm/api/account/create>.

### Option 1 — Session key (recommended for long-running apps)

Authenticate once to get a session key; the key is valid indefinitely.
Store it somewhere safe and reuse it on subsequent runs.

```go
client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithSessionKey(storedSessionKey),
)
```

### Option 2 — Username + password (mobile auth)

The library authenticates automatically on the first write operation.

```go
client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithUsername(username),
    lastfm.WithPasswordHash(lastfm.MD5(password)),
)

// Or authenticate explicitly and capture the resulting session key:
err := client.AuthenticateWithPassword(ctx, username, lastfm.MD5(password))
```

### Option 3 — Web auth (OAuth-style, for user-facing apps)

Redirect the user to Last.fm, wait for their approval, then exchange the
token for a session key.

```go
gen := lastfm.NewSessionKeyGenerator(client)

// Step 1: get the URL and open it in the user's browser.
authURL, err := gen.GetWebAuthURL(ctx)
fmt.Println("Authorise here:", authURL)

// Step 2: after the user approves, retrieve the session key.
sessionKey, username, err := gen.GetWebAuthSessionKeyAndUsername(ctx, authURL, "")
```

See [examples/webauth](examples/webauth/main.go) for a complete runnable example.

## Examples

All examples read credentials from environment variables and accept flags for
the entity to look up. Run any of them with:

```sh
export LASTFM_API_KEY=...
export LASTFM_API_SECRET=...
```

| Example | Command | What it shows |
|---|---|---|
| [track-info](examples/track-info/main.go) | `go run ./examples/track-info -artist "Iron Maiden" -track "The Trooper"` | Core info, duration, wiki, top tags, similar tracks |
| [artist-info](examples/artist-info/main.go) | `go run ./examples/artist-info -artist "Iron Maiden"` | Bio, similar artists, top albums and tracks |
| [album-info](examples/album-info/main.go) | `go run ./examples/album-info -artist "Iron Maiden" -album "Piece of Mind"` | Track listing, cover image URL, wiki |
| [user-info](examples/user-info/main.go) | `go run ./examples/user-info -user <username>` | Profile, now-playing, recent/loved/top tracks |
| [search](examples/search/main.go) | `go run ./examples/search -type track -query "Trooper" -artist "Iron Maiden"` | Paginated search across artists, albums, tracks |
| [scrobble](examples/scrobble/main.go) | `go run ./examples/scrobble` | Authenticate, send now-playing, scrobble a track |
| [webauth](examples/webauth/main.go) | `go run ./examples/webauth` | Full browser-based OAuth-style auth flow |

## Scrobbling

```go
// Single scrobble — Timestamp must be a Unix timestamp (int64).
err := client.Scrobble(ctx, lastfm.ScrobbleParams{
    Artist:    "Iron Maiden",
    Title:     "The Trooper",
    Album:     "Piece of Mind",   // optional
    Timestamp: time.Now().Add(-3 * time.Minute).Unix(),
})

// Batch scrobble — automatically split into chunks of 50.
err := client.ScrobbleMany(ctx, []lastfm.ScrobbleParams{
    {Artist: "Artist A", Title: "Track 1", Timestamp: ts1},
    {Artist: "Artist B", Title: "Track 2", Timestamp: ts2},
    // ...up to any number; batching is handled internally
})

// Now-playing notification.
err := client.UpdateNowPlaying(ctx, lastfm.NowPlayingParams{
    Artist:   "Iron Maiden",
    Title:    "The Trooper",
    Duration: 248, // seconds, optional
})
```

See [examples/scrobble](examples/scrobble/main.go) for a complete runnable example.

## Reading data

### Artists

```go
artist := client.GetArtist("Iron Maiden")

info, err := artist.GetInfo(ctx)
// info.Name, info.Listeners, info.Playcount, info.BioSummary, ...

similar, err := artist.GetSimilar(ctx, 10)       // []SimilarItem[*Artist]
albums,  err := artist.GetTopAlbums(ctx, 5)       // []TopItem[*Album]
tracks,  err := artist.GetTopTracks(ctx, 5)       // []TopItem[*Track]
tags,    err := artist.GetTopTags(ctx, 10)         // []TopItem[*Tag]
```

### Tracks

```go
track := client.GetTrack("Iron Maiden", "The Trooper")

info, err := track.GetInfo(ctx)
// info.Title, info.Artist, info.Album, info.Duration (seconds), ...

similar, err := track.GetSimilar(ctx, 10)         // []SimilarItem[*Track]
tags,    err := track.GetTopTags(ctx, 10)          // []TopItem[*Tag]

err = track.Love(ctx)   // mark as loved (authenticated)
err = track.Unlove(ctx) // remove love mark (authenticated)
```

### Albums

```go
album := client.GetAlbum("Iron Maiden", "Piece of Mind")

info, err := album.GetInfo(ctx)
// info.Title, info.Artist, info.Playcount, info.Tracks, ...

imageURL := info.Images[lastfm.SizeExtraLarge]
```

### Users

```go
user := client.GetUser("someusername")

info,          err := user.GetInfo(ctx)
recentTracks,  err := user.GetRecentTracks(ctx, 50, 1)  // limit, page
lovedTracks,   err := user.GetLovedTracks(ctx, 50, 1)
topArtists,    err := user.GetTopArtists(ctx, lastfm.PeriodOverall, 10)
topAlbums,     err := user.GetTopAlbums(ctx, lastfm.Period7Days, 5)
topTracks,     err := user.GetTopTracks(ctx, lastfm.Period1Month, 5)
nowPlaying,    err := user.GetNowPlaying(ctx) // returns nil if not playing

// Weekly charts
dates,   err := user.GetWeeklyChartDates(ctx)
artists, err := user.GetWeeklyArtistCharts(ctx, dates[0].From, dates[0].To)
tracks,  err := user.GetWeeklyTrackCharts(ctx, "", "")  // empty = most recent
albums,  err := user.GetWeeklyAlbumCharts(ctx, "", "")
```

### Tags

```go
topTags, err := client.GetTopTags(ctx, 20)  // globally most used tags

tag := client.GetTag("heavy metal")
artists, err := tag.GetTopArtists(ctx, 10)
tracks,  err := tag.GetTopTracks(ctx, 10)
albums,  err := tag.GetTopAlbums(ctx, 10)
```

### Geographic charts

```go
artists, err := client.GetGeoTopArtists(ctx, "Germany", 10)
tracks,  err := client.GetGeoTopTracks(ctx, "Germany", "", 10)
```

### MBID lookups

```go
artist, err := client.GetArtistByMBID(ctx, "ca891d65-d9b0-4258-89f7-e6ba29d83767")
track,  err := client.GetTrackByMBID(ctx, "...")
album,  err := client.GetAlbumByMBID(ctx, "...")
```

## Tagging (authenticated)

```go
track := client.GetTrack("Iron Maiden", "The Trooper")

err = track.AddTags(ctx, []string{"classic", "nwobhm"})
err = track.RemoveTag(ctx, "classic")

userTags, err := track.GetTags(ctx) // tags set by the authenticated user
```

`AddTags`, `RemoveTag`, and `GetTags` are also available on `Artist` and `Album`.

## Search

```go
// Paginated search — call GetNextPage() repeatedly for more results.
search := client.SearchForArtist("iron maiden")
page1, err := search.GetNextPage(ctx) // []*Artist
page2, err := search.GetNextPage(ctx)

// Or request a specific page directly:
results, err := client.SearchForTrack("Iron Maiden", "Trooper").GetPage(ctx, 1)
results, err := client.SearchForAlbum("piece of mind").GetPage(ctx, 1)
```

## Caching

Caching avoids redundant network calls for repeated read-only requests.

```go
// In-memory cache (lost when process exits):
client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithCache(lastfm.NewMemoryCache()),
)

// Persistent disk cache (survives restarts, backed by bbolt):
cache, err := lastfm.NewBoltCache("/tmp/lastfm-cache.db")
if err != nil { log.Fatal(err) }
defer cache.Close()

client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithCache(cache),
)
```

Any type that satisfies the `CacheBackend` interface can be used:

```go
type CacheBackend interface {
    Get(key string) (string, bool)
    Set(key, value string)
}
```

## Rate limiting

The [Last.fm API Terms of Service §4.4](https://www.last.fm/api/tos) require
at least 200 ms between consecutive requests. Enable the built-in rate limiter
with `WithRateLimit()`:

```go
client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithRateLimit(),
)
```

## Custom HTTP client

Pass a custom `*http.Client` to configure proxies, transport settings, or
timeouts beyond the default 30-second deadline:

```go
httpClient := &http.Client{
    Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
    Timeout:   15 * time.Second,
}
client := lastfm.NewLastFMClient(apiKey, apiSecret,
    lastfm.WithHTTPClient(httpClient),
)
```

## Libre.fm

Use `NewLibreFMClient` to talk to [Libre.fm](https://libre.fm) instead:

```go
client := lastfm.NewLibreFMClient(apiKey, apiSecret,
    lastfm.WithSessionKey(sessionKey),
)
```

The API is identical — all methods work exactly the same way.

## Error handling

```go
info, err := artist.GetInfo(ctx)
if err != nil {
    var wsErr *lastfm.WSError
    var netErr *lastfm.NetworkError
    var malErr *lastfm.MalformedResponseError

    switch {
    case errors.As(err, &wsErr):
        // API-level error (invalid key, rate limit, etc.)
        fmt.Println("API error", wsErr.Status, wsErr.Details)
    case errors.As(err, &netErr):
        // Transport failure (no internet, timeout, etc.)
        fmt.Println("network error:", netErr)
    case errors.As(err, &malErr):
        // Unexpected response format
        fmt.Println("bad response:", malErr)
    default:
        fmt.Println("error:", err)
    }
}
```

Common `WSError` status codes are exported as constants (`lastfm.StatusInvalidAPIKey`,
`lastfm.StatusRateLimitExceeded`, `lastfm.StatusTokenUnauthorized`, etc.).

## Running tests

```sh
# Unit tests (no network, no credentials required):
go test ./...

# Integration tests against the real Last.fm API:
export LASTFM_API_KEY=...
export LASTFM_API_SECRET=...
export LASTFM_USERNAME=...          # required for user + write tests
export LASTFM_PASSWORD_HASH=$(echo -n "yourpassword" | md5sum | cut -d' ' -f1)

go test -tags integration -v -timeout 60s ./lastfm/...
```

Integration tests that perform write operations (scrobbles, love, tags) clean
up after themselves so your profile data stays unaffected.

## Project layout

```
scrobble.go/
├── lastfm/               # library package (import as .../lastfm)
│   ├── client.go         # Client struct, constructors, functional options
│   ├── auth.go           # SessionKeyGenerator (web + mobile auth)
│   ├── request.go        # HTTP transport, signing, rate limiting
│   ├── artist.go         # Artist type and methods
│   ├── album.go          # Album type and methods
│   ├── track.go          # Track type and methods
│   ├── user.go           # User type and methods
│   ├── tag.go            # Tag type and methods
│   ├── country.go        # Country type and methods
│   ├── scrobble.go       # Scrobble and UpdateNowPlaying
│   ├── search.go         # Paginated search (artist, album, track)
│   ├── network.go        # Client-level chart and factory methods
│   ├── cache.go          # CacheBackend, MemoryCache, BoltCache
│   ├── xml.go            # Internal XML parser
│   ├── entity.go         # Shared extraction helpers
│   ├── errors.go         # Error types and status codes
│   ├── types.go          # TopItem, SimilarItem, PlayedTrack, etc.
│   ├── utils.go          # MD5, number parsing
│   └── lastfm.go         # Package doc and constants
├── examples/
│   ├── track-info/       # Track info, wiki, similar tracks
│   ├── artist-info/      # Artist bio, similar artists, top albums/tracks
│   ├── album-info/       # Album info, track listing, cover image
│   ├── user-info/        # User profile, recent/loved/top tracks
│   ├── search/           # Paginated search for artists, albums, tracks
│   ├── scrobble/         # Scrobble + now-playing example
│   └── webauth/          # Web authentication flow example
├── go.mod
├── LICENSE
└── README.md
```

## Contributing

Contributions are welcome. Please follow these steps:

1. Fork the repository and create a branch from `main`.
2. Install the pre-commit hook:
   ```sh
   go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest
   go-pre-commit install
   ```
3. Make your changes. Add or update tests as appropriate.
4. Ensure all checks pass before opening a pull request:
   ```sh
   go-pre-commit run --all-files   # formatting, linting, mod tidy
   go test ./...                   # unit tests
   ```
5. Open a pull request against `main` with a clear description of what the change does and why.

Please keep pull requests focused — one feature or fix per PR. For larger changes, open an issue first to discuss the approach.

## License

MIT — see [LICENSE](LICENSE).
