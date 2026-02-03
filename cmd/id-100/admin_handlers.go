package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html"
	"log"
	"math"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
)

// setPlayerNameHandler handles the name entry form submission
func setPlayerNameHandler(c echo.Context) error {
	// Protect against large request bodies before parsing form values
	const maxFormSize = int64(2 * 1024 * 1024) // 2 MiB
	if strings.Contains(c.Request().Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxFormSize)
	}

	playerName := c.FormValue("player_name")
	token := c.FormValue("token")

	if playerName == "" || token == "" {
		return c.String(http.StatusBadRequest, "Name und Token erforderlich")
	}

	// Consent checkbox (required)
	consent := c.FormValue("agree_privacy")
	if consent == "" {
		// try to fetch bag name for nicer rendering
		var bagName string
		_ = db.QueryRow(context.Background(), "SELECT COALESCE(bag_name,'') FROM upload_tokens WHERE token = $1", token).Scan(&bagName)
		return c.Render(http.StatusBadRequest, "layout", map[string]interface{}{
			"Title":           "Willkommen bei ID-100!",
			"ContentTemplate": "enter_name.content",
			"Token":           token,
			"BagName":         bagName,
			"FormError":       "Bitte bestätige die Datenschutzerklärung und dass du keine erkennbaren Personen ohne Einwilligung hochlädst.",
		})
	}

	playerCity := strings.TrimSpace(c.FormValue("player_city"))

	// Save name and city in session
	session, _ := store.Get(c.Request(), "id-100-session")
	session.Values["player_name"] = playerName
	session.Values["player_city"] = playerCity
	session.Save(c.Request(), c.Response())

	// Update database with city
	_, err := db.Exec(context.Background(),
		"UPDATE upload_tokens SET current_player = $1, current_player_city = $2, session_started_at = NOW() WHERE token = $3",
		playerName, playerCity, token)

	if err != nil {
		log.Printf("Error setting player name: %v", err)
	}

	// Redirect to upload page
	return c.Redirect(http.StatusSeeOther, "/upload?token="+token)
}

// adminDashboardHandler shows the admin dashboard
func adminDashboardHandler(c echo.Context) error {
	// Get all tokens
	rows, err := db.Query(context.Background(), `
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

	type TokenInfo struct {
		ID                int
		Token             string
		BagName           string
		CurrentPlayer     string
		CurrentPlayerCity string
		IsActive          bool
		MaxUploads        int
		TotalUploads      int
		TotalSessions     int
		SessionStartedAt  time.Time
		CreatedAt         time.Time
		Remaining         int
	}

	var tokens []TokenInfo
	for rows.Next() {
		var t TokenInfo
		if err := rows.Scan(&t.ID, &t.Token, &t.BagName, &t.CurrentPlayer, &t.CurrentPlayerCity, &t.IsActive,
			&t.MaxUploads, &t.TotalUploads, &t.TotalSessions, &t.SessionStartedAt, &t.CreatedAt); err != nil {
			continue
		}
		t.Remaining = t.MaxUploads - t.TotalUploads
		tokens = append(tokens, t)
	}

	// Get recent contributions
	type RecentContrib struct {
		ImageUrl     string
		PlayerName   string
		DeriveNumber int
	}

	contribRows, err := db.Query(context.Background(), `
		SELECT c.image_url, COALESCE(ul.player_name, 'Anonym'), ul.derive_number
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		ORDER BY c.created_at DESC
		LIMIT 20
	`)
	if err != nil {
		log.Printf("Failed to fetch recent contributions: %v", err)
		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "Admin Dashboard",
			"ContentTemplate": "admin_dashboard.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"Tokens":          tokens,
			"RecentContribs":  []RecentContrib{},
		})
	}
	defer contribRows.Close()

	var recentContribs []RecentContrib
	for contribRows.Next() {
		var rc RecentContrib
		if err := contribRows.Scan(&rc.ImageUrl, &rc.PlayerName, &rc.DeriveNumber); err != nil {
			continue
		}
		rc.ImageUrl = ensureFullImageURL(rc.ImageUrl)
		recentContribs = append(recentContribs, rc)
	}

	// Fetch bag requests (with optional status filter)
	type BagRequest struct {
		ID        int
		Email     string
		CreatedAt time.Time
		Handled   bool
	}

	status := c.QueryParam("bag_status")
	// selected tab (server-side)
	tab := c.QueryParam("tab")
	if tab != "tokens" && tab != "requests" && tab != "contribs" {
		tab = "tokens"
	}

	// counts for filter badges
	var openCount, handledCount int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM bag_requests WHERE handled = FALSE").Scan(&openCount); err != nil {
		openCount = 0
	}
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM bag_requests WHERE handled = TRUE").Scan(&handledCount); err != nil {
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

	reqRows, err := db.Query(context.Background(), query)
	var bagRequests []BagRequest
	if err != nil {
		log.Printf("Failed to fetch bag requests: %v", err)
	} else {
		defer reqRows.Close()
		for reqRows.Next() {
			var br BagRequest
			if err := reqRows.Scan(&br.ID, &br.Email, &br.CreatedAt, &br.Handled); err == nil {
				bagRequests = append(bagRequests, br)
			}
		}
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Admin Dashboard",
		"ContentTemplate": "admin_dashboard.content",
		"Tokens":          tokens,
		"RecentContribs":  recentContribs,
		"BagRequests":     bagRequests,
		"BagStatus":       status,
		"OpenCount":       openCount,
		"HandledCount":    handledCount,
		"Tab":             tab,
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
	})

}

// POST /admin/bag-requests/:id/complete
func adminBagRequestCompleteHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	res, err := db.Exec(context.Background(), "UPDATE bag_requests SET handled = TRUE WHERE id = $1", id)
	if err != nil {
		log.Printf("Failed to mark bag_request handled: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	if res.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Session helpers for middleware
func getSessionNumber(v interface{}) (int, bool) {
	// Determine platform int bounds using strconv.IntSize
	bits := strconv.IntSize
	maxInt64 := int64(1<<(bits-1) - 1)
	minInt64 := -int64(1 << (bits - 1))

	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		if n >= minInt64 && n <= maxInt64 {
			return int(n), true
		}
		return 0, false
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, false
		}
		if math.Trunc(n) != n {
			return 0, false
		}
		// Ensure the float value fits in the platform int range before converting
		if n < float64(minInt64) || n > float64(maxInt64) {
			return 0, false
		}
		asInt64 := int64(n)
		return int(asInt64), true
	case string:
		if x, err := strconv.Atoi(n); err == nil {
			return x, true
		}
	}
	return 0, false
}

func getSessionTime(v interface{}) (time.Time, bool) {
	// Compute safe Unix second bounds such that sec*1e9 doesn't overflow int64
	maxInt64 := int64(^uint64(0) >> 1)
	minInt64 := -maxInt64 - 1
	maxSec := maxInt64 / 1e9
	minSec := minInt64 / 1e9

	switch t := v.(type) {
	case time.Time:
		return t, true
	case string:
		if tm, err := time.Parse(time.RFC3339, t); err == nil {
			return tm, true
		}
	case int64:
		if t >= minSec && t <= maxSec {
			return time.Unix(t, 0), true
		}
		return time.Time{}, false
	case int:
		sec := int64(t)
		if sec >= minSec && sec <= maxSec {
			return time.Unix(sec, 0), true
		}
		return time.Time{}, false
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return time.Time{}, false
		}
		if math.Trunc(t) != t {
			return time.Time{}, false
		}
		sec := int64(t)
		if sec >= minSec && sec <= maxSec {
			return time.Unix(sec, 0), true
		}
		return time.Time{}, false
	}
	return time.Time{}, false
}

// tokenMiddlewareWithSession is an updated middleware with session support
func tokenMiddlewareWithSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		session, err := store.Get(c.Request(), "id-100-session")
		if err != nil {
			log.Printf("Session error: %v", err)
		}

		// Get token from query param (QR code), POST form (only for form-encoded or explicit routes), or session
		token := c.QueryParam("token")
		// If missing, for POST requests consider parsing form-encoded bodies only. Do NOT parse multipart uploads here.
		if token == "" && c.Request().Method == "POST" {
			const maxFormSize = int64(2 * 1024 * 1024) // 2 MiB
			contentType := c.Request().Header.Get("Content-Type")
			isFormEncoded := strings.Contains(contentType, "application/x-www-form-urlencoded")
			// Allow explicit route(s) that accept form-encoded tokens (e.g., the upload set-name endpoint)
			if !isFormEncoded && c.Request().URL.Path == "/upload/set-name" {
				isFormEncoded = true
			}
			if isFormEncoded {
				// Limit the body size before any parsing to guard against large uploads
				c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxFormSize)
				if formToken := c.FormValue("token"); formToken != "" {
					token = formToken
				}
			}
		}
		if token == "" {
			// Try to get from session
			if sessToken, ok := session.Values["token"].(string); ok {
				token = sessToken
			}
		}

		if token == "" {
			return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
				"Title":           "Zugang verweigert",
				"ContentTemplate": "access_denied.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			})
		}

		// Validate token
		var tokenID int
		var isActive bool
		var maxUploads, totalUploads, totalSessions int
		var currentPlayer, currentPlayerCity, bagName string
		var sessionStartedAt time.Time

		err = db.QueryRow(context.Background(),
			`SELECT id, is_active, max_uploads, total_uploads, total_sessions,
			 COALESCE(current_player, ''), COALESCE(current_player_city, ''), COALESCE(bag_name, ''), COALESCE(session_started_at, created_at)
			 FROM upload_tokens WHERE token = $1`,
			token).Scan(&tokenID, &isActive, &maxUploads, &totalUploads, &totalSessions, &currentPlayer, &currentPlayerCity, &bagName, &sessionStartedAt)

		if err != nil {
			log.Printf("Token validation error: %v", err)
			return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
				"Title":           "Ungültiger Token",
				"ContentTemplate": "invalid_token.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			})
		}

		// Save token in session for subsequent requests
		session.Values["token"] = token
		session.Values["token_id"] = tokenID
		session.Values["bag_name"] = bagName

		// session freshness: ensure session_number and session_started_at exist and match DB
		sessNumVal := session.Values["session_number"]
		if existing, ok := getSessionNumber(sessNumVal); ok {
			if existing != totalSessions {
				// session is stale: clear the stored player name and update stored session meta
				delete(session.Values, "player_name")
				session.Values["session_number"] = totalSessions
				session.Values["session_started_at"] = sessionStartedAt
				// force name flow
				currentPlayer = ""
			}
		} else {
			// initialize session metadata for this token so future admin resets can be detected
			session.Values["session_number"] = totalSessions
			session.Values["session_started_at"] = sessionStartedAt
		}

		// Additionally check session_started_at mismatch (covers manual DB edits without incrementing total_sessions)
		sessStartVal := session.Values["session_started_at"]
		if existingStart, ok := getSessionTime(sessStartVal); ok {
			if !existingStart.Equal(sessionStartedAt) {
				// session is stale: clear player_name and update stored session meta
				delete(session.Values, "player_name")
				session.Values["session_started_at"] = sessionStartedAt
				session.Values["session_number"] = totalSessions
				currentPlayer = ""
			}
		} else {
			// ensure the session has the start time
			session.Values["session_started_at"] = sessionStartedAt
		}

		session.Save(c.Request(), c.Response())

		// If DB currently has no current_player (e.g. after an admin reset), remove any stored player_name from the session
		if currentPlayer == "" {
			if _, ok := session.Values["player_name"].(string); ok {
				delete(session.Values, "player_name")
				// persist session changes
				session.Save(c.Request(), c.Response())
			}
		}

		// Check if player name is set (first-time user flow)
		if currentPlayer == "" {
			// If this is a POST to /upload/set-name, let the handler process it
			if c.Request().Method == "POST" && c.Request().URL.Path == "/upload/set-name" {
				return next(c)
			}

			// Check if name is in session
			if sessName, ok := session.Values["player_name"].(string); ok && sessName != "" {
				// Update DB with name from session
				result, err := db.Exec(context.Background(),
					"UPDATE upload_tokens SET current_player = $1, session_started_at = NOW() WHERE id = $2",
					sessName, tokenID)

				if err != nil {
					log.Printf("Failed to update current_player for token_id=%d with name=%s: %v", tokenID, sessName, err)
					// Don't fail the request, but keep currentPlayer empty so name form shows again
					return c.Render(http.StatusOK, "layout", map[string]interface{}{
						"Title":           "Willkommen",
						"ContentTemplate": "enter_name.content",
						"CurrentPath":     c.Request().URL.Path,
						"CurrentYear":     time.Now().Year(),
						"BagName":         bagName,
						"Token":           token,
					})
				}

				rows := result.RowsAffected()
				if rows == 0 {
					log.Printf("No rows updated when setting current_player for token_id=%d", tokenID)
				}

				// Only set currentPlayer after successful DB update
				currentPlayer = sessName
			} else {
				// Show name entry form
				return c.Render(http.StatusOK, "layout", map[string]interface{}{
					"Title":           "Willkommen",
					"ContentTemplate": "enter_name.content",
					"CurrentPath":     c.Request().URL.Path,
					"CurrentYear":     time.Now().Year(),
					"BagName":         bagName,
					"Token":           token,
				})
			}
		} else {
			// Save player name and city in session if not already there
			session.Values["player_name"] = currentPlayer
			session.Values["player_city"] = currentPlayerCity
			session.Save(c.Request(), c.Response())
		}

		// Store token info in context for handler
		c.Set("token_id", tokenID)
		c.Set("token", token)
		c.Set("current_player", currentPlayer)
		c.Set("bag_name", bagName)
		c.Set("session_number", totalSessions)
		c.Set("uploads_remaining", maxUploads-totalUploads)
		c.Set("current_player_city", currentPlayerCity)

		// Check if token is active
		if !isActive {
			return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
				"Title":           "Token deaktiviert",
				"ContentTemplate": "token_deactivated.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			})
		}

		// Check upload limit
		if totalUploads >= maxUploads {
			return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
				"Title":           "Upload-Limit erreicht",
				"ContentTemplate": "limit_reached.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
				"TotalUploads":    totalUploads,
				"MaxUploads":      maxUploads,
			})
		}

		// For POST requests: Check cooldown (5 seconds)
		if c.Request().Method == "POST" {
			var lastUpload *time.Time
			err = db.QueryRow(context.Background(),
				"SELECT MAX(uploaded_at) FROM upload_logs WHERE token_id = $1 AND session_number = $2",
				tokenID, totalSessions).Scan(&lastUpload)

			if err == nil && lastUpload != nil {
				timeSince := time.Since(*lastUpload)
				cooldownDuration := 5 * time.Second

				if timeSince < cooldownDuration {
					remainingSeconds := int(cooldownDuration.Seconds() - timeSince.Seconds())
					return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
						"error":             "Bitte warte zwischen Uploads",
						"remaining_seconds": remainingSeconds,
					})
				}
			}
		}

		// Store token info in context for handler
		c.Set("token_id", tokenID)
		c.Set("token", token)
		c.Set("current_player", currentPlayer)
		c.Set("bag_name", bagName)
		c.Set("session_number", totalSessions)
		c.Set("uploads_remaining", maxUploads-totalUploads)

		return next(c)
	}
}

// adminTokenResetHandler resets a token for the next player
func adminTokenResetHandler(c echo.Context) error {
	tokenID := c.Param("id")

	result, err := db.Exec(context.Background(),
		`UPDATE upload_tokens 
		 SET total_uploads = 0, 
		     total_sessions = total_sessions + 1,
		     session_started_at = NOW(),
		     current_player = NULL,
		     is_active = true
		 WHERE id = $1`,
		tokenID)

	if err != nil {
		log.Printf("Database error in adminTokenResetHandler: %v", err)
		return c.String(http.StatusInternalServerError, "Database error")
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Tasche wurde zurückgesetzt und kann an den nächsten Spieler weitergegeben werden",
	})
}

// adminTokenDeactivateHandler deactivates a token
func adminTokenDeactivateHandler(c echo.Context) error {
	tokenID := c.Param("id")

	result, err := db.Exec(context.Background(),
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

// adminTokenAssignHandler assigns a token to a specific player
func adminTokenAssignHandler(c echo.Context) error {
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

	result, err := db.Exec(context.Background(),
		`UPDATE upload_tokens 
		 SET current_player = $1,
		     session_started_at = NOW(),
		     is_active = true
		 WHERE id = $2`,
		req.PlayerName, tokenID)

	if err != nil {
		log.Printf("Database error in adminTokenAssignHandler: %v", err)
		return c.String(http.StatusInternalServerError, "Database error")
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return c.String(http.StatusNotFound, "Token not found")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Tasche wurde an %s vergeben", req.PlayerName),
	})
}

// adminTokenListHandler returns JSON list of all tokens
func adminTokenListHandler(c echo.Context) error {
	rows, err := db.Query(context.Background(), `
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

	type TokenInfo struct {
		ID                int       `json:"id"`
		Token             string    `json:"token"`
		BagName           string    `json:"bag_name"`
		CurrentPlayer     string    `json:"current_player"`
		CurrentPlayerCity string    `json:"current_player_city"`
		IsActive          bool      `json:"is_active"`
		MaxUploads        int       `json:"max_uploads"`
		TotalUploads      int       `json:"total_uploads"`
		TotalSessions     int       `json:"total_sessions"`
		SessionStartedAt  time.Time `json:"session_started_at"`
		CreatedAt         time.Time `json:"created_at"`
		Remaining         int       `json:"remaining"`
	}

	var tokens []TokenInfo
	for rows.Next() {
		var t TokenInfo
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

// basicAuthMiddleware protects admin endpoints
func basicAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		adminUser := os.Getenv("ADMIN_USERNAME")
		adminPass := os.Getenv("ADMIN_PASSWORD")

		if adminUser == "" || adminPass == "" {
			log.Printf("ADMIN_USERNAME or ADMIN_PASSWORD not set")
			return c.String(http.StatusInternalServerError, "Server misconfiguration")
		}

		username, password, ok := c.Request().BasicAuth()
		if !ok {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		// Use constant-time comparison to prevent timing attacks
		userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(adminUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(adminPass)) == 1

		if !userMatch || !passMatch {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}
		return next(c)
	}
}

// adminCreateTokenHandler creates a new token/bag
func adminCreateTokenHandler(c echo.Context) error {
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
	token, err := generateSecureToken(40)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Insert into database
	var tokenID int
	err = db.QueryRow(context.Background(),
		`INSERT INTO upload_tokens (token, bag_name, max_uploads, total_sessions) 
		 VALUES ($1, $2, $3, 1) RETURNING id`,
		token, req.BagName, req.MaxUploads).Scan(&tokenID)

	if err != nil {
		log.Printf("Failed to create token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Generate upload URL (use global baseURL from env)
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

// adminDownloadQRHandler generates and returns QR code as SVG
func adminDownloadQRHandler(c echo.Context) error {
	tokenID := c.Param("id")

	// Get token from database
	var token, bagName string
	err := db.QueryRow(context.Background(),
		"SELECT token, COALESCE(bag_name, '') FROM upload_tokens WHERE id = $1",
		tokenID).Scan(&token, &bagName)

	if err != nil {
		return c.String(http.StatusNotFound, "Token not found")
	}

	// Generate upload URL (use global baseURL from env)
	uploadURL := fmt.Sprintf("%s/upload?token=%s", baseURL, token)

	// Check format parameter
	format := c.QueryParam("format")
	if format == "" {
		format = "png" // default
	}

	switch format {
	case "svg":
		// Generate SVG QR code using custom SVG generation
		svg := generateQRCodeSVG(uploadURL, bagName)
		c.Response().Header().Set("Content-Type", "image/svg+xml")
		// Use mime.FormatMediaType to safely encode filename
		filename := fmt.Sprintf("qr_%s.svg", sanitizeFilename(bagName))
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
		// Use sanitizeFilename to prevent header injection
		filename := fmt.Sprintf("qr_%s.png", sanitizeFilename(bagName))
		c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return c.Blob(http.StatusOK, "image/png", pngBytes)

	default:
		return c.String(http.StatusBadRequest, "Invalid format. Use 'svg' or 'png'")
	}
}

// adminUpdateQuotaHandler updates the max_uploads quota for a token
func adminUpdateQuotaHandler(c echo.Context) error {
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

	result, err := db.Exec(context.Background(),
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

// generateSecureToken generates a cryptographically secure token
func generateSecureToken(length int) (string, error) {
	// Calculate bytes needed to get desired length after base64 encoding
	// base64 encoding produces 4 chars for every 3 bytes
	bytesNeeded := (length*3 + 3) / 4
	b := make([]byte, bytesNeeded)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	if len(encoded) > length {
		return encoded[:length], nil
	}
	return encoded, nil
}

// generateQRCodeSVG generates a simple SVG QR code with label
func generateQRCodeSVG(url, label string) string {
	// Generate QR code
	qr, _ := qrcode.New(url, qrcode.High)

	// Get bitmap
	bitmap := qr.Bitmap()
	size := len(bitmap)
	scale := 10 // pixels per module
	padding := 40
	labelHeight := 60

	svgWidth := size*scale + 2*padding
	svgHeight := size*scale + 2*padding + labelHeight

	// Escape label to prevent XSS
	escapedLabel := html.EscapeString(label)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
	<rect width="%d" height="%d" fill="white"/>
	<text x="%d" y="30" text-anchor="middle" font-family="Arial" font-size="24" font-weight="bold" fill="black">%s</text>
	<text x="%d" y="%d" text-anchor="middle" font-family="Arial" font-size="16" fill="#666">Scanne für Upload</text>
	<g transform="translate(%d, %d)">`,
		svgWidth, svgHeight, svgWidth, svgHeight,
		svgWidth, svgHeight,
		svgWidth/2, escapedLabel,
		svgWidth/2, svgHeight-20,
		padding, padding+40)

	// Draw QR modules
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if bitmap[y][x] {
				svg += fmt.Sprintf(`
		<rect x="%d" y="%d" width="%d" height="%d" fill="black"/>`,
					x*scale, y*scale, scale, scale)
			}
		}
	}

	svg += `
	</g>
</svg>`

	return svg
}

// sanitizeFilename removes characters that could cause header injection
func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, name)
}
