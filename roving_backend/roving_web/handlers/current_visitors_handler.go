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

type CurrentVisitorsResult struct {
	Count uint64
}

// TODO: Migrate to repository layer
func CurrentVisitorsHandler(w http.ResponseWriter, r *http.Request) {
	siteId := r.URL.Query().Get("siteId")
	if err := utils.ValidateSiteId(siteId); err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Define the time range for the past 30 minutes
	endTime := time.Now().UTC()
	startTime := endTime.Add(-30 * time.Minute)
	formattedStartTime := startTime.Format("2006-01-02 15:04:05")
	formattedEndTime := endTime.Format("2006-01-02 15:04:05")

	query := `
		SELECT COUNT(DISTINCT JourneyId) as Count
		FROM roving.web_traffic_event
		WHERE SiteId = ? AND Timestamp BETWEEN ? AND ?
	`

	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch data from ClickHouse
	row := clickhouseConn.QueryRow(
		ctx, query,
		siteId,
		formattedStartTime,
		formattedEndTime,
	)

	var result CurrentVisitorsResult

	if err := row.Scan(&result.Count); err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
