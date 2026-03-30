package lastfm

import "fmt"

// API error status codes returned by the Last.fm web service.
const (
	StatusInvalidService      = 2
	StatusInvalidMethod       = 3
	StatusAuthFailed          = 4
	StatusInvalidFormat       = 5
	StatusInvalidParams       = 6
	StatusInvalidResource     = 7
	StatusOperationFailed     = 8
	StatusInvalidSK           = 9
	StatusInvalidAPIKey       = 10
	StatusOffline             = 11
	StatusSubscribersOnly     = 12
	StatusInvalidSignature    = 13
	StatusTokenUnauthorized   = 14
	StatusTokenExpired        = 15
	StatusTemporarilyUnavail  = 16
	StatusLoginRequired       = 17
	StatusTrialExpired        = 18
	StatusNotEnoughContent    = 20
	StatusNotEnoughMembers    = 21
	StatusNotEnoughFans       = 22
	StatusNotEnoughNeighbours = 23
	StatusNoPeakRadio         = 24
	StatusRadioNotFound       = 25
	StatusAPIKeySuspended     = 26
	StatusDeprecated          = 27
	StatusRateLimitExceeded   = 29
)

// WSError is returned when the Last.fm web service responds with an error.
type WSError struct {
	// Status is either a Last.fm API error code (e.g. StatusInvalidAPIKey)
	// or an HTTP status code (e.g. 502) for transient server failures.
	Status int
	// Details is the human-readable error message from the API or HTTP layer.
	Details     string
	networkName string
}

func (e *WSError) Error() string {
	return fmt.Sprintf("last.fm API error %d: %s", e.Status, e.Details)
}

// MalformedResponseError is returned when the API response cannot be parsed.
type MalformedResponseError struct {
	NetworkName     string
	UnderlyingError error
}

func (e *MalformedResponseError) Error() string {
	return fmt.Sprintf("malformed response from %s: %v", e.NetworkName, e.UnderlyingError)
}

func (e *MalformedResponseError) Unwrap() error { return e.UnderlyingError }

// NetworkError is returned when a transport-level error occurs.
type NetworkError struct {
	NetworkName     string
	UnderlyingError error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error talking to %s: %v", e.NetworkName, e.UnderlyingError)
}

func (e *NetworkError) Unwrap() error { return e.UnderlyingError }
