package app

import (
	"context"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"id-100/internal/database"
	"id-100/internal/middleware"
	"id-100/internal/templates"
	"id-100/internal/utils"
)

// ReleaseBagHandler handles releasing the bag (unbinding session)
func ReleaseBagHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.String(http.StatusForbidden, "Token nicht gefunden")
	}

	sessionUUID, ok := c.Get("session_uuid").(string)
	if !ok {
		return c.String(http.StatusInternalServerError, "Session UUID nicht gefunden")
	}

	// Get session to clear player data
	session, err := middleware.Store.Get(c.Request(), "id-100-session")
	if err != nil {
		log.Printf("Session error: %v", err)
	}

	// Clear session_uuid binding and current_player in database
	_, err = database.DB.Exec(context.Background(),
		`UPDATE upload_tokens 
		 SET session_uuid = NULL, current_player = NULL, current_player_city = NULL
		 WHERE id = $1 AND session_uuid = $2`,
		tokenID, sessionUUID)

	if err != nil {
		log.Printf("Failed to release bag: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Zurückgeben des Werkzeugs")
	}

	// Deactivate all invitations for this token
	_, err = database.DB.Exec(context.Background(),
		`UPDATE invitations SET is_active = false WHERE token_id = $1`,
		tokenID)

	if err != nil {
		log.Printf("Failed to deactivate invitations: %v", err)
	}

	// Deactivate active session
	_, err = database.DB.Exec(context.Background(),
		`UPDATE active_sessions SET is_active = false WHERE token_id = $1 AND session_uuid = $2`,
		tokenID, sessionUUID)

	if err != nil {
		log.Printf("Failed to deactivate session: %v", err)
	}

	// Clear session data
	delete(session.Values, "token")
	delete(session.Values, "token_id")
	delete(session.Values, "player_name")
	delete(session.Values, "player_city")
	delete(session.Values, "bag_name")
	session.Save(c.Request(), c.Response())

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           "Werkzeug zurückgegeben",
		"ContentTemplate": "bag_released.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
	}))
}

// GenerateInvitationHandler creates an invitation code for the current token
func GenerateInvitationHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.String(http.StatusForbidden, "Token nicht gefunden")
	}

	sessionUUID, ok := c.Get("session_uuid").(string)
	if !ok {
		return c.String(http.StatusInternalServerError, "Session UUID nicht gefunden")
	}

	// Generate invitation code
	invitationCode, err := utils.GenerateInvitationCode()
	if err != nil {
		log.Printf("Failed to generate invitation code: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Erstellen der Einladung")
	}

	// Set expiry to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)

	// Insert invitation into database
	_, err = database.DB.Exec(context.Background(),
		`INSERT INTO invitations (token_id, invitation_code, created_by_session_uuid, expires_at, is_active)
		 VALUES ($1, $2, $3, $4, true)`,
		tokenID, invitationCode, sessionUUID, expiresAt)

	if err != nil {
		log.Printf("Failed to insert invitation: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Speichern der Einladung")
	}

	// Return the invitation code
	return c.JSON(http.StatusOK, map[string]interface{}{
		"invitation_code": invitationCode,
		"expires_at":      expiresAt,
	})
}

// AcceptInvitationHandler handles accepting an invitation
func AcceptInvitationHandler(c echo.Context) error {
	invitationCode := c.FormValue("invitation_code")
	if invitationCode == "" {
		return c.String(http.StatusBadRequest, "Einladungscode erforderlich")
	}

	// Get session
	session, err := middleware.Store.Get(c.Request(), "id-100-session")
	if err != nil {
		log.Printf("Session error: %v", err)
	}

	// Get or create session UUID
	sessionUUID, err := middleware.GetOrCreateSessionUUID(session)
	if err != nil {
		log.Printf("Failed to create session UUID: %v", err)
		return c.String(http.StatusInternalServerError, "Session initialization failed")
	}

	// Check if invitation is valid
	var tokenID int
	var createdBySessionUUID string
	var isActive bool
	var expiresAt time.Time
	var token, bagName string

	err = database.DB.QueryRow(context.Background(),
		`SELECT i.token_id, i.created_by_session_uuid, i.is_active, i.expires_at,
		        ut.token, ut.bag_name
		 FROM invitations i
		 JOIN upload_tokens ut ON i.token_id = ut.id
		 WHERE i.invitation_code = $1`,
		invitationCode).Scan(&tokenID, &createdBySessionUUID, &isActive, &expiresAt, &token, &bagName)

	if err != nil {
		log.Printf("Invitation not found: %v", err)
		return c.Render(http.StatusNotFound, "layout", templates.MergeTemplateData(map[string]interface{}{
			"Title":           "Ungültige Einladung",
			"ContentTemplate": "invalid_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
		}))
	}

	// Check if invitation is still active and not expired
	if !isActive || time.Now().After(expiresAt) {
		return c.Render(http.StatusForbidden, "layout", templates.MergeTemplateData(map[string]interface{}{
			"Title":           "Einladung abgelaufen",
			"ContentTemplate": "expired_invitation.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
		}))
	}

	// Mark invitation as accepted
	_, err = database.DB.Exec(context.Background(),
		`UPDATE invitations 
		 SET accepted_by_session_uuid = $1, accepted_at = NOW()
		 WHERE invitation_code = $2`,
		sessionUUID, invitationCode)

	if err != nil {
		log.Printf("Failed to accept invitation: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Annehmen der Einladung")
	}

	// Store token and invitation flag in session
	session.Values["token"] = token
	session.Values["token_id"] = tokenID
	session.Values["bag_name"] = bagName
	session.Values["session_uuid"] = sessionUUID
	session.Values["from_invitation"] = true  // Flag to indicate this user came from an invitation
	session.Save(c.Request(), c.Response())

	// Redirect to name entry page so user can enter their name
	// The TokenWithSession middleware will recognize they have a valid token from invitation
	return c.Redirect(http.StatusSeeOther, "/upload?token="+token)
}

// ListActiveSessionsHandler shows all active sessions for the current token
func ListActiveSessionsHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.String(http.StatusForbidden, "Token nicht gefunden")
	}

	sessionUUID, ok := c.Get("session_uuid").(string)
	if !ok {
		return c.String(http.StatusInternalServerError, "Session UUID nicht gefunden")
	}

	bagName, _ := c.Get("bag_name").(string)

	// Get all active sessions for this token
	rows, err := database.DB.Query(context.Background(),
		`SELECT session_uuid, player_name, player_city, started_at, last_activity_at, is_active
		 FROM active_sessions
		 WHERE token_id = $1 AND is_active = true
		 ORDER BY started_at DESC`,
		tokenID)

	if err != nil {
		log.Printf("Failed to get active sessions: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Laden der Sitzungen")
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var uuid, playerName, playerCity string
		var startedAt, lastActivityAt time.Time
		var isActive bool

		err := rows.Scan(&uuid, &playerName, &playerCity, &startedAt, &lastActivityAt, &isActive)
		if err != nil {
			log.Printf("Failed to scan session: %v", err)
			continue
		}

		// Sanitize player name to prevent XSS
		playerName = html.EscapeString(playerName)
		playerCity = html.EscapeString(playerCity)

		sessions = append(sessions, map[string]interface{}{
			"uuid":              uuid,
			"player_name":       playerName,
			"player_city":       playerCity,
			"started_at":        startedAt,
			"last_activity_at":  lastActivityAt,
			"is_active":         isActive,
			"is_current":        uuid == sessionUUID,
		})
	}

	csrfToken, _ := c.Get("csrf_token").(string)

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           "Aktive Sitzungen",
		"ContentTemplate": "active_sessions.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"Sessions":        sessions,
		"BagName":         bagName,
		"CSRFToken":       csrfToken,
	}))
}

// RevokeSessionHandler revokes a specific session
func RevokeSessionHandler(c echo.Context) error {
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.String(http.StatusForbidden, "Token nicht gefunden")
	}

	currentSessionUUID, ok := c.Get("session_uuid").(string)
	if !ok {
		return c.String(http.StatusInternalServerError, "Session UUID nicht gefunden")
	}

	// Get the session UUID to revoke
	revokeSessionUUID := c.Param("session_uuid")
	if revokeSessionUUID == "" {
		return c.String(http.StatusBadRequest, "Session UUID erforderlich")
	}

	// Don't allow revoking own session
	if revokeSessionUUID == currentSessionUUID {
		return c.String(http.StatusBadRequest, "Du kannst deine eigene Sitzung nicht widerrufen")
	}

	// Deactivate the session
	_, err := database.DB.Exec(context.Background(),
		`UPDATE active_sessions 
		 SET is_active = false 
		 WHERE token_id = $1 AND session_uuid = $2`,
		tokenID, revokeSessionUUID)

	if err != nil {
		log.Printf("Failed to revoke session: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Widerrufen der Sitzung")
	}

	// Deactivate invitations for this session
	_, err = database.DB.Exec(context.Background(),
		`UPDATE invitations 
		 SET is_active = false 
		 WHERE token_id = $1 AND (created_by_session_uuid = $2 OR accepted_by_session_uuid = $2)`,
		tokenID, revokeSessionUUID)

	if err != nil {
		log.Printf("Failed to deactivate invitations: %v", err)
	}

	return c.Redirect(http.StatusSeeOther, "/upload/sessions")
}
