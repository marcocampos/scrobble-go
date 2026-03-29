# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-30

### Added

- Go client for the Last.fm API and Libre.fm
- Artist, album, track, user, tag, and country support
- Paginated search for artists, albums, and tracks
- Scrobbling and now-playing notifications
- Three authentication methods: session key, mobile (username + password), and web auth (OAuth-style)
- In-memory and persistent (bbolt) cache backends
- Built-in rate limiter (≥200 ms between calls, per Last.fm ToS §4.4)
- Custom HTTP client support
- Structured error types: `WSError`, `NetworkError`, `MalformedResponseError`
- Examples for all major features
- Pre-commit hooks via go-pre-commit (fumpt, golangci-lint, mod-tidy, eof, whitespace)
- GitHub Actions CI pipeline (build, test, lint)
- GitHub Actions CD pipeline (automatic GitHub Release on version tag)

[Unreleased]: https://github.com/marcocampos/scrobble-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/marcocampos/scrobble-go/releases/tag/v0.1.0
