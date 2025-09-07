package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"
)

// CreateSession creates a new session in the database and returns the session token
func CreateSession(db *sql.DB, email string) (string, error) {
	// Generate a secure session token
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	// Get the next available ID
	var maxID sql.NullInt64
	err = db.QueryRow("SELECT max(id) FROM user_sessions").Scan(&maxID)
	if err != nil {
		return "", err
	}

	nextID := int64(1)
	if maxID.Valid {
		nextID = maxID.Int64 + 1
	}

	// Insert the session into the database
	_, err = db.Exec(`
		INSERT INTO user_sessions (id, email, session_token, created_at)
		VALUES (?, ?, ?, ?)
	`, nextID, email, token, time.Now())

	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateSession checks if a session token is valid and returns the associated email
func ValidateSession(db *sql.DB, token string) (string, bool) {
	var email string
	err := db.QueryRow(`
		SELECT email 
		FROM user_sessions 
		WHERE session_token = ? 
		AND created_at > now() - INTERVAL 12 HOUR
	`, token).Scan(&email)

	if err != nil {
		return "", false
	}

	return email, true
}

// DeleteSession removes a session from the database
func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec("DELETE FROM user_sessions WHERE session_token = ?", token)
	return err
}

// GetUserByEmail retrieves user information by email
func GetUserByEmail(db *sql.DB, email string) (User, error) {
	var id uint64
	var userEmail string
	var groups []string

	err := db.QueryRow("SELECT id, email, groups FROM users WHERE email = ?", email).Scan(&id, &userEmail, &groups)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:     int(id),
		Email:  userEmail,
		Groups: groups,
	}, nil
}

// generateSessionToken generates a cryptographically secure session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
