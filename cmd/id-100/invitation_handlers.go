package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// SessionInvitation represents an invitation to join a token session
type SessionInvitation struct {
	ID                  int
	TokenID             int
	InvitationCode      string
	InvitedBySessionUUID string
	InvitedSessionUUID  sql.NullString
	CreatedAt           time.Time
	ExpiresAt           time.Time
	AcceptedAt          sql.NullTime
	RevokedAt           sql.NullTime
	IsActive            bool
	MaxUses             int
	UseCount            int
}

// AuthorizedSession represents a session authorized to use a token
type AuthorizedSession struct {
	ID             int
	TokenID        int
	SessionUUID    string
	PlayerName     string
	InvitationID   sql.NullInt32
	CreatedAt      time.Time
	LastActivityAt time.Time
	ExpiresAt      sql.NullTime
	IsActive       bool
}

// POST /upload/invite - Generate an invitation link for the current session
func generateInvitationHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Token not found"})
	}

	sessionUUID, _ := c.Get("session_uuid").(string)
	if sessionUUID == "" {
		session, _ := store.Get(c.Request(), "id-100-session")
		sessionUUID, _ = session.Values["session_uuid"].(string)
	}

	if sessionUUID == "" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Session not found"})
	}

	// Check if user is authorized for this token
	var isAuthorized bool
	err := db.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM upload_tokens 
			WHERE id = $1 AND session_uuid = $2 AND is_active = true
		) OR EXISTS(
			SELECT 1 FROM authorized_sessions
			WHERE token_id = $1 AND session_uuid = $2 AND is_active = true
		)
	`, tokenID, sessionUUID).Scan(&isAuthorized)

	if err != nil || !isAuthorized {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized"})
	}

	// Parse optional expiration duration (default 24 hours)
	expirationHours := 24
	if hours := c.QueryParam("hours"); hours != "" {
		fmt.Sscanf(hours, "%d", &expirationHours)
	}
	if expirationHours < 1 {
		expirationHours = 1
	}
	if expirationHours > 168 { // Max 7 days
		expirationHours = 168
	}

	// Generate secure invitation code
	invitationCode, err := generateSecureToken(32)
	if err != nil {
		log.Printf("Failed to generate invitation code: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate invitation"})
	}

	// Insert invitation
	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	var invitationID int
	err = db.QueryRow(context.Background(), `
		INSERT INTO session_invitations (
			token_id, invitation_code, invited_by_session_uuid, expires_at, max_uses
		) VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, tokenID, invitationCode, sessionUUID, expiresAt, 1).Scan(&invitationID)

	if err != nil {
		log.Printf("Failed to create invitation: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invitation"})
	}

	invitationURL := fmt.Sprintf("%s/upload/accept-invite?code=%s", baseURL, invitationCode)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":          "success",
		"invitation_id":   invitationID,
		"invitation_code": invitationCode,
		"invitation_url":  invitationURL,
		"expires_at":      expiresAt,
	})
}

// GET /upload/accept-invite - Accept an invitation and join the session
func acceptInvitationHandler(c echo.Context) error {
	invitationCode := c.QueryParam("code")
	if invitationCode == "" {
		return c.Render(http.StatusBadRequest, "layout", map[string]interface{}{
			"Title":           "Ungültige Einladung",
			"ContentTemplate": "status/invalid_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"ErrorMessage":    "Kein Einladungscode angegeben",
		})
	}

	// Get or create session UUID for this browser
	session, err := store.Get(c.Request(), "id-100-session")
	if err != nil {
		log.Printf("Session error: %v", err)
	}

	var sessionUUID string
	if sid, ok := session.Values["session_uuid"].(string); ok && sid != "" {
		sessionUUID = sid
	} else {
		// Generate new session UUID
		newUUID, err := generateSecureToken(32)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to generate session UUID")
		}
		sessionUUID = newUUID
		session.Values["session_uuid"] = sessionUUID
		if err := session.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Failed to save session: %v", err)
			return c.String(http.StatusInternalServerError, "Session error")
		}
	}

	// Validate invitation
	var invitation SessionInvitation
	err = db.QueryRow(context.Background(), `
		SELECT id, token_id, invitation_code, invited_by_session_uuid, 
		       created_at, expires_at, accepted_at, is_active, max_uses, use_count
		FROM session_invitations
		WHERE invitation_code = $1
	`, invitationCode).Scan(
		&invitation.ID, &invitation.TokenID, &invitation.InvitationCode,
		&invitation.InvitedBySessionUUID, &invitation.CreatedAt, &invitation.ExpiresAt,
		&invitation.AcceptedAt, &invitation.IsActive, &invitation.MaxUses, &invitation.UseCount,
	)

	if err != nil {
		log.Printf("Invitation lookup error: %v", err)
		return c.Render(http.StatusNotFound, "layout", map[string]interface{}{
			"Title":           "Einladung nicht gefunden",
			"ContentTemplate": "status/invalid_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"ErrorMessage":    "Diese Einladung existiert nicht",
		})
	}

	// Check if invitation is valid
	now := time.Now()
	if !invitation.IsActive {
		return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
			"Title":           "Einladung deaktiviert",
			"ContentTemplate": "status/invalid_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"ErrorMessage":    "Diese Einladung wurde deaktiviert",
		})
	}

	if now.After(invitation.ExpiresAt) {
		return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
			"Title":           "Einladung abgelaufen",
			"ContentTemplate": "status/invalid_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"ErrorMessage":    "Diese Einladung ist abgelaufen",
		})
	}

	if invitation.UseCount >= invitation.MaxUses {
		return c.Render(http.StatusForbidden, "layout", map[string]interface{}{
			"Title":           "Einladung bereits verwendet",
			"ContentTemplate": "status/invalid_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"ErrorMessage":    "Diese Einladung wurde bereits verwendet",
		})
	}

	// Check if this session is already authorized
	var alreadyAuthorized bool
	err = db.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM authorized_sessions
			WHERE token_id = $1 AND session_uuid = $2 AND is_active = true
		)
	`, invitation.TokenID, sessionUUID).Scan(&alreadyAuthorized)

	if err != nil {
		log.Printf("Authorization check error: %v", err)
		return c.String(http.StatusInternalServerError, "Database error")
	}

	if alreadyAuthorized {
		// Already authorized, just redirect to upload page
		var token string
		db.QueryRow(context.Background(), "SELECT token FROM upload_tokens WHERE id = $1", invitation.TokenID).Scan(&token)
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/upload?token=%s", token))
	}

	// Get token info
	var token, bagName string
	err = db.QueryRow(context.Background(), `
		SELECT token, COALESCE(bag_name, '') 
		FROM upload_tokens WHERE id = $1
	`, invitation.TokenID).Scan(&token, &bagName)

	if err != nil {
		log.Printf("Token lookup error: %v", err)
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Check if player name is needed
	playerName, _ := session.Values["player_name"].(string)
	if playerName == "" {
		// Show name entry form with invitation context
		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "Willkommen",
			"ContentTemplate": "user/enter_name_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"BagName":         bagName,
			"InvitationCode":  invitationCode,
		})
	}

	// Authorize this session
	_, err = db.Exec(context.Background(), `
		INSERT INTO authorized_sessions (
			token_id, session_uuid, player_name, invitation_id, expires_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token_id, session_uuid) DO UPDATE 
		SET player_name = EXCLUDED.player_name, 
		    is_active = true,
		    last_activity_at = NOW()
	`, invitation.TokenID, sessionUUID, playerName, invitation.ID, invitation.ExpiresAt)

	if err != nil {
		log.Printf("Failed to authorize session: %v", err)
		return c.String(http.StatusInternalServerError, "Authorization failed")
	}

	// Update invitation use count
	_, err = db.Exec(context.Background(), `
		UPDATE session_invitations 
		SET use_count = use_count + 1,
		    accepted_at = CASE WHEN accepted_at IS NULL THEN NOW() ELSE accepted_at END
		WHERE id = $1
	`, invitation.ID)

	if err != nil {
		log.Printf("Failed to update invitation: %v", err)
	}

	// Store token in session
	session.Values["token"] = token
	session.Values["token_id"] = invitation.TokenID
	session.Values["bag_name"] = bagName
	session.Save(c.Request(), c.Response())

	// Redirect to upload page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/upload?token=%s", token))
}

// POST /upload/invite/set-name - Set player name when accepting invitation
func setPlayerNameInvitationHandler(c echo.Context) error {
	playerName := c.FormValue("player_name")
	invitationCode := c.FormValue("invitation_code")

	if playerName == "" || invitationCode == "" {
		return c.String(http.StatusBadRequest, "Name und Einladungscode erforderlich")
	}

	// Consent checkbox (required)
	consent := c.FormValue("agree_privacy")
	if consent == "" {
		return c.Render(http.StatusBadRequest, "layout", map[string]interface{}{
			"Title":           "Willkommen",
			"ContentTemplate": "user/enter_name_invitation.content",
			"InvitationCode":  invitationCode,
			"FormError":       "Bitte bestätige die Datenschutzerklärung",
		})
	}

	// Save name in session
	session, _ := store.Get(c.Request(), "id-100-session")
	session.Values["player_name"] = playerName
	session.Save(c.Request(), c.Response())

	// Redirect back to accept-invite with the name now set
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/upload/accept-invite?code=%s", invitationCode))
}

// GET /upload/sessions - List active sessions for current token
func listSessionsHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Token not found"})
	}

	sessionUUID, _ := c.Get("session_uuid").(string)
	currentPlayer, _ := c.Get("current_player").(string)

	// Check if user is authorized (either primary or invited session)
	var isAuthorized bool
	err := db.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM upload_tokens 
			WHERE id = $1 AND session_uuid = $2 AND is_active = true
		) OR EXISTS(
			SELECT 1 FROM authorized_sessions
			WHERE token_id = $1 AND session_uuid = $2 AND is_active = true
		)
	`, tokenID, sessionUUID).Scan(&isAuthorized)

	if err != nil || !isAuthorized {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized"})
	}

	// Get all active sessions
	rows, err := db.Query(context.Background(), `
		SELECT session_uuid, player_name, created_at, last_activity_at
		FROM authorized_sessions
		WHERE token_id = $1 AND is_active = true
		ORDER BY created_at ASC
	`, tokenID)

	if err != nil {
		log.Printf("Failed to list sessions: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	defer rows.Close()

	type SessionInfo struct {
		SessionUUID    string    `json:"session_uuid"`
		PlayerName     string    `json:"player_name"`
		CreatedAt      time.Time `json:"created_at"`
		LastActivityAt time.Time `json:"last_activity_at"`
		IsCurrent      bool      `json:"is_current"`
	}

	var sessions []SessionInfo
	for rows.Next() {
		var s SessionInfo
		if err := rows.Scan(&s.SessionUUID, &s.PlayerName, &s.CreatedAt, &s.LastActivityAt); err != nil {
			continue
		}
		s.IsCurrent = (s.SessionUUID == sessionUUID)
		sessions = append(sessions, s)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessions":       sessions,
		"current_player": currentPlayer,
	})
}

// POST /upload/sessions/:uuid/revoke - Revoke a session's access
func revokeSessionHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Token not found"})
	}

	sessionUUID, _ := c.Get("session_uuid").(string)
	targetSessionUUID := c.Param("uuid")

	// Only the primary session (in upload_tokens.session_uuid) can revoke
	var isPrimary bool
	err := db.QueryRow(context.Background(), `
		SELECT session_uuid = $2 FROM upload_tokens WHERE id = $1
	`, tokenID, sessionUUID).Scan(&isPrimary)

	if err != nil || !isPrimary {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Only primary session can revoke access"})
	}

	// Can't revoke self
	if targetSessionUUID == sessionUUID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot revoke your own session"})
	}

	// Revoke the session
	result, err := db.Exec(context.Background(), `
		UPDATE authorized_sessions
		SET is_active = false
		WHERE token_id = $1 AND session_uuid = $2
	`, tokenID, targetSessionUUID)

	if err != nil {
		log.Printf("Failed to revoke session: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke session"})
	}

	if result.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Session not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Session revoked"})
}
