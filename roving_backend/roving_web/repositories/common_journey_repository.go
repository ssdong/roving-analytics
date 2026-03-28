package repositories

import (
	"context"
	"log"
	"roving_web/db"
)

type CommonJourneyResult struct {
	Sequence   []string
	Count      uint64
	Percentage float64
}

type CommonJourneyRepository struct{}

func NewCommonJourneyRepository() *CommonJourneyRepository {
    return &CommonJourneyRepository{}
}

func (r *CommonJourneyRepository) GetCommonJourneyStat(siteId, timezone, timestampStart, timestampEnd string) ([]CommonJourneyResult, error) {
	query := `
		WITH JourneySequences AS (
			SELECT 
				JourneyId, 
				groupArray(PathName) as Paths
			FROM (
				SELECT JourneyId, PathName
				FROM roving.web_traffic_event
				WHERE SiteId = ?
					AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
				ORDER BY JourneyId, Timestamp
			)
			GROUP BY JourneyId
		),
		
		TotalJourneys AS (
			SELECT count() as TotalCount
			FROM JourneySequences
		)
		
		SELECT 
			Paths as Sequence, 
			count() as Count,
			ROUND(100.0 * count() / TotalCount, 2) as Percentage
		FROM JourneySequences, TotalJourneys
		GROUP BY Sequence, TotalCount
		ORDER BY Count DESC
		LIMIT 10;
		`

	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		return nil, err
	}

	rows, err := clickhouseConn.Query(ctx, query, siteId, timezone, timestampStart, timestampEnd)
	if err != nil {
		log.Default().Println(err)
		return nil, err
	}
	defer rows.Close()

	var results []CommonJourneyResult = make([]CommonJourneyResult, 0)
	for rows.Next() {
		var result CommonJourneyResult
		if err := rows.Scan(&result.Sequence, &result.Count, &result.Percentage); err != nil {
			log.Default().Println(err)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}