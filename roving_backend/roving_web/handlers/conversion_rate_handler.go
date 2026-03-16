package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"roving_web/db"
	"roving_web/utils"
)

type ConversionRateResult struct {
	ConversionRatePercentage float64
}

// TODO: Migrate to repository layer
func ConversionRateHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r, true, true)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if params.TargetConversionPage == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ConversionRateResult{ConversionRatePercentage: 0.0})
		return
	}

	// The query string
	query := `
		WITH SuccessfulConversions AS (
			SELECT COUNT(DISTINCT JourneyId) as SuccessCount
			FROM roving.web_traffic_event
			WHERE SiteId = ? 
				AND PathName LIKE concat(?, '%')
				AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
		),
		
		TotalJourneys AS (
			SELECT COUNT(DISTINCT JourneyId) as TotalCount
			FROM roving.web_traffic_event
			WHERE SiteId = ? 
				AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
		)
		
		SELECT 
			COALESCE(ROUND(100 * SuccessCount / NULLIF(TotalCount, 0), 2), 0) as ConversionRate
		FROM SuccessfulConversions, TotalJourneys
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
		params.SiteId, params.TargetConversionPage, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted,
		params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted,
	)

	var result ConversionRateResult
	if err := row.Scan(&result.ConversionRatePercentage); err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
