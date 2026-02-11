package app

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
	"id-100/internal/utils"
)

// UserDeleteContributionHandler allows users to delete their own contributions from the current session
func UserDeleteContributionHandler(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.Atoi(contributionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}

	// Get token info from middleware context
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Token not found"})
	}

	sessionNumber, _ := c.Get("session_number").(int)

	// Verify that this contribution belongs to the current user's session
	imageURL, err := repository.GetContributionForDeletion(context.Background(), contributionID, tokenID, sessionNumber)
	if err != nil {
		log.Printf("Failed to fetch contribution or ownership mismatch: %v", err)
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You can only delete your own uploads from this session"})
	}

	// Delete from upload_logs first
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

	// Decrement the total_uploads counter
	err = repository.DecrementTokenUploadCount(context.Background(), tokenID)
	if err != nil {
		log.Printf("Failed to decrement upload counter: %v", err)
	}

	// Delete from S3 storage
	if imageURL != "" {
		s3Err := utils.DeleteFromS3(imageURL)
		if s3Err != nil {
			log.Printf("Failed to delete from S3 (continuing anyway): %v", s3Err)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Upload deleted successfully",
	})
}
