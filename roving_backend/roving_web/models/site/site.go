package site

import (
	"time"
)

type Site struct {
	Domain         string
	Timezone       string // represents the timezone of the user's domain
	Locked         bool   // indicates if the user's domain is locked or not(due to none payment or other violations)
	StatsStartDate time.Time

	IngestRateLimitScaleSeconds int // used for rate limiting
	IngestRateLimitThreshold    int
	DomainChangedFrom           string    // indicates previous domain if the domain was changed
	DomainChangedAt             time.Time // timestamp of when the domain was changed

	// Members                      []User        // list of members associated with the site
	// Memberships                  []Membership  // list of memberships related to the site
	// WeeklyReport                 WeeklyReport
	// MonthlyReport                MonthlyReport
	// FromCache                    bool          // virtual field to indicate if data is fetched from cache
	CreatedAt, UpdatedAt time.Time // timestamp fields
}
