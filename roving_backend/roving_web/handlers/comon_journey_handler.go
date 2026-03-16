package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"roving_web/repositories"
	"roving_web/utils"
)

func CommonJourneyHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.ParseAndValidateQueryParams(r)
	if err != nil {
		log.Default().Println(err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	repo := repositories.NewCommonJourneyRepository()

	results, err := repo.GetCommonJourneyStat(params.SiteId, params.Timezone, params.TimestampStartFormatted, params.TimestampEndFormatted)

	if err != nil {
        log.Printf("Common Journey Error: %v", err)
        http.Error(w, "Internal Error", 500)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
