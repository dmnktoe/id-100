package handlers

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"

	"id-100/internal/database"
	"id-100/internal/models"
	"id-100/internal/utils"
)

// AdminDashboardHandler shows the admin dashboard
func AdminDashboardHandler(c echo.Context) error {
	// Get all tokens
	rows, err := database.DB.Query(context.Background(), `
		SELECT id, token, COALESCE(bag_name, ''), COALESCE(current_player, ''), COALESCE(current_player_city, ''),
		       is_active, max_uploads, total_uploads, total_sessions,
		       COALESCE(session_started_at, created_at), created_at
		FROM upload_tokens
		ORDER BY id ASC
	`)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var tokens []models.TokenInfo
	for rows.Next() {
		var t models.TokenInfo
		if err := rows.Scan(&t.ID, &t.Token, &t.BagName, &t.CurrentPlayer, &t.CurrentPlayerCity, &t.IsActive,
			&t.MaxUploads, &t.TotalUploads, &t.TotalSessions, &t.SessionStartedAt, &t.CreatedAt); err != nil {
			continue
		}
		t.Remaining = t.MaxUploads - t.TotalUploads
		tokens = append(tokens, t)
	}

	// Get recent contributions
	contribRows, err := database.DB.Query(context.Background(), `
		SELECT c.id, c.image_url, COALESCE(ul.player_name, 'Anonym'), ul.derive_number
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		ORDER BY c.created_at DESC
		LIMIT 20
	`)
	if err != nil {
		log.Printf("Failed to fetch recent contributions: %v", err)
		return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
			"Title":           "Admin Dashboard",
			"ContentTemplate": "admin_dashboard.content",
			"AdditionalCSS":   "admin.styles.css",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"Tokens":          tokens,
			"RecentContribs":  []models.RecentContrib{},
		}))
	}
	defer contribRows.Close()

	var recentContribs []models.RecentContrib
	for contribRows.Next() {
		var rc models.RecentContrib
		if err := contribRows.Scan(&rc.ID, &rc.ImageUrl, &rc.PlayerName, &rc.DeriveNumber); err != nil {
			continue
		}
		rc.ImageUrl = utils.EnsureFullImageURL(rc.ImageUrl)
		recentContribs = append(recentContribs, rc)
	}

	// Fetch bag requests (with optional status filter)
	status := c.QueryParam("bag_status")
	// selected tab (server-side)
	tab := c.QueryParam("tab")
	if tab != "tokens" && tab != "requests" && tab != "contribs" {
		tab = "tokens"
	}

	// counts for filter badges
	var openCount, handledCount int
	if err := database.DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM bag_requests WHERE handled = FALSE").Scan(&openCount); err != nil {
		openCount = 0
	}
	if err := database.DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM bag_requests WHERE handled = TRUE").Scan(&handledCount); err != nil {
		handledCount = 0
	}

	var query string
	switch status {
	case "open":
		query = "SELECT id, email, created_at, handled FROM bag_requests WHERE handled = FALSE ORDER BY created_at DESC LIMIT 50"
	case "handled":
		query = "SELECT id, email, created_at, handled FROM bag_requests WHERE handled = TRUE ORDER BY created_at DESC LIMIT 50"
	default:
		query = "SELECT id, email, created_at, handled FROM bag_requests ORDER BY created_at DESC LIMIT 50"
	}

	reqRows, err := database.DB.Query(context.Background(), query)
	var bagRequests []models.BagRequest
	if err != nil {
		log.Printf("Failed to fetch bag requests: %v", err)
	} else {
		defer reqRows.Close()
		for reqRows.Next() {
			var br models.BagRequest
			if err := reqRows.Scan(&br.ID, &br.Email, &br.CreatedAt, &br.Handled); err == nil {
				bagRequests = append(bagRequests, br)
			}
		}
	}

	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Admin Dashboard",
		"ContentTemplate": "admin_dashboard.content",
		"AdditionalCSS":   "admin.styles.css",
		"Tokens":          tokens,
		"RecentContribs":  recentContribs,
		"BagRequests":     bagRequests,
		"BagStatus":       status,
		"OpenCount":       openCount,
		"HandledCount":    handledCount,
		"Tab":             tab,
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
	}))
}

// AdminBagRequestCompleteHandler marks a bag request as handled
func AdminBagRequestCompleteHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	res, err := database.DB.Exec(context.Background(), "UPDATE bag_requests SET handled = TRUE WHERE id = $1", id)
	if err != nil {
		log.Printf("Failed to mark bag_request handled: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	if res.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// AdminTokenResetHandler resets a token for the next player
func AdminTokenResetHandler(c echo.Context) error {
	tokenID := c.Param("id")

	result, err := database.DB.Exec(context.Background(),
		`UPDATE upload_tokens 
		 SET total_uploads = 0, 
		     total_sessions = total_sessions + 1,
		     session_started_at = NOW(),
		     current_player = NULL,
		     is_active = true
		 WHERE id = $1`,
		tokenID)

	if err != nil {
		log.Printf("Database error in AdminTokenResetHandler: %v", err)
		return c.String(http.StatusInternalServerError, "Database error")
	}

	rows := result.RowsAffected()
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

	result, err := database.DB.Exec(context.Background(),
		"UPDATE upload_tokens SET is_active = false WHERE id = $1",
		tokenID)

	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	rows := result.RowsAffected()
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

	result, err := database.DB.Exec(context.Background(),
		`UPDATE upload_tokens 
		 SET current_player = $1,
		     session_started_at = NOW(),
		     is_active = true
		 WHERE id = $2`,
		req.PlayerName, tokenID)

	if err != nil {
		log.Printf("Database error in AdminTokenAssignHandler: %v", err)
		return c.String(http.StatusInternalServerError, "Database error")
	}

	rows := result.RowsAffected()
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
	rows, err := database.DB.Query(context.Background(), `
		SELECT id, token, COALESCE(bag_name, ''), COALESCE(current_player, ''), COALESCE(current_player_city,''),
		       is_active, max_uploads, total_uploads, total_sessions,
		       COALESCE(session_started_at, created_at), created_at
		FROM upload_tokens
		ORDER BY id ASC
	`)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var tokens []models.TokenInfo
	for rows.Next() {
		var t models.TokenInfo
		if err := rows.Scan(&t.ID, &t.Token, &t.BagName, &t.CurrentPlayer, &t.CurrentPlayerCity, &t.IsActive,
			&t.MaxUploads, &t.TotalUploads, &t.TotalSessions, &t.SessionStartedAt, &t.CreatedAt); err != nil {
			continue
		}
		t.Remaining = t.MaxUploads - t.TotalUploads
		tokens = append(tokens, t)
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Insert into database
	var tokenID int
	err = database.DB.QueryRow(context.Background(),
		`INSERT INTO upload_tokens (token, bag_name, max_uploads, total_sessions) 
		 VALUES ($1, $2, $3, 1) RETURNING id`,
		token, req.BagName, req.MaxUploads).Scan(&tokenID)

	if err != nil {
		log.Printf("Failed to create token: %v", err)
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

// AdminDownloadQRHandler generates and returns QR code as SVG or PNG
func AdminDownloadQRHandler(c echo.Context, baseURL string) error {
	tokenID := c.Param("id")

	// Get token from database
	var token, bagName string
	err := database.DB.QueryRow(context.Background(),
		"SELECT token, COALESCE(bag_name, '') FROM upload_tokens WHERE id = $1",
		tokenID).Scan(&token, &bagName)

	if err != nil {
		return c.String(http.StatusNotFound, "Token not found")
	}

	// Generate upload URL
	uploadURL := fmt.Sprintf("%s/upload?token=%s", baseURL, token)

	// Check format parameter
	format := c.QueryParam("format")
	if format == "" {
		format = "png" // default
	}

	switch format {
	case "svg":
		// Generate SVG QR code using custom SVG generation
		svg := utils.GenerateQRCodeSVG(uploadURL, bagName)
		c.Response().Header().Set("Content-Type", "image/svg+xml")
		// Use mime.FormatMediaType to safely encode filename
		filename := fmt.Sprintf("qr_%s.svg", utils.SanitizeFilename(bagName))
		c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return c.String(http.StatusOK, svg)

	case "png":
		// Generate PNG QR code
		qr, err := qrcode.New(uploadURL, qrcode.High)
		if err != nil {
			return c.String(http.StatusInternalServerError, "QR generation failed")
		}

		pngBytes, err := qr.PNG(512)
		if err != nil {
			return c.String(http.StatusInternalServerError, "PNG generation failed")
		}

		c.Response().Header().Set("Content-Type", "image/png")
		// Use SanitizeFilename to prevent header injection
		filename := fmt.Sprintf("qr_%s.png", utils.SanitizeFilename(bagName))
		c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return c.Blob(http.StatusOK, "image/png", pngBytes)

	default:
		return c.String(http.StatusBadRequest, "Invalid format. Use 'svg' or 'png'")
	}
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

	result, err := database.DB.Exec(context.Background(),
		"UPDATE upload_tokens SET max_uploads = $1 WHERE id = $2",
		req.MaxUploads, tokenID)

	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "success",
		"max_uploads": req.MaxUploads,
	})
}

// AdminDeleteContributionHandler deletes a contribution from the admin panel
func AdminDeleteContributionHandler(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.Atoi(contributionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}

	// Get the image URL and token_id before deletion
	var imageURL string
	var tokenID int
	err = database.DB.QueryRow(context.Background(), `
		SELECT c.image_url, ul.token_id
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		WHERE c.id = $1
	`, contributionID).Scan(&imageURL, &tokenID)

	if err != nil {
		log.Printf("Failed to fetch contribution: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Delete from upload_logs first (foreign key reference)
	_, err = database.DB.Exec(context.Background(),
		"DELETE FROM upload_logs WHERE contribution_id = $1",
		contributionID)

	if err != nil {
		log.Printf("Failed to delete from upload_logs: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete upload log"})
	}

	// Delete from contributions table
	result, err := database.DB.Exec(context.Background(),
		"DELETE FROM contributions WHERE id = $1",
		contributionID)

	if err != nil {
		log.Printf("Failed to delete contribution: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contribution"})
	}

	if result.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Decrement the total_uploads counter for the token
	_, err = database.DB.Exec(context.Background(),
		"UPDATE upload_tokens SET total_uploads = total_uploads - 1 WHERE id = $1 AND total_uploads > 0",
		tokenID)

	if err != nil {
		log.Printf("Failed to decrement upload counter: %v", err)
		// Don't fail the request
	}

	// Delete from S3 storage if the image exists
	if imageURL != "" {
		s3Err := utils.DeleteFromS3(imageURL)
		if s3Err != nil {
			log.Printf("Failed to delete from S3 (continuing anyway): %v", s3Err)
			// Don't fail the request if S3 deletion fails
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Contribution deleted successfully",
	})
}
