package admin

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
	"id-100/internal/utils"
)

// AdminDeleteContributionHandler deletes a contribution from the admin panel
func AdminDeleteContributionHandler(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.Atoi(contributionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}

	// Get the image URL and token_id before deletion
	imageURL, tokenID, err := repository.GetContributionForAdminDeletion(context.Background(), contributionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Delete from upload_logs first (foreign key reference)
	err = repository.DeleteUploadLog(context.Background(), contributionID)
	if err != nil {
		log.Printf("Failed to delete from upload_logs: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete upload log"})
	}

	// Delete from contributions table
	rowsAffected, err := repository.DeleteContribution(context.Background(), contributionID)
	if err != nil {
		log.Printf("Failed to delete contribution: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contribution"})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Decrement the total_uploads counter for the token
	err = repository.DecrementTokenUploadCount(context.Background(), tokenID)
	if err != nil {
		log.Printf("Failed to decrement upload counter: %v", err)
	}

	// Delete from S3 storage if the image exists
	if imageURL != "" {
		s3Err := utils.DeleteFromS3(imageURL)
		if s3Err != nil {
			log.Printf("Failed to delete from S3 (continuing anyway): %v", s3Err)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Contribution deleted successfully",
	})
}
