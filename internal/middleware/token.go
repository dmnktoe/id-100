package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"id-100/internal/database"
)

const (
	// UploadCooldownDuration is the minimum time between uploads for the same token
	UploadCooldownDuration = 5 * time.Second
)

// TokenWithSession is a middleware with session support for token validation
func TokenWithSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		session, err := Store.Get(c.Request(), "id-100-session")
		if err != nil {
			log.Printf("Session error (creating new session): %v", err)
			// Create a new session when old session can't be decoded
			// This handles cases where old cookies exist from before gob.Register fix
			session, err = Store.New(c.Request(), "id-100-session")
			if err != nil {
				log.Printf("Failed to create new session: %v", err)
				return c.String(http.StatusInternalServerError, "Session initialization failed")
			}
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
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "Zugang verweigert",
				"ContentTemplate": "access_denied.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		// Validate token
		var tokenID int
		var isActive bool
		var maxUploads, totalUploads, totalSessions int
		var currentPlayer, currentPlayerCity, bagName string
		var sessionStartedAt time.Time
		var dbSessionUUID *string // Session UUID bound to this token

		err = database.DB.QueryRow(context.Background(),
			`SELECT id, is_active, max_uploads, total_uploads, total_sessions,
			 COALESCE(current_player, ''), COALESCE(current_player_city, ''), COALESCE(bag_name, ''), COALESCE(session_started_at, created_at),
			 session_uuid
			 FROM upload_tokens WHERE token = $1`,
			token).Scan(&tokenID, &isActive, &maxUploads, &totalUploads, &totalSessions, &currentPlayer, &currentPlayerCity, &bagName, &sessionStartedAt, &dbSessionUUID)

		if err != nil {
			log.Printf("Token validation error: %v", err)
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "Ungültiger Token",
				"ContentTemplate": "invalid_token.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		// Get or create session UUID for this browser
		sessionUUID, err := GetOrCreateSessionUUID(session)
		if err != nil {
			log.Printf("Failed to create session UUID: %v", err)
			return c.String(http.StatusInternalServerError, "Session initialization failed")
		}

		// Save token in session for subsequent requests
		session.Values["token"] = token
		session.Values["token_id"] = tokenID
		session.Values["bag_name"] = bagName
		session.Values["session_uuid"] = sessionUUID

		// Generate CSRF token for this session (needed for forms)
		csrfToken, err := GetOrCreateCSRFToken(session)
		if err != nil {
			log.Printf("Failed to create CSRF token: %v", err)
			csrfToken = ""
		}

		// session freshness: ensure session_number and session_started_at exist and match DB
		sessNumVal := session.Values["session_number"]
		if existing, ok := GetSessionNumber(sessNumVal); ok {
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
		if existingStart, ok := GetSessionTime(sessStartVal); ok {
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
				// Set session_uuid in context so handler can use it
				c.Set("session_uuid", sessionUUID)
				// Save session before passing to handler so session_uuid is in cookie
				if err := session.Save(c.Request(), c.Response()); err != nil {
					log.Printf("Failed to save session before handler: %v", err)
					// Continue anyway as handler will try to save again
				}
				return next(c)
			}

			// Check if name is in session
			if sessName, ok := session.Values["player_name"].(string); ok && sessName != "" {
				// Update DB with name from session and bind session_uuid
				result, err := database.DB.Exec(context.Background(),
					"UPDATE upload_tokens SET current_player = $1, session_started_at = NOW(), session_uuid = $2 WHERE id = $3",
					sessName, sessionUUID, tokenID)

				if err != nil {
					log.Printf("Failed to update current_player for token_id=%d with name=%s: %v", tokenID, sessName, err)
					// Don't fail the request, but keep currentPlayer empty so name form shows again
					return c.Render(http.StatusOK, "layout", mergeTemplateData(map[string]interface{}{
						"Title":           "Willkommen",
						"ContentTemplate": "enter_name.content",
						"CurrentPath":     c.Request().URL.Path,
						"CurrentYear":     time.Now().Year(),
						"BagName":         bagName,
						"Token":           token,
						"CSRFToken":       csrfToken,
					}))
				}

				rows := result.RowsAffected()
				if rows == 0 {
					log.Printf("No rows updated when setting current_player for token_id=%d", tokenID)
				}

				// Only set currentPlayer after successful DB update
				currentPlayer = sessName
			} else {
				// Show name entry form
				return c.Render(http.StatusOK, "layout", mergeTemplateData(map[string]interface{}{
					"Title":           "Willkommen",
					"ContentTemplate": "enter_name.content",
					"CurrentPath":     c.Request().URL.Path,
					"CurrentYear":     time.Now().Year(),
					"BagName":         bagName,
					"Token":           token,
					"CSRFToken":       csrfToken,
				}))
			}
		} else {
			// currentPlayer is set - check for session conflict
			// Only check if another browser is using this token when player is already set
			if dbSessionUUID != nil && *dbSessionUUID != "" && *dbSessionUUID != sessionUUID {
				// Token is already bound to a different session
				// Check if this session has an active invitation
				var invitationExists bool
				err = database.DB.QueryRow(context.Background(),
					`SELECT EXISTS(
						SELECT 1 FROM invitations 
						WHERE token_id = $1 
						AND (accepted_by_session_uuid = $2 OR created_by_session_uuid = $2)
						AND is_active = true
						AND expires_at > NOW()
					)`,
					tokenID, sessionUUID).Scan(&invitationExists)

				if err != nil {
					log.Printf("Failed to check invitation: %v", err)
				}

				if !invitationExists {
					// Return 409 Conflict - bag is in use by another device
					return c.Render(http.StatusConflict, "layout", mergeTemplateData(map[string]interface{}{
						"Title":           "Werkzeug wird bereits verwendet",
						"ContentTemplate": "bag_in_use.content",
						"CurrentPath":     c.Request().URL.Path,
						"CurrentYear":     time.Now().Year(),
						"BagName":         bagName,
						"CurrentPlayer":   currentPlayer,
					}))
				}
			}

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
		c.Set("session_uuid", sessionUUID)
		c.Set("csrf_token", csrfToken)
		session.Save(c.Request(), c.Response())

		// Check if token is active
		if !isActive {
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "Token deaktiviert",
				"ContentTemplate": "token_deactivated.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		// Check upload limit
		if totalUploads >= maxUploads {
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "Upload-Limit erreicht",
				"ContentTemplate": "limit_reached.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
				"TotalUploads":    totalUploads,
				"MaxUploads":      maxUploads,
			}))
		}

		// For POST requests: Check cooldown
		if c.Request().Method == "POST" {
			var lastUpload *time.Time
			err = database.DB.QueryRow(context.Background(),
				"SELECT MAX(uploaded_at) FROM upload_logs WHERE token_id = $1 AND session_number = $2",
				tokenID, totalSessions).Scan(&lastUpload)

			if err == nil && lastUpload != nil {
				timeSince := time.Since(*lastUpload)

				if timeSince < UploadCooldownDuration {
					remainingSeconds := int(UploadCooldownDuration.Seconds() - timeSince.Seconds())
					return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
						"error":             "Bitte warte zwischen Uploads",
						"remaining_seconds": remainingSeconds,
					})
				}
			}
		}

		return next(c)
	}
}
