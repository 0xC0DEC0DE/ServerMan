package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/callowaysutton/servercon/auth"
	"github.com/gin-gonic/gin"
)

type ServerCredentials struct {
	RootPassword string `json:"root_password"`
	VncHost      string `json:"vnc_host"`
	VncPort      int    `json:"vnc_port"`
	VncPassword  string `json:"vnc_password"`
}

// GetServerCredentials handles GET requests to retrieve credentials for a specific server.
func GetServerCredentials(c *gin.Context, ch *sql.DB) gin.HandlerFunc {
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

		// Get credentials for this server
		var creds ServerCredentials
		err = ch.QueryRow("SELECT root_password, vnc_host, vnc_port, vnc_password FROM server_credentials_cache WHERE server_id = ?", serverID).Scan(&creds.RootPassword, &creds.VncHost, &creds.VncPort, &creds.VncPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, StatusMessage{Status: "error", Message: "Credentials not found"})
			} else {
				c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query credentials"})
			}
			return
		}

		c.JSON(http.StatusOK, creds)
	}
}
