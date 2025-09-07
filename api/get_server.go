package api

import (
	"database/sql"
	"net/http"

	"github.com/callowaysutton/servercon/auth"
	"github.com/gin-gonic/gin"
)

// GetServer handles GET requests to retrieve a server from cache.
func GetServer(c *gin.Context, ch *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		server := ServerDetail{}

		// Get specific server ID from URL parameters
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "Server ID is required"})
			return
		}

		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		// Get the user groups
		user, err := auth.GetUserByEmail(ch, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user groups"})
			return
		}

		// Query server details from cache
		err = ch.QueryRow(`
			SELECT name, state, ip_address, operating_system, memory, disk, cpu, vnc_status, daily_snapshots 
			FROM server_details_cache 
			WHERE server_id = ?
		`, id).Scan(&server.Name, &server.State, &server.IpAddress, &server.OperatingSystem,
			&server.Memory, &server.Disk, &server.Cpu, &server.VncStatus, &server.DailySnapshots)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, StatusMessage{Status: "error", Message: "Server not found"})
			} else {
				c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query server details"})
			}
			return
		}

		// If the user has the group "*", return the server
		for _, group := range user.Groups {
			if group == "*" {
				c.JSON(http.StatusOK, server)
				return
			}
		}

		// Filter server based on user groups, the group is the root domain of the server's FQDN
		rootDomain := getRootDomain(server.Name)
		for _, group := range user.Groups {
			if group == rootDomain {
				c.JSON(http.StatusOK, server)
				return
			}
		}

		c.JSON(http.StatusForbidden, StatusMessage{Status: "error", Message: "User does not have access to this server, the server does not exist or the server was terminated"})
	}
}
