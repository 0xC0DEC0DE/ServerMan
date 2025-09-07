package api

import (
	"database/sql"
	"net/http"

	auth "github.com/callowaysutton/servercon/auth"
	"github.com/gin-gonic/gin"
)

func GetUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			c.JSON(http.StatusUnauthorized, StatusMessage{Status: "error", Message: "User not authenticated"})
			return
		}

		// Get user data from database
		user, err := auth.GetUserByEmail(db, email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to retrieve user data"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
