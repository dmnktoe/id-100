package admin

import (
	"context"
	"log"
	"net/http"
	"strconv"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
)

// AdminBagRequestCompleteHandler marks a bag request as handled
func AdminBagRequestCompleteHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	rowsAffected, err := repository.MarkBagRequestHandled(context.Background(), id)
	if err != nil {
		log.Printf("Failed to mark bag_request handled: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.CaptureException(err)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
