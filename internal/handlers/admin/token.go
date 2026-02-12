package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
	"id-100/internal/utils"
)

// AdminTokenResetHandler resets a token for the next player
func AdminTokenResetHandler(c echo.Context) error {
	tokenID := c.Param("id")

	rows, err := repository.ResetToken(context.Background(), tokenID)
	if err != nil {
		log.Printf("Database error in AdminTokenResetHandler: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.CaptureException(err)
		}
		return c.String(http.StatusInternalServerError, "Database error")
	}

	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Werkzeug wurde zurückgesetzt und kann an den nächsten Spieler weitergegeben werden",
	})
}

// AdminTokenDeactivateHandler deactivates a token
func AdminTokenDeactivateHandler(c echo.Context) error {
	tokenID := c.Param("id")

	rows, err := repository.DeactivateToken(context.Background(), tokenID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

// AdminTokenAssignHandler assigns a token to a specific player
func AdminTokenAssignHandler(c echo.Context) error {
	tokenID := c.Param("id")

	type AssignRequest struct {
		PlayerName string `json:"player_name"`
	}

	var req AssignRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.PlayerName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "player_name is required",
		})
	}

	rows, err := repository.AssignTokenToPlayer(context.Background(), tokenID, req.PlayerName)
	if err != nil {
		log.Printf("Database error in AdminTokenAssignHandler: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.CaptureException(err)
		}
		return c.String(http.StatusInternalServerError, "Database error")
	}

	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Werkzeug wurde an %s vergeben", req.PlayerName),
	})
}

// AdminTokenListHandler returns JSON list of all tokens
func AdminTokenListHandler(c echo.Context) error {
	tokens, err := repository.GetAllTokens(context.Background())
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	})
}

// AdminCreateTokenHandler creates a new token/bag
func AdminCreateTokenHandler(c echo.Context, baseURL string) error {
	type CreateRequest struct {
		BagName    string `json:"bag_name"`
		MaxUploads int    `json:"max_uploads"`
	}

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.BagName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "bag_name is required",
		})
	}

	if req.MaxUploads <= 0 {
		req.MaxUploads = 100 // Default
	}

	// Generate secure token
	token, err := utils.GenerateSecureToken(40)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.CaptureException(err)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Insert into database
	tokenID, err := repository.CreateToken(context.Background(), token, req.BagName, req.MaxUploads)
	if err != nil {
		log.Printf("Failed to create token: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.CaptureException(err)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Generate upload URL
	uploadURL := fmt.Sprintf("%s/upload?token=%s", baseURL, token)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "success",
		"token_id":   tokenID,
		"token":      token,
		"bag_name":   req.BagName,
		"upload_url": uploadURL,
		"qr_url":     fmt.Sprintf("%s/admin/tokens/%d/qr", baseURL, tokenID),
	})
}

// AdminUpdateQuotaHandler updates the max_uploads quota for a token
func AdminUpdateQuotaHandler(c echo.Context) error {
	tokenID := c.Param("id")

	type QuotaRequest struct {
		MaxUploads int `json:"max_uploads"`
	}

	var req QuotaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.MaxUploads <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "max_uploads must be greater than 0",
		})
	}

	rows, err := repository.UpdateTokenQuota(context.Background(), tokenID, req.MaxUploads)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "success",
		"max_uploads": req.MaxUploads,
	})
}
