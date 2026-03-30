// Package lastfm provides a Go client for the Last.fm API (and compatible
// networks such as Libre.fm).
//
// Basic usage:
//
//	client := lastfm.NewLastFMClient(apiKey, apiSecret,
//	    lastfm.WithUsername(username),
//	    lastfm.WithPasswordHash(lastfm.MD5("yourpassword")),
//	)
//
//	track, err := client.GetTrack(ctx, "Iron Maiden", "The Nomad")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	err = track.Love(ctx)
package lastfm

// Period constants used in chart and top-item requests.
const (
	PeriodOverall  = "overall"
	Period7Days    = "7day"
	Period1Month   = "1month"
	Period3Months  = "3month"
	Period6Months  = "6month"
	Period12Months = "12month"
)

// Domain identifies a localised Last.fm URL variant.
type Domain int

// Domain constants for localised Last.fm URLs.
const (
	DomainEnglish    Domain = 0
	DomainGerman     Domain = 1
	DomainSpanish    Domain = 2
	DomainFrench     Domain = 3
	DomainItalian    Domain = 4
	DomainPolish     Domain = 5
	DomainPortuguese Domain = 6
	DomainSwedish    Domain = 7
	DomainTurkish    Domain = 8
	DomainRussian    Domain = 9
	DomainJapanese   Domain = 10
	DomainChinese    Domain = 11
)

// ImageSize identifies the resolution of a cover image.
type ImageSize int

// Size constants for cover images.
const (
	SizeSmall      ImageSize = 0
	SizeMedium     ImageSize = 1
	SizeLarge      ImageSize = 2
	SizeExtraLarge ImageSize = 3
	SizeMega       ImageSize = 4
)

// Image ordering constants.
const (
	ImagesOrderPopularity = "popularity"
	ImagesOrderDate       = "dateadded"
)
