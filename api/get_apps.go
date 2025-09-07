package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetApps handles GET requests to retrieve a list of apps from cache.
func GetApps(c *gin.Context, ch *sql.DB) {
	apps := []AppType{}

	rows, err := ch.Query("SELECT id, app, name FROM apps_cache ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query apps"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var app AppType
		if err := rows.Scan(&app.ID, &app.App, &app.Name); err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to scan app"})
			return
		}
		apps = append(apps, app)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Error reading apps"})
		return
	}

	c.JSON(http.StatusOK, apps)
}
