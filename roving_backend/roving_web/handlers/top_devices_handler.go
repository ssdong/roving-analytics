package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"log"
	"roving_web/db"
	"roving_web/utils"
)

type DeviceResult struct {
	Device             string
	PercentageOfDevice float64
}

// TODO: Migrate to repository layer
func TopDevicesHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	query := `
		WITH TotalSessions AS (
			SELECT COUNT(DISTINCT JourneyId) as TotalSessionsCount
			FROM roving.web_traffic_event
			WHERE SiteId = ? AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
		)
		
		SELECT
			Browser as Device,
			ROUND(
				(COUNT(DISTINCT JourneyId) * 100.0) / TotalSessionsCount, 2
			) as PercentageOfDevices
		FROM roving.web_traffic_event, TotalSessions
		WHERE SiteId = ? AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
		GROUP BY Browser, TotalSessionsCount
		ORDER BY PercentageOfDevices DESC;
	`

	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rows, err := clickhouseConn.Query(ctx, query,
		params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted,
		params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted,
	)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []DeviceResult = make([]DeviceResult, 0)

	for rows.Next() {
		var result DeviceResult
		if err := rows.Scan(&result.Device, &result.PercentageOfDevice); err != nil {
			log.Default().Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
