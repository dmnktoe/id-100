package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"id-100/internal/database"
	"id-100/internal/utils"
)

const (
	// UploadCooldownDuration is the minimum time between uploads for the same token
	UploadCooldownDuration = 5 * time.Second
)

// TokenWithSession is a middleware with session support for token validation and conflict detection
func TokenWithSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		session, err := Store.Get(c.Request(), "id-100-session")
		if err != nil {
			log.Printf("Session error: %v", err)
		}

		// Get or generate session_uuid for this browser
		var sessionUUID string
		if uuid, ok := session.Values["session_uuid"].(string); ok && uuid != "" {
			sessionUUID = uuid
		} else {
			// Generate new session_uuid (44-char random string)
			sessionUUID, err = utils.GenerateSessionUUID()
			if err != nil {
				log.Printf("Failed to generate session UUID: %v", err)
				return c.String(http.StatusInternalServerError, "Server error")
			}
			session.Values["session_uuid"] = sessionUUID
			session.Save(c.Request(), c.Response())
			log.Printf("Generated new session UUID: %s", utils.MaskToken(sessionUUID))
		}

		// Make session_uuid available in context
		c.Set("session_uuid", sessionUUID)

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
		var currentPlayer, currentPlayerCity, bagName, boundSessionUUID string
		var sessionStartedAt time.Time

		err = database.DB.QueryRow(context.Background(),
			`SELECT id, is_active, max_uploads, total_uploads, total_sessions,
			 COALESCE(current_player, ''), COALESCE(current_player_city, ''), COALESCE(bag_name, ''), 
			 COALESCE(session_started_at, created_at), COALESCE(session_uuid, '')
			 FROM upload_tokens WHERE token = $1`,
			token).Scan(&tokenID, &isActive, &maxUploads, &totalUploads, &totalSessions, &currentPlayer, &currentPlayerCity, &bagName, &sessionStartedAt, &boundSessionUUID)

		if err != nil {
			log.Printf("Token validation error for token %s: %v", utils.MaskToken(token), err)
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

		// CONFLICT DETECTION: Check if token is bound to a different session
		// Skip conflict check if we're on the set-name endpoint (first-time binding)
		if currentPlayer != "" && boundSessionUUID != "" && c.Request().URL.Path != "/upload/set-name" {
			// Check if this session is the owner or has been invited
			isAllowed := false
			
			// Check if this is the primary bound session
			if boundSessionUUID == sessionUUID {
				isAllowed = true
			} else {
				// Check if this session is in the session_bindings table (invited)
				var bindingExists bool
				err = database.DB.QueryRow(context.Background(),
					`SELECT EXISTS(SELECT 1 FROM session_bindings WHERE token_id = $1 AND session_uuid = $2)`,
					tokenID, sessionUUID).Scan(&bindingExists)
				if err != nil {
					log.Printf("Error checking session bindings: %v", err)
				} else if bindingExists {
					isAllowed = true
					// Update last_active_at for this session
					database.DB.Exec(context.Background(),
						`UPDATE session_bindings SET last_active_at = NOW() WHERE token_id = $1 AND session_uuid = $2`,
						tokenID, sessionUUID)
				}
			}

			if !isAllowed {
				// Return 409 Conflict - bag is in use by another session
				log.Printf("Session conflict: token %s is bound to %s, but session %s is trying to access it",
					utils.MaskToken(token), utils.MaskToken(boundSessionUUID), utils.MaskToken(sessionUUID))
				return c.Render(http.StatusConflict, "layout", map[string]interface{}{
					"Title":           "Werkzeug wird bereits verwendet",
					"ContentTemplate": "session_conflict.content",
					"CurrentPath":     c.Request().URL.Path,
					"CurrentYear":     time.Now().Year(),
					"BagName":         bagName,
					"CurrentPlayer":   utils.SanitizePlayerName(currentPlayer),
				})
			}
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
				return next(c)
			}

			// Check if name is in session
			if sessName, ok := session.Values["player_name"].(string); ok && sessName != "" {
				// Sanitize player name before storing
				sessName = utils.SanitizePlayerName(sessName)
				
				// Bind this session_uuid to the token when setting player name
				result, err := database.DB.Exec(context.Background(),
					"UPDATE upload_tokens SET current_player = $1, session_started_at = NOW(), session_uuid = $2 WHERE id = $3",
					sessName, sessionUUID, tokenID)

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
				} else {
					log.Printf("Bound token %s to session UUID %s for player %s", 
						utils.MaskToken(token), utils.MaskToken(sessionUUID), sessName)
					// Also create entry in session_bindings for multi-session support
					_, err = database.DB.Exec(context.Background(),
						`INSERT INTO session_bindings (token_id, session_uuid, player_name, player_city, is_owner)
						 VALUES ($1, $2, $3, $4, true)
						 ON CONFLICT (token_id, session_uuid) DO UPDATE SET 
						 player_name = EXCLUDED.player_name, 
						 player_city = EXCLUDED.player_city,
						 last_active_at = NOW()`,
						tokenID, sessionUUID, sessName, session.Values["player_city"])
					if err != nil {
						log.Printf("Failed to create session binding: %v", err)
					}
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

		// Store token info in context for handler (sanitize player name)
		c.Set("token_id", tokenID)
		c.Set("token", token)
		c.Set("current_player", utils.SanitizePlayerName(currentPlayer))
		c.Set("bag_name", bagName)
		c.Set("session_number", totalSessions)
		c.Set("uploads_remaining", maxUploads-totalUploads)
		c.Set("current_player_city", utils.SanitizePlayerName(currentPlayerCity))

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
