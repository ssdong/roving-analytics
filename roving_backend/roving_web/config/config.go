package config

import (
	"os"
)

type Config struct {
	ReferrerSpamFilePath string
	GeoliteDBPath        string
	ClickHouseURL        string
	ClickHouseDB         string
	ClickHouseUserName   string
	ClickHousePassword   string
	RovingJourneyKey     string
}

var AppConfig *Config

func Load() {
	if AppConfig != nil {
		return
	}

	AppConfig = &Config{
		ReferrerSpamFilePath: getEnv("REFERRER_SPAM_FILE_PATH", "./referrer_spam_list.txt"),
		GeoliteDBPath:        getEnv("GEOLITE2_MMDB_FILE_PATH", "./GeoLite2-City.mmdb"),
		ClickHouseURL:        getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDB:         getEnv("CLICKHOUSE_DB", "roving"),
		ClickHouseUserName:   getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword:   getEnv("CLICKHOUSE_PASS", ""),
		RovingJourneyKey:     getEnv("ROVING_JOURNEY_KEY", "12345678901234567890123456789012"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

	