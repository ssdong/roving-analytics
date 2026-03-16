package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type QueryParams struct {
	SiteId                  string
	TimestampStartFormatted string
	TimestampEndFormatted   string
	Timezone                string
	TargetConversionPage    string
}

const MaxSiteIdLength = 10                 // There are approximately 1.9 billion websites are there in the world as of Sept 2023. 10 digits should be enough to accommodate for that.
const DateStringLength = 19                // for "2006-01-02 15:04:05"
const MaxTimezoneLength = 30               // Maximum length for timezone string
const MaxTargetConversionPageLength = 2000 // Maximum length for targetConversionPage string

func isNaturalNumber(s string) bool {
	// Convert string to int
	num, err := strconv.Atoi(s)

	// Ensure there's no error in conversion and the number is greater than zero
	return err == nil && num >= 0
}

func formatAndValidateTimestamp(ts string) (string, error) {
	// Try to parse the input as a date-time in UTC format.
	t, err := time.Parse("2006-01-02 15:04:05", ts)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02 15:04:05"), nil
}

func ValidateSiteId(siteId string) error {
	if len(siteId) > MaxSiteIdLength {
		return fmt.Errorf("siteId too long")
	}

	if siteId == "" {
		return fmt.Errorf("siteId missing")
	}

	if !isNaturalNumber(siteId) {
		return fmt.Errorf("invalid siteId")
	}

	return nil
}

// ParseAndValidateQueryParams parses and validates the query parameters in the request.
// The options parameter is a variadic parameter that can be used to specify additional
// validation requirements. The first option is a boolean that specifies whether the
// targetConversionPage parameter is required. The second option is a boolean that
// specifies whether the timezone parameter is required.
func ParseAndValidateQueryParams(r *http.Request, options ...bool) (*QueryParams, error) {
	requireTargetConversionPage := false
	requireTimeZone := false
	if len(options) > 0 {
		requireTargetConversionPage = options[0]
		requireTimeZone = options[1]
	}

	siteId := r.URL.Query().Get("siteId")
	timestampStart := r.URL.Query().Get("timestampStart")
	timestampEnd := r.URL.Query().Get("timestampEnd")
	timezone := r.URL.Query().Get("timezone")
	targetConversionPage := r.URL.Query().Get("targetConversionPage")

	if requireTargetConversionPage && targetConversionPage == "" {
		return nil, fmt.Errorf("targetConversionPage parameter is required")
	}

	// If timezone parameter is empty, default to "UTC"
	if timezone == "" {
		if requireTimeZone {
			return nil, fmt.Errorf("timezone parameter is required")
		} else {
			timezone = "UTC"
		}
	}

	if err := ValidateSiteId(siteId); err != nil {
		return nil, err
	}

	if len(timestampStart) != DateStringLength || (len(timestampEnd) != DateStringLength && timestampEnd != "now") {
		return nil, fmt.Errorf("invalid timestamp format")
	}

	if len(timezone) > MaxTimezoneLength {
		return nil, fmt.Errorf("timezone string too long")
	}

	if len(targetConversionPage) > MaxTargetConversionPageLength {
		return nil, fmt.Errorf("targetConversionPage string too long")
	}

	if timestampStart == "" || timestampEnd == "" {
		return nil, fmt.Errorf("missing query parameters")
	}

	timestampStartFormatted, err := formatAndValidateTimestamp(timestampStart)
	if err != nil {
		return nil, fmt.Errorf("invalid timestampStart")
	}

	var timestampEndFormatted string

	// Convert "now" to the current server time
	if timestampEnd == "now" {
		t := time.Now().UTC() // Current server time in UTC

		// By setting timestampEnd to the start of the next day, we aim to:
		// 1. Accommodate for potential time zone differences between the server and the user.
		//    This ensures the data is up to date for users irrespective of their local time.
		// 2. Ensure we're capturing all events up to the current moment and slightly beyond
		//    to avoid missing any due to minor clock skews or discrepancies.

		// Add 24 hours to the current time to ensure we're in the next day.
		nextDay := t.Add(24 * time.Hour)

		// Truncate to the start of the next day.
		startOfNextDay := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, time.UTC)

		timestampEndFormatted = startOfNextDay.Format("2006-01-02 15:04:05")
	} else {
		timestampFormatted, err := formatAndValidateTimestamp(timestampEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid timestampEnd")
		}
		timestampEndFormatted = timestampFormatted
	}

	return &QueryParams{
		SiteId:                  siteId,
		TimestampStartFormatted: timestampStartFormatted,
		TimestampEndFormatted:   timestampEndFormatted,
		Timezone:                timezone,
		TargetConversionPage:    targetConversionPage,
	}, nil
}
