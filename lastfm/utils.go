package lastfm

import (
	"crypto/md5" //nolint:gosec // MD5 is required by the Last.fm API for password hashing and request signing
	"fmt"
	"strconv"
)

// MD5 returns the hex-encoded MD5 digest of s.
// This is required by the Last.fm API for password hashing and request signing.
func MD5(s string) string {
	h := md5.Sum([]byte(s)) //nolint:gosec // MD5 is required by the Last.fm API for password hashing
	return fmt.Sprintf("%x", h)
}

// parseNumber converts a string to float64.
// Returns 0.0 if the string is empty or cannot be parsed.
func parseNumber(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseInt converts a string to int.
// Returns 0 if the string is empty or cannot be parsed.
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// parseInt64 converts a string to int64.
// Returns 0 if the string is empty or cannot be parsed.
// Use this for large counters (listeners, playcounts) to avoid int overflow on 32-bit platforms.
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
