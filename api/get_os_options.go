package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetOsOptions(c *gin.Context, db *sql.DB) {
	os_options := []OsType{}

	rows, err := db.Query("SELECT id, name FROM os_types_cache ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to query OS types"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var osType OsType
		if err := rows.Scan(&osType.ID, &osType.Name); err != nil {
			c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Failed to scan OS type"})
			return
		}
		os_options = append(os_options, osType)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, StatusMessage{Status: "error", Message: "Error reading OS types"})
		return
	}

	c.JSON(http.StatusOK, os_options)
}
