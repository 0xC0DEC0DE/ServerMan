package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func Connect() *sql.DB {
	dsn := os.Getenv("CLICKHOUSE_DSN")
	if dsn == "" {
		dsn = "clickhouse://default:@localhost:9000/default"
	}

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatalf("failed to connect to ClickHouse: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("ClickHouse not reachable: %v", err)
	}

	fmt.Println("Connected to ClickHouse 🚀")
	return db
}
