package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"id-100/internal/database"
	"id-100/internal/utils"
)

// LivenessHandler is a lightweight probe that only reports whether the process
// is up. It never touches external dependencies so orchestrators and container
// healthchecks do not restart the app over a transient database/storage blip.
func LivenessHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "id-100",
	})
}

// ReadinessHandler verifies that the critical dependencies (database and object
// storage) are reachable. It returns 503 when any dependency is unavailable so
// that uptime monitoring reflects real readiness instead of flapping on every
// page that happens to hit a degraded backend.
func ReadinessHandler(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	ready := true

	if err := database.DB.Ping(ctx); err != nil {
		log.Printf("Readiness: database check failed: %v", err)
		checks["database"] = "down"
		ready = false
	} else {
		checks["database"] = "ok"
	}

	if err := utils.CheckS3(ctx); err != nil {
		log.Printf("Readiness: storage check failed: %v", err)
		checks["storage"] = "down"
		ready = false
	} else {
		checks["storage"] = "ok"
	}

	status := http.StatusOK
	overall := "ok"
	if !ready {
		status = http.StatusServiceUnavailable
		overall = "degraded"
	}

	return c.JSON(status, map[string]interface{}{
		"status":  overall,
		"service": "id-100",
		"checks":  checks,
	})
}
