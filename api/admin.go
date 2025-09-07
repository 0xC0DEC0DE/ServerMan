package api

import (
	"database/sql"
	"net/http"

	"github.com/callowaysutton/servercon/auth"
	"github.com/gin-gonic/gin"
)

// hasAdminAccess checks if user has "*" or "callowaysutton" group access
func hasAdminAccess(groups []string) bool {
	for _, group := range groups {
		if group == "*" || group == "callowaysutton" {
			return true
		}
	}
	return false
}

// GetAllUsers returns all users for admin management
func GetAllUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		if !hasAdminAccess(user.Groups) {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Admin access required"})
			return
		}

		rows, err := db.Query("SELECT id, email, groups FROM users ORDER BY email")
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve users"})
			return
		}
		defer rows.Close()

		var users []auth.User
		for rows.Next() {
			var user auth.User
			var groups []string
			err := rows.Scan(&user.ID, &user.Email, &groups)
			if err != nil {
				continue
			}
			user.Groups = groups
			users = append(users, user)
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   users,
		})
	}
}

// AddUser creates a new user with specified email and groups
func AddUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		if !hasAdminAccess(user.Groups) {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Admin access required"})
			return
		}

		var request struct {
			Email  string   `json:"email" binding:"required"`
			Groups []string `json:"groups" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "Invalid request format"})
			return
		}

		// Get the next available ID
		var maxID sql.NullInt64
		err = db.QueryRow("SELECT max(id) FROM users").Scan(&maxID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to generate user ID"})
			return
		}

		nextID := int64(1)
		if maxID.Valid {
			nextID = maxID.Int64 + 1
		}

		// Insert the user
		_, err = db.Exec("INSERT INTO users (id, email, groups) VALUES (?, ?, ?)", nextID, request.Email, request.Groups)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to create user"})
			return
		}

		c.JSON(http.StatusOK, StatusMessage{Status: "success", Message: "User created successfully"})
	}
}

// UpdateUserGroups updates the groups for an existing user
// Note: ClickHouse arrays are immutable, so we use DELETE + INSERT instead of UPDATE
func UpdateUserGroups(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		if !hasAdminAccess(user.Groups) {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Admin access required"})
			return
		}

		userEmail := c.Param("email")
		if userEmail == "" {
			c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "User email is required"})
			return
		}

		var request struct {
			Groups []string `json:"groups" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "Invalid request format"})
			return
		}

		// ClickHouse arrays are immutable, so we need to delete and re-insert the row
		// Use a transaction to ensure atomicity
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Will be ignored if already committed

		// First, get the current user data
		var userID int
		var currentEmail string
		err = tx.QueryRow("SELECT id, email FROM users WHERE email = ?", userEmail).Scan(&userID, &currentEmail)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, StatusMessage{Status: "error", Message: "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to find user"})
			}
			return
		}

		// Delete the existing row
		_, err = tx.Exec("DELETE FROM users WHERE email = ?", userEmail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to delete existing user record"})
			return
		}

		// Insert the row with updated groups
		_, err = tx.Exec("INSERT INTO users (id, email, groups) VALUES (?, ?, ?)", userID, currentEmail, request.Groups)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to update user groups"})
			return
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to commit changes"})
			return
		}

		c.JSON(http.StatusOK, StatusMessage{Status: "success", Message: "User groups updated successfully"})
	}
}

// DeleteUser removes a user from the system
func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		if !hasAdminAccess(user.Groups) {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Admin access required"})
			return
		}

		userEmail := c.Param("email")
		if userEmail == "" {
			c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "User email is required"})
			return
		}

		// Delete the user
		_, err = db.Exec("DELETE FROM users WHERE email = ?", userEmail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, StatusMessage{Status: "success", Message: "User deleted successfully"})
	}
}

// TriggerSync allows admins to manually trigger a sync operation
func TriggerSync(db *sql.DB, syncFunc func() error) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		// Get the user groups - only allow users with "*" or "callowaysutton" groups to trigger sync
		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		if !hasAdminAccess(user.Groups) {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Admin access required"})
			return
		}

		// Trigger the sync
		if err := syncFunc(); err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to trigger sync: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, StatusMessage{Status: "success", Message: "Sync triggered successfully"})
	}
}

// GetAPIKeyStatus allows admins to view the current API key LRU status
func GetAPIKeyStatus(db *sql.DB, statusFunc func() []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		// Get the user groups - only allow users with "*" or "callowaysutton" groups to view API key status
		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		if !hasAdminAccess(user.Groups) {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Admin access required"})
			return
		}

		// Get API key status
		keyStatus := statusFunc()

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "API key status retrieved successfully",
			"data": gin.H{
				"api_keys":   keyStatus,
				"total_keys": len(keyStatus),
			},
		})
	}
}
