package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"roving_web/db"
	"roving_web/utils"
)

type BounceRateResult struct {
	BounceRatePercentage float64
}

// TODO: Migrate to repository layer
func BounceRateHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r, false, true)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// If there are two identical JourneyIds and they correspond to two different pages,
	// it is not a bounce. However, if those two identical JourneyIds correspond to the
	// same page (due to, for example, a page refresh), then it could still be considered a
	// bounce, because the user only interacted with one page.
	// Therefore we need to do a GROUP BY JourneyId and HAVING COUNT(DISTINCT PathName) = 1
	query := `
		WITH SinglePageSessionsCount AS (
			SELECT COUNT(*) as SinglePageCount
			FROM (
				SELECT JourneyId
				FROM roving.web_traffic_event
				WHERE SiteId = ? 
					AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
				GROUP BY JourneyId
				HAVING COUNT(DISTINCT PathName) = 1
			)
		),
		
		TotalJourneys AS (
			SELECT COUNT(DISTINCT JourneyId) as TotalCount
			FROM roving.web_traffic_event
			WHERE SiteId = ? 
				AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
		)
		
		SELECT
			COALESCE(ROUND(100 * SinglePageCount / NULLIF(TotalCount, 0), 2), 0) as BounceRate
		FROM SinglePageSessionsCount, TotalJourneys	
		`

	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	row := clickhouseConn.QueryRow(
		ctx, query,
		params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted,
		params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted,
	)

	var result BounceRateResult
	if err := row.Scan(&result.BounceRatePercentage); err != nil {
		log.Default().Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
