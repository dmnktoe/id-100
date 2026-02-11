package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"id-100/internal/database"
)

// StatsHandler returns lightweight stats for badges and integrations.
func StatsHandler(c echo.Context) error {
	totalDeriven, totalContribs, activeUsers, totalCities, lastActivity := database.GetFooterStats()

	response := map[string]interface{}{
		"total_contributions": totalContribs,
		"total_deriven":       totalDeriven,
		"active_users":        activeUsers,
		"total_cities":        totalCities,
		"last_activity":       nil,
	}

	if lastActivity.Valid {
		response["last_activity"] = lastActivity.Time.UTC().Format(time.RFC3339)
	}

	return c.JSON(http.StatusOK, response)
}
