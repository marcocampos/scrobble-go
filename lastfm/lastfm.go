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

// Domain constants for localised Last.fm URLs.
const (
	DomainEnglish    = 0
	DomainGerman     = 1
	DomainSpanish    = 2
	DomainFrench     = 3
	DomainItalian    = 4
	DomainPolish     = 5
	DomainPortuguese = 6
	DomainSwedish    = 7
	DomainTurkish    = 8
	DomainRussian    = 9
	DomainJapanese   = 10
	DomainChinese    = 11
)

// Size constants for cover images.
const (
	SizeSmall      = 0
	SizeMedium     = 1
	SizeLarge      = 2
	SizeExtraLarge = 3
	SizeMega       = 4
)

// Image ordering constants.
const (
	ImagesOrderPopularity = "popularity"
	ImagesOrderDate       = "dateadded"
)
