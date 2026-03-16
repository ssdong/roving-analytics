package repositories

import (
	"context"
	"fmt"
	"log"
	"roving_web/db"
	"strings"
)

type FunnelResult struct {
	Sequence []string
	Count    uint64
}

type FunnelRepository struct{}

func NewFunnelRepository() *FunnelRepository {
    return &FunnelRepository{}
}

func (r *FunnelRepository) GetFunnelStats(urls []string, siteId, timezone, timestampStart, timestampEnd string) ([]FunnelResult, error) {
	ctx := context.Background()
	clickhouseConn, err := db.GetConnection()
	if err != nil {
		log.Default().Println(err)
		return nil, err
	}

	sql := r.generateFunnelSQL(urls, siteId, timezone, timestampStart, timestampEnd)

	if sql == "" {
		return []FunnelResult{}, nil
	}

	rows, err := clickhouseConn.Query(ctx, sql)
	if err != nil {
		log.Default().Println(err)
		return nil, err
	}
	defer rows.Close()

	results := make([]FunnelResult, 0, len(urls))
	for rows.Next() {
		var result FunnelResult
		if err := rows.Scan(&result.Sequence, &result.Count); err != nil {
			log.Default().Println(err)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}


// An example of a generated SQL based on sequence ["/home", "/blog", "/press", "/activation"] looks like
/**
	WITH JourneySequences AS (
		SELECT
			JourneyId,
			groupArray(PathName) AS urls
		FROM (
			SELECT JourneyId, PathName
			FROM roving.web_traffic_event
			WHERE SiteId = 1 AND toTimeZone(Timestamp,'America/Toronto')
			BETWEEN '2026-03-01' AND '2026-03-08'
			ORDER BY JourneyId, Timestamp
		)
		GROUP BY JourneyId
	)

	SELECT ['/home'] AS Sequence, count() AS Count
	FROM JourneySequences
	WHERE arraySlice(urls,1,1) = ['/home']

	UNION ALL

	SELECT ['/home','/blog'] AS Sequence, count() AS Count
	FROM JourneySequences
	WHERE arraySlice(urls,1,2) = ['/home','/blog']

	UNION ALL

	SELECT ['/home','/blog','/press'] AS Sequence, count() AS Count
	FROM JourneySequences
	WHERE arraySlice(urls,1,3) = ['/home','/blog','/press']

	UNION ALL

	SELECT ['/home','/blog','/press','/activation'] AS Sequence, count() AS Count
	FROM JourneySequences
	WHERE arraySlice(urls,1,4) = ['/home','/blog','/press','/activation']
**/
func (r *FunnelRepository) generateFunnelSQL(urls []string, siteId string, timezone string, timestampStart string, timestampEnd string) string {
	if len(urls) == 0 {
		return ""
	}

	baseCTE := `
        WITH JourneySequences AS (
            SELECT
                JourneyId,
                groupArray(PathName) AS urls
            FROM (
                SELECT JourneyId, PathName
                FROM roving.web_traffic_event
                WHERE SiteId = %s AND toTimeZone(Timestamp, '%s') BETWEEN '%s' AND '%s'
                ORDER BY JourneyId, Timestamp
            )
            GROUP BY JourneyId
        )
    `

	// Insert timezone and timestamps into the base CTE
	baseCTE = fmt.Sprintf(baseCTE, siteId, timezone, timestampStart, timestampEnd)

	queries := make([]string, 0, len(urls))
	for i := 1; i <= len(urls); i++ {
		currentSequence := urls[:i]
		// Convert the slice of strings to a single string with comma-separated values
		joinedSequence := "['" + strings.Join(currentSequence, "', '") + "']"
		query := fmt.Sprintf(`
            SELECT
                %s AS Sequence,
                count() AS Count
            FROM JourneySequences
            WHERE arraySlice(urls, 1, %d) = %s
        `, joinedSequence, i, joinedSequence)
		queries = append(queries, query)
	}

	return baseCTE + strings.Join(queries, "\nUNION ALL\n")
}
