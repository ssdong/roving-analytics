package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"log"
	"roving_web/db"
	"roving_web/utils"
)

type CountryVisitors struct {
	CountryCode    string
	UniqueVisitors uint64
}

// TODO: Migrate to repository layer
func CountryRankingHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
    		CountryCode, 
    		COUNT(DISTINCT JourneyId) as UniqueVisitors
		FROM roving.web_traffic_event
		WHERE SiteId = ?
		AND Timestamp BETWEEN ? AND ?
		GROUP BY CountryCode
		ORDER BY UniqueVisitors DESC;
	`

	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch data from ClickHouse
	rows, err := clickhouseConn.Query(ctx, query, params.SiteId, params.TimestampStartFormatted, params.TimestampEndFormatted)

	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []CountryVisitors = make([]CountryVisitors, 0)

	for rows.Next() {
		var result CountryVisitors
		if err := rows.Scan(&result.CountryCode, &result.UniqueVisitors); err != nil {
			log.Default().Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
