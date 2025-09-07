package auth

import (
	"database/sql"
	"log"
	"time"
)

// CleanupExpiredSessions removes expired sessions from the database
func CleanupExpiredSessions(db *sql.DB) error {
	result, err := db.Exec(`
		DELETE FROM user_sessions 
		WHERE created_at < now() - INTERVAL 12 HOUR
	`)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired sessions", rowsAffected)
	}

	return nil
}

// StartSessionCleanup starts a goroutine that periodically cleans up expired sessions
func StartSessionCleanup(db *sql.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := CleanupExpiredSessions(db); err != nil {
				log.Printf("Failed to cleanup expired sessions: %v", err)
			}
		}
	}()
}
