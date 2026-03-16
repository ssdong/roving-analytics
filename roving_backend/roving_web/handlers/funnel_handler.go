package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"roving_web/models"
	"roving_web/repositories"
	"roving_web/utils"
)

func FunnelHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	rawUrls := r.URL.Query().Get("url")

	// Decode the JSON into a Go slice (array)
	urlArray := strings.Split(rawUrls, ",")

	var decodedUrls []string
	for _, rawURL := range urlArray {
		decodedURL, err := url.PathUnescape(rawURL)
		if err != nil {
			fmt.Printf("Error decoding URL: %v\n", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Check if the decoded URL exceeds the maximum length
		if len(decodedURL) > models.MaximumUrlLength {
			fmt.Printf("Invalid URL: %s is too long\n", decodedURL)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		decodedUrls = append(decodedUrls, decodedURL)
	}

	repo := repositories.NewFunnelRepository()

	/**
	A result that comes back from Clickhouse will look like

	┌─Sequence─────┬─Count────┐
	│ ['/sign-up'] │       89 │
	└──────────────┴──────────┘
	┌─Sequence────────────────────────────┬─Count────┐
	│ ['/sign-up','/sign-in','/settings'] │       84 │
	└─────────────────────────────────────┴──────────┘
	┌─Sequence─────────────────────────────────────────┬─Count────┐
	│ ['/sign-up','/sign-in','/settings','/dashboard'] │    32293 │
	└──────────────────────────────────────────────────┴──────────┘
	┌─Sequence────────────────┬─Count────┐
	│ ['/sign-up','/sign-in'] │       83 │
	└─────────────────────────┴──────────┘

	However, in terms of Funnel analysis, the queries we executed are looking for distinct journeys
	where these events happened in the exact sequence. We would need to aggregate the numbers such that
	users that have made to the last step, e.g. /dashboard, should also be considered the users that have
	made it through '/sign-up' cuz that's exactly their first step
	**/

	results, err := repo.GetFunnelStats(decodedUrls, params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted)

	if err != nil {
        log.Printf("Funnel Error: %v", err)
        http.Error(w, "Internal Error", 500)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
