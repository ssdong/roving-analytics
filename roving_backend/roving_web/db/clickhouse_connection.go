package db

import (
	"context"
	"fmt"
	"log"
	"roving_web/config"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var (
	instance    driver.Conn
    once        sync.Once
)

func GetConnection() (driver.Conn, error) {
    var err error
    once.Do(func() {
        instance, err = initializeDBConnection()
        if err != nil {
            log.Fatalf("Failed to connect to ClickHouse: %v", err)
        }
    })
    return instance, err
}

func initializeDBConnection() (driver.Conn, error) {
    ctx := context.Background()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{config.AppConfig.ClickHouseURL},
		Auth: clickhouse.Auth{
			Database: config.AppConfig.ClickHouseDB,
			Username: config.AppConfig.ClickHouseUserName,
			Password: config.AppConfig.ClickHousePassword,
		},
		DialTimeout:     time.Second * 60,
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,

		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "roving-web-analytics", Version: "0.1"},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	return conn, nil
}

func isConnectionHealthy(conn *driver.Conn) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := (*conn).Ping(ctx); err != nil {
		return false
	}
	return true
}
