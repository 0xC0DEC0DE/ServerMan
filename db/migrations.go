package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB, dir string) {
	ctx := context.Background()

	// Ensure migrations table exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id String,
			applied_at DateTime DEFAULT now()
		) ENGINE = TinyLog
	`)
	if err != nil {
		log.Fatalf("failed to ensure schema_migrations table: %v", err)
	}

	// Get applied migrations
	applied := map[string]bool{}
	rows, err := db.QueryContext(ctx, "SELECT id FROM schema_migrations")
	if err != nil {
		log.Fatalf("failed to fetch applied migrations: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		rows.Scan(&id)
		applied[id] = true
	}

	// Load migration files
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		log.Fatalf("failed to list migrations: %v", err)
	}
	sort.Strings(files)

	// Apply missing migrations
	for _, file := range files {
		id := filepath.Base(file)
		if applied[id] {
			fmt.Printf("Skipping %s (already applied)\n", id)
			continue
		}

		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("failed to read %s: %v", file, err)
		}

		queries := strings.Split(string(sqlBytes), ";")
		for _, q := range queries {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			if _, err := db.ExecContext(ctx, q); err != nil {
				log.Fatalf("migration %s failed: %v", id, err)
			}
		}

		_, err = db.ExecContext(ctx,
			"INSERT INTO schema_migrations (id) VALUES (?)", id)
		if err != nil {
			log.Fatalf("failed to record migration %s: %v", id, err)
		}

		fmt.Printf("Applied migration %s\n", id)
	}
}
