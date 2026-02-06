package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"id-100/internal/database"
	"id-100/internal/middleware"
	"id-100/internal/models"
	"id-100/internal/utils"
)

// ReleaseBagHandler allows users to release their bag and unbind their session
func ReleaseBagHandler(c echo.Context) error {
	// Get session_uuid and token_id from context
	sessionUUID, _ := c.Get("session_uuid").(string)
	tokenID, _ := c.Get("token_id").(int)
	token, _ := c.Get("token").(string)

	if sessionUUID == "" || tokenID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid session"})
	}

	// Check if this session is the owner (primary session)
	var boundSessionUUID string
	err := database.DB.QueryRow(context.Background(),
		"SELECT COALESCE(session_uuid, '') FROM upload_tokens WHERE id = $1",
		tokenID).Scan(&boundSessionUUID)

	if err != nil {
		log.Printf("Error checking session ownership: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}

	// Clear the primary session binding if this is the owner
	if boundSessionUUID == sessionUUID {
		_, err = database.DB.Exec(context.Background(),
			`UPDATE upload_tokens 
			 SET session_uuid = '', 
			     current_player = NULL, 
			     current_player_city = NULL
			 WHERE id = $1`,
			tokenID)

		if err != nil {
			log.Printf("Error releasing bag for token %s: %v", utils.MaskToken(token), err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
		}
		log.Printf("Session %s released bag for token %s", utils.MaskToken(sessionUUID), utils.MaskToken(token))
	}

	// Remove from session_bindings
	_, err = database.DB.Exec(context.Background(),
		"DELETE FROM session_bindings WHERE token_id = $1 AND session_uuid = $2",
		tokenID, sessionUUID)

	if err != nil {
		log.Printf("Error removing session binding: %v", err)
	}

	// Clear session values
	session, _ := middleware.Store.Get(c.Request(), "id-100-session")
	delete(session.Values, "token")
	delete(session.Values, "token_id")
	delete(session.Values, "player_name")
	delete(session.Values, "player_city")
	delete(session.Values, "bag_name")
	session.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Werkzeug erfolgreich zurückgegeben",
	})
}

// GenerateInvitationHandler creates an invitation code for the current token
func GenerateInvitationHandler(c echo.Context) error {
	tokenID, _ := c.Get("token_id").(int)
	sessionUUID, _ := c.Get("session_uuid").(string)

	if tokenID == 0 || sessionUUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid session"})
	}

	// Check if this session is the owner or has access
	var hasAccess bool
	err := database.DB.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM upload_tokens WHERE id = $1 AND session_uuid = $2
			UNION
			SELECT 1 FROM session_bindings WHERE token_id = $1 AND session_uuid = $2
		)`,
		tokenID, sessionUUID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	// Generate invitation code
	code, err := utils.GenerateInvitationCode()
	if err != nil {
		log.Printf("Failed to generate invitation code: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}

	// Insert invitation code (expires in 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = database.DB.Exec(context.Background(),
		`INSERT INTO invitation_codes (token_id, code, created_by_session_uuid, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		tokenID, code, sessionUUID, expiresAt)

	if err != nil {
		log.Printf("Failed to insert invitation code: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}

	log.Printf("Generated invitation code %s for token_id %d by session %s", 
		code, tokenID, utils.MaskToken(sessionUUID))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "success",
		"code":       code,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// AcceptInvitationHandler accepts an invitation code and binds the session
func AcceptInvitationHandler(c echo.Context) error {
	type AcceptRequest struct {
		Code string `json:"code"`
	}

	var req AcceptRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	sessionUUID, _ := c.Get("session_uuid").(string)
	if sessionUUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid session"})
	}

	// Look up invitation code
	var invitation models.InvitationCode
	err := database.DB.QueryRow(context.Background(),
		`SELECT id, token_id, code, created_by_session_uuid, expires_at, used
		 FROM invitation_codes
		 WHERE code = $1`,
		req.Code).Scan(&invitation.ID, &invitation.TokenID, &invitation.Code,
		&invitation.CreatedBySessionUUID, &invitation.ExpiresAt, &invitation.Used)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Invalid invitation code"})
	}

	// Check if already used
	if invitation.Used {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invitation code already used"})
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invitation code expired"})
	}

	// Get token info
	var token, bagName, currentPlayer string
	err = database.DB.QueryRow(context.Background(),
		"SELECT token, COALESCE(bag_name, ''), COALESCE(current_player, '') FROM upload_tokens WHERE id = $1",
		invitation.TokenID).Scan(&token, &bagName, &currentPlayer)

	if err != nil {
		log.Printf("Error fetching token info: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}

	// Get player name from session or prompt for it
	session, _ := middleware.Store.Get(c.Request(), "id-100-session")
	playerName, _ := session.Values["player_name"].(string)
	playerCity, _ := session.Values["player_city"].(string)

	if playerName == "" {
		// Store the invitation code in session for later use
		session.Values["pending_invitation"] = req.Code
		session.Values["token"] = token
		session.Values["token_id"] = invitation.TokenID
		session.Values["bag_name"] = bagName
		session.Save(c.Request(), c.Response())

		// Redirect to name entry
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":       "need_name",
			"redirect_url": fmt.Sprintf("/upload?token=%s", token),
		})
	}

	// Sanitize player name
	playerName = utils.SanitizePlayerName(playerName)
	playerCity = utils.SanitizePlayerName(playerCity)

	// Create session binding
	_, err = database.DB.Exec(context.Background(),
		`INSERT INTO session_bindings (token_id, session_uuid, player_name, player_city, is_owner)
		 VALUES ($1, $2, $3, $4, false)
		 ON CONFLICT (token_id, session_uuid) DO UPDATE SET 
		 player_name = EXCLUDED.player_name,
		 player_city = EXCLUDED.player_city,
		 last_active_at = NOW()`,
		invitation.TokenID, sessionUUID, playerName, playerCity)

	if err != nil {
		log.Printf("Failed to create session binding: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}

	// Mark invitation as used
	_, err = database.DB.Exec(context.Background(),
		`UPDATE invitation_codes 
		 SET used = true, used_by_session_uuid = $1, used_at = NOW()
		 WHERE id = $2`,
		sessionUUID, invitation.ID)

	if err != nil {
		log.Printf("Failed to mark invitation as used: %v", err)
	}

	// Store token in session
	session.Values["token"] = token
	session.Values["token_id"] = invitation.TokenID
	session.Values["bag_name"] = bagName
	session.Save(c.Request(), c.Response())

	log.Printf("Session %s accepted invitation to token %s", 
		utils.MaskToken(sessionUUID), utils.MaskToken(token))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       "success",
		"redirect_url": fmt.Sprintf("/upload?token=%s", token),
		"bag_name":     bagName,
	})
}

// ListActiveSessionsHandler lists all active sessions for a token
func ListActiveSessionsHandler(c echo.Context) error {
	tokenID, _ := c.Get("token_id").(int)
	sessionUUID, _ := c.Get("session_uuid").(string)

	if tokenID == 0 || sessionUUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid session"})
	}

	// Check if this session has access
	var hasAccess bool
	err := database.DB.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM upload_tokens WHERE id = $1 AND session_uuid = $2
			UNION
			SELECT 1 FROM session_bindings WHERE token_id = $1 AND session_uuid = $2
		)`,
		tokenID, sessionUUID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	// Get all active sessions
	rows, err := database.DB.Query(context.Background(),
		`SELECT id, session_uuid, player_name, COALESCE(player_city, ''), is_owner, created_at, last_active_at
		 FROM session_bindings
		 WHERE token_id = $1
		 ORDER BY is_owner DESC, last_active_at DESC`,
		tokenID)

	if err != nil {
		log.Printf("Error fetching active sessions: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var s models.SessionBinding
		if err := rows.Scan(&s.ID, &s.SessionUUID, &s.PlayerName, &s.PlayerCity, 
			&s.IsOwner, &s.CreatedAt, &s.LastActiveAt); err != nil {
			continue
		}

		sessions = append(sessions, map[string]interface{}{
			"id":              s.ID,
			"session_uuid":    utils.MaskToken(s.SessionUUID),
			"player_name":     utils.SanitizePlayerName(s.PlayerName),
			"player_city":     utils.SanitizePlayerName(s.PlayerCity),
			"is_owner":        s.IsOwner,
			"is_current":      s.SessionUUID == sessionUUID,
			"created_at":      s.CreatedAt.Format(time.RFC3339),
			"last_active_at":  s.LastActiveAt.Format(time.RFC3339),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "success",
		"sessions": sessions,
	})
}

// RevokeSessionHandler revokes access for a specific session (owner only)
func RevokeSessionHandler(c echo.Context) error {
	tokenID, _ := c.Get("token_id").(int)
	sessionUUID, _ := c.Get("session_uuid").(string)
	targetSessionID := c.Param("session_id")

	if tokenID == 0 || sessionUUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid session"})
	}

	// Check if this session is the owner
	var isOwner bool
	err := database.DB.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM upload_tokens WHERE id = $1 AND session_uuid = $2
		)`,
		tokenID, sessionUUID).Scan(&isOwner)

	if err != nil || !isOwner {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Only the owner can revoke access"})
	}

	// Delete the session binding
	result, err := database.DB.Exec(context.Background(),
		"DELETE FROM session_bindings WHERE id = $1 AND token_id = $2 AND is_owner = false",
		targetSessionID, tokenID)

	if err != nil {
		log.Printf("Error revoking session: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Server error"})
	}

	if result.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Session not found or cannot revoke owner"})
	}

	log.Printf("Session %s revoked access for session ID %s", 
		utils.MaskToken(sessionUUID), targetSessionID)

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Session access revoked",
	})
}
