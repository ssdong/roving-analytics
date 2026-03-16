package repositories

import (
	"context"
	"fmt"
	"log"
	"roving_web/db"
	"sync"
)

type siteIdCache struct {
	mu      sync.RWMutex
	siteIds map[string]uint32
}

type SiteRepository struct {
	cache *siteIdCache
}

var (
	instance *SiteRepository
	once sync.Once
)

func (c *siteIdCache) getSiteId(hostname string) (uint32, bool) {
	c.mu.RLock()
	id, ok := c.siteIds[hostname]
	c.mu.RUnlock()
	return id, ok
}

func (c *siteIdCache) setSiteId(hostname string, id uint32) {
	c.mu.Lock()
	c.siteIds[hostname] = id
	c.mu.Unlock()
}


func GetSiteRepository() *SiteRepository {
	once.Do(func() {
		instance = &SiteRepository{
			cache: &siteIdCache{
				siteIds: make(map[string]uint32),
			},
		}
	})

	return instance
}

func(r *SiteRepository) GetSiteIdByHostname(hostname string) (uint32, error) {
	siteId, found := r.cache.getSiteId(hostname)
	
	if found {
		return siteId, nil
	}

	ctx := context.Background()
	conn, err := db.GetConnection()
	if err != nil {
		return 0, fmt.Errorf("Could not retrieve db connection: %w", err)
	}

	query := `
		SELECT SiteId 
		FROM roving.hostname_to_site_id 
		WHERE Hostname = ? 
	`

	err = conn.QueryRow(ctx, query, hostname).Scan(&siteId)
	if err != nil {
		log.Default().Printf("SiteId Lookup failed for %s: %v", hostname, err)
		return 0, err
	}

	r.cache.setSiteId(hostname, siteId)

	return siteId, nil
}
