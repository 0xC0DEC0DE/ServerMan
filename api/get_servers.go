package api

import (
	"database/sql"
	"net/http"

	"github.com/callowaysutton/servercon/auth"
	"github.com/gin-gonic/gin"
)

// GetServers handles GET requests to retrieve a list of servers from cache.
func GetServers(c *gin.Context, ch *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		servers := []Server{}

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

		rows, err := ch.Query("SELECT id, domain, reg_date, billing_cycle, next_due_date, domain_status FROM servers_cache ORDER BY id")
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query servers"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var server Server
			if err := rows.Scan(&server.ID, &server.Domain, &server.RegDate, &server.BillingCycle, &server.NextDueDate, &server.DomainStatus); err != nil {
				c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to scan server"})
				return
			}
			servers = append(servers, server)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Error reading servers"})
			return
		}

		// If the user has the group "*", return all servers
		for _, group := range user.Groups {
			if group == "*" {
				c.JSON(http.StatusOK, servers)
				return
			}
		}

		// Filter servers based on user groups, the group is the root domain of the server's FQDN
		filteredServers := []Server{}
		for _, server := range servers {
			rootDomain := getRootDomain(server.Domain)
			for _, group := range user.Groups {
				if group == rootDomain {
					filteredServers = append(filteredServers, server)
				}
			}
		}

		c.JSON(http.StatusOK, filteredServers)
	}
}
