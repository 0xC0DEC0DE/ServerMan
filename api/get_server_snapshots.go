package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/callowaysutton/servercon/auth"
	"github.com/gin-gonic/gin"
)

// GetServerSnapshots handles GET requests to retrieve snapshots for a specific server.
func GetServerSnapshots(c *gin.Context, ch *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverIDStr := c.Param("id")
		serverID, err := strconv.Atoi(serverIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "Invalid server ID"})
			return
		}

		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		// Get the user groups for authorization
		user, err := auth.GetUserByEmail(ch, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		// Check if user has access to this server
		var server Server
		err = ch.QueryRow("SELECT id, domain FROM servers_cache WHERE id = ?", serverID).Scan(&server.ID, &server.Domain)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, StatusMessage{Status: "error", Message: "Server not found"})
			} else {
				c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query server"})
			}
			return
		}

		// Check user permissions
		hasAccess := false
		for _, group := range user.Groups {
			if group == "*" {
				hasAccess = true
				break
			}
			rootDomain := getRootDomain(server.Domain)
			if group == rootDomain {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "Access denied"})
			return
		}

		// Get snapshots for this server
		snapshots := []ServerSnapshot{}
		rows, err := ch.Query("SELECT snapshot_id, name, created_at, size_gb, status FROM server_snapshots_cache WHERE server_id = ? ORDER BY created_at DESC", serverID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query snapshots"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var snapshot ServerSnapshot
			if err := rows.Scan(&snapshot.ID, &snapshot.Name, &snapshot.CreatedAt, &snapshot.SizeGB, &snapshot.Status); err != nil {
				c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to scan snapshot"})
				return
			}
			snapshots = append(snapshots, snapshot)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Error reading snapshots"})
			return
		}

		c.JSON(http.StatusOK, snapshots)
	}
}
