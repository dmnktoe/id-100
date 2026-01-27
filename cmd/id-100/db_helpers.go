package main

import (
	"context"
	"database/sql"
	"time"
)

// TokenMeta is a lightweight representation of the upload token DB row used in middleware
type TokenMeta struct {
	ID               int
	IsActive         bool
	MaxUploads       int
	TotalUploads     int
	TotalSessions    int
	CurrentPlayer    string
	BagName          string
	SessionStartedAt time.Time
	SessionUUID      sql.NullString
}

// function variables so tests can override DB behaviour
var getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
	var m TokenMeta
	var sessStart time.Time
	var sessUUID sql.NullString
	err := db.QueryRow(ctx, `SELECT id, is_active, max_uploads, total_uploads, total_sessions,
		 COALESCE(current_player, ''), COALESCE(bag_name, ''), COALESCE(session_started_at, created_at), COALESCE(session_uuid, '')
		 FROM upload_tokens WHERE token = $1`, token).Scan(&m.ID, &m.IsActive, &m.MaxUploads, &m.TotalUploads, &m.TotalSessions, &m.CurrentPlayer, &m.BagName, &sessStart, &sessUUID)
	if err != nil {
		return m, err
	}
	m.SessionStartedAt = sessStart
	m.SessionUUID = sessUUID
	return m, nil
}

var updateTokenCurrentPlayer = func(ctx context.Context, token, playerName string, sessionUUID interface{}) error {
	// Use COALESCE on session_uuid so existing binding stays if set
	_, err := db.Exec(ctx, "UPDATE upload_tokens SET current_player = $1, session_started_at = NOW(), session_uuid = COALESCE(session_uuid, $3) WHERE token = $2", playerName, token, sessionUUID)
	return err
}

// fetchDerivenList returns the list of derives for the upload page (used by handler and tests)
var fetchDerivenList = func(ctx context.Context) ([]Derive, error) {
	rows, err := db.Query(ctx, `
SELECT d.number, d.title, COALESCE((SELECT COUNT(*) FROM contributions WHERE derive_id = d.id),0) as contrib_count
FROM deriven d
ORDER BY d.number ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Derive
	for rows.Next() {
		var d Derive
		if err := rows.Scan(&d.Number, &d.Title, &d.ContribCount); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

// fetchSessionUploads returns contributions uploaded in this session for display under the form
var fetchSessionUploads = func(ctx context.Context, tokenID, sessionNumber int) ([]map[string]interface{}, error) {
	uRows, err := db.Query(ctx, `
		SELECT c.id, d.number, c.image_url, COALESCE(c.image_lqip, '')
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		JOIN deriven d ON d.id = c.derive_id
		WHERE ul.token_id = $1 AND ul.session_number = $2
		ORDER BY ul.uploaded_at DESC
	`, tokenID, sessionNumber)
	if err != nil {
		return nil, err
	}
	defer uRows.Close()

	var sessionContribs []map[string]interface{}
	for uRows.Next() {
		var id int
		var deriveNumber int
		var imageUrl string
		var imageLqip string
		if err := uRows.Scan(&id, &deriveNumber, &imageUrl, &imageLqip); err != nil {
			continue
		}
		sessionContribs = append(sessionContribs, map[string]interface{}{
			"id":         id,
			"number":     deriveNumber,
			"image_url":  ensureFullImageURL(imageUrl),
			"image_lqip": imageLqip,
		})
	}
	return sessionContribs, nil
}
