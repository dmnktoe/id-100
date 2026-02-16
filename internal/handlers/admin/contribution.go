package admin

import (
	"log"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
	"id-100/internal/sentryhelper"
	"id-100/internal/utils"
)

// AdminDeleteContributionHandler deletes a contribution from the admin panel
func AdminDeleteContributionHandler(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.Atoi(contributionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}

	ctx := c.Request().Context()

	// Get the image URL and token_id before deletion
	imageURL, tokenID, err := repository.GetContributionForAdminDeletion(ctx, contributionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Delete from upload_logs first (foreign key reference)
	err = repository.DeleteUploadLog(ctx, contributionID)
	if err != nil {
		log.Printf("Failed to delete from upload_logs: %v", err)
		sentryhelper.CaptureException(c, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete upload log"})
	}

	// Delete from contributions table
	rowsAffected, err := repository.DeleteContribution(ctx, contributionID)
	if err != nil {
		log.Printf("Failed to delete contribution: %v", err)
		sentryhelper.CaptureException(c, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contribution"})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Decrement the total_uploads counter for the token
	err = repository.DecrementTokenUploadCount(ctx, tokenID)
	if err != nil {
		log.Printf("Failed to decrement upload counter: %v", err)
		sentryhelper.CaptureError(c, err, sentry.LevelWarning)
	}

	// Delete from S3 storage if the image exists
	if imageURL != "" {
		s3Err := utils.DeleteFromS3(ctx, imageURL)
		if s3Err != nil {
			log.Printf("Failed to delete from S3 (continuing anyway): %v", s3Err)
			sentryhelper.CaptureError(c, s3Err, sentry.LevelWarning)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Contribution deleted successfully",
	})
}
