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

	sql, args := r.generateFunnelSQL(urls, siteId, timezone, timestampStart, timestampEnd)

	if sql == "" {
		return []FunnelResult{}, nil
	}

	rows, err := clickhouseConn.Query(ctx, sql, args...)
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
                WHERE SiteId = '1' AND toTimeZone(Timestamp, 'America/Toronto') BETWEEN '2026-03-01 00:00:00' AND '2026-03-28 23:59:59'
                ORDER BY JourneyId, Timestamp
            )
            GROUP BY JourneyId
        )
    
			SELECT
                ['/home'] AS Sequence,
                count() AS Count
            FROM JourneySequences
            WHERE arraySlice(urls, 1, 1) = ['/home']
		
	UNION ALL

			SELECT
                ['/home', '/blog'] AS Sequence,
                count() AS Count
            FROM JourneySequences
            WHERE arraySlice(urls, 1, 2) = ['/home', '/blog']
		
	UNION ALL

			SELECT
                ['/home', '/blog', '/press'] AS Sequence,
                count() AS Count
            FROM JourneySequences
            WHERE arraySlice(urls, 1, 3) = ['/home', '/blog', '/press']
		
	UNION ALL

			SELECT
                ['/home', '/blog', '/press', '/activation'] AS Sequence,
                count() AS Count
            FROM JourneySequences
            WHERE arraySlice(urls, 1, 4) = ['/home', '/blog', '/press', '/activation']
**/
func (r *FunnelRepository) generateFunnelSQL(urls []string, siteId string, timezone string, timestampStart string, timestampEnd string) (string, []any) {
	if len(urls) == 0 {
		return "", nil
	}

	baseCTE := `
        WITH JourneySequences AS (
            SELECT
                JourneyId,
                groupArray(PathName) AS urls
            FROM (
                SELECT JourneyId, PathName
                FROM roving.web_traffic_event
                WHERE SiteId = ? AND toTimeZone(Timestamp, ?) BETWEEN ? AND ?
                ORDER BY JourneyId, Timestamp
            )
            GROUP BY JourneyId
        )
    `

	var query strings.Builder
	args := make([]any, 0, 4 + len(urls))

	args = append(args, siteId, timezone, timestampStart, timestampEnd)

	query.WriteString(baseCTE)

	for i := 1; i <= len(urls); i++ {
		if i > 1 {
			query.WriteString("\nUNION ALL\n")
		}

		currentSequence := make([]string, i)
		copy(currentSequence, urls[:i])

		query.WriteString(fmt.Sprintf(`
			SELECT
                ? AS Sequence,
                count() AS Count
            FROM JourneySequences
            WHERE arraySlice(urls, 1, %d) = ?
		`, i))

		args = append(args, currentSequence, currentSequence)
	}

	return query.String(), args
}
