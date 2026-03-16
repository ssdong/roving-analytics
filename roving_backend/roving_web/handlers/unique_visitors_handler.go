package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"log"
	"roving_web/db"
	"roving_web/utils"
	"time"
)

type UniqueVisitorsResult struct {
	Date           string
	UniqueVisitors uint64
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// TODO: Migrate to repository layer
func UniqueVisitorsHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	/*
		The SQL query below is crafted to handle a specific dilemma tied to the
		representation of web traffic timestamps and their relation to user-specific time zones:

		Dilemma:
		The underlying data store saves web traffic events with timestamps in UTC. However, when
		analyzing traffic, especially on a day-by-day basis, it's vital to respect the local time zones
		of the website or the user querying the data. The difference between UTC and local times can
		lead to significant variations in reporting. For instance, traffic peaks that occur late evening
		in one timezone might get registered as the next day's traffic if interpreted in UTC.

		Example:
		Consider a website primarily serving users in the "America/Toronto" time zone. If there's a
		significant traffic surge at 11:00 PM EST on January 1st, this traffic, when viewed in UTC,
		would be timestamped as 4:00 AM on January 2nd. Without time zone conversion, a daily report
		would incorrectly attribute this traffic to January 2nd instead of January 1st.

		Solution:
		To address this, we employ the `toTimeZone` function to convert the stored UTC timestamp into
		the user-specified local time zone before performing any aggregation or filtering. This ensures
		that all date and time-based operations respect the local context, providing accurate day-by-day
		traffic insights.

		Note: It's crucial to maintain consistent usage of the `toTimeZone` function throughout the query. This
		ensures synchronization in date-time calculations and results that align with the user's intended time zone.
	*/
	query := `
		SELECT 
			DATE(toTimeZone(Timestamp, ?)) as Date,
			COUNT(DISTINCT JourneyId) as UniqueVisitors
		FROM roving.web_traffic_event
		WHERE SiteId = ? AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
		GROUP BY DATE(toTimeZone(Timestamp, ?))
		ORDER BY DATE(toTimeZone(Timestamp, ?));

	`

	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rows, err := clickhouseConn.Query(
		ctx, query,
		params.Timezone,
		params.SiteId,
		params.Timezone,
		params.TimestampStartFormatted,
		params.TimestampEndFormatted,
		params.Timezone,
		params.Timezone,
	)

	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []UniqueVisitorsResult = make([]UniqueVisitorsResult, 0)
	for rows.Next() {
		var result UniqueVisitorsResult
		var date time.Time

		if err := rows.Scan(&date, &result.UniqueVisitors); err != nil {
			log.Default().Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		result.Date = formatDate(date)
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
