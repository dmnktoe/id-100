package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"id-100/internal/database"
	"id-100/internal/models"
)

// API layer for database queries - extracted from app.go and admin.go

// GetDistinctCities retrieves all distinct cities from contributions
func GetDistinctCities(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT user_city FROM contributions WHERE user_city IS NOT NULL AND user_city != '' ORDER BY user_city ASC`
	rows, err := database.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err == nil {
			cities = append(cities, city)
		}
	}
	return cities, nil
}

// GetDerivenCount returns the total count of deriven (optionally filtered by city)
func GetDerivenCount(ctx context.Context, cityFilter string) (int, error) {
	var totalCount int
	var err error

	if cityFilter != "" {
		countQuery := `SELECT COUNT(DISTINCT d.id) FROM deriven d 
		              INNER JOIN contributions c ON c.derive_id = d.id 
		              WHERE c.user_city = $1`
		err = database.DB.QueryRow(ctx, countQuery, cityFilter).Scan(&totalCount)
	} else {
		countQuery := "SELECT COUNT(*) FROM deriven"
		err = database.DB.QueryRow(ctx, countQuery).Scan(&totalCount)
	}

	return totalCount, err
}

// GetDerivenList retrieves a paginated list of deriven (optionally filtered by city)
func GetDerivenList(ctx context.Context, cityFilter string, limit, offset int) ([]models.Derive, error) {
	var rows pgx.Rows
	var err error

	if cityFilter != "" {
		query := `
            SELECT 
                d.id, d.number, d.title, d.description, 
                COALESCE(c.image_url, ''), COALESCE(c.image_lqip, ''),
                (SELECT COUNT(*) FROM contributions WHERE derive_id = d.id) as contrib_count,
                d.points
            FROM deriven d
            INNER JOIN contributions city_contrib ON city_contrib.derive_id = d.id AND city_contrib.user_city = $1
            LEFT JOIN LATERAL (
                SELECT image_url, image_lqip FROM contributions 
                WHERE derive_id = d.id AND user_city = $1
                ORDER BY created_at DESC LIMIT 1
            ) c ON true
            GROUP BY d.id, d.number, d.title, d.description, c.image_url, c.image_lqip, d.points
            ORDER BY d.number ASC 
            LIMIT $2 OFFSET $3`
		rows, err = database.DB.Query(ctx, query, cityFilter, limit, offset)
	} else {
		query := `
            SELECT 
                d.id, d.number, d.title, d.description, 
                COALESCE(c.image_url, ''), COALESCE(c.image_lqip, ''),
                (SELECT COUNT(*) FROM contributions WHERE derive_id = d.id) as contrib_count,
                d.points
            FROM deriven d
            LEFT JOIN LATERAL (
                SELECT image_url, image_lqip FROM contributions 
                WHERE derive_id = d.id 
                ORDER BY created_at DESC LIMIT 1
            ) c ON true
            ORDER BY d.number ASC 
            LIMIT $1 OFFSET $2`
		rows, err = database.DB.Query(ctx, query, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deriven []models.Derive
	for rows.Next() {
		var d models.Derive
		if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.ImageLqip, &d.ContribCount, &d.Points); err != nil {
			return nil, err
		}
		deriven = append(deriven, d)
	}

	return deriven, nil
}

// GetDeriveByNumber retrieves a single derive by its number
func GetDeriveByNumber(ctx context.Context, number string) (*models.Derive, error) {
	var d models.Derive
	query := `
            SELECT d.id, d.number, d.title, d.description, COALESCE(c.image_url, ''), d.points
            FROM deriven d
            LEFT JOIN LATERAL (
                SELECT image_url FROM contributions WHERE derive_id = d.id ORDER BY created_at DESC LIMIT 1
            ) c ON true
            WHERE d.number = $1`

	err := database.DB.QueryRow(ctx, query, number).Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.Points)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// GetDeriveContributions retrieves contributions for a derive (optionally filtered by city)
func GetDeriveContributions(ctx context.Context, deriveID int, cityFilter string) ([]models.Contribution, error) {
	var rows pgx.Rows
	var err error

	if cityFilter != "" {
		rows, err = database.DB.Query(ctx,
			"SELECT image_url, COALESCE(image_lqip,''), user_name, COALESCE(user_city,''), COALESCE(user_comment,''), created_at FROM contributions WHERE derive_id = $1 AND user_city = $2 ORDER BY created_at DESC", deriveID, cityFilter)
	} else {
		rows, err = database.DB.Query(ctx,
			"SELECT image_url, COALESCE(image_lqip,''), user_name, COALESCE(user_city,''), COALESCE(user_comment,''), created_at FROM contributions WHERE derive_id = $1 ORDER BY created_at DESC", deriveID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contribs []models.Contribution
	for rows.Next() {
		var ct models.Contribution
		rows.Scan(&ct.ImageUrl, &ct.ImageLqip, &ct.UserName, &ct.UserCity, &ct.UserComment, &ct.CreatedAt)
		contribs = append(contribs, ct)
	}

	return contribs, nil
}

// GetDerivenForUpload retrieves all deriven for the upload form
func GetDerivenForUpload(ctx context.Context) ([]models.Derive, error) {
	rows, err := database.DB.Query(ctx, `
SELECT d.number, d.title, COALESCE(d.points, 0) as points, COALESCE((SELECT COUNT(*) FROM contributions WHERE derive_id = d.id),0) as contrib_count
FROM deriven d
ORDER BY d.number ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Derive
	for rows.Next() {
		var d models.Derive
		if err := rows.Scan(&d.Number, &d.Title, &d.Points, &d.ContribCount); err != nil {
			return nil, err
		}
		list = append(list, d)
	}

	return list, nil
}

// GetSessionUploads retrieves uploads for a specific token and session
func GetSessionUploads(ctx context.Context, tokenID, sessionNumber int) ([]map[string]interface{}, error) {
	uRows, err := database.DB.Query(ctx, `
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
			"image_url":  imageUrl,
			"image_lqip": imageLqip,
		})
	}

	return sessionContribs, nil
}

// GetDeriveIDByNumber retrieves the internal ID for a derive by its number
func GetDeriveIDByNumber(ctx context.Context, deriveNumber string) (int, error) {
	var internalID int
	err := database.DB.QueryRow(ctx, "SELECT id FROM deriven WHERE number = $1", deriveNumber).Scan(&internalID)
	return internalID, err
}

// InsertContribution inserts a new contribution and returns its ID
func InsertContribution(ctx context.Context, deriveID int, imageURL, imageLqip, userName, userCity, userComment string) (int, error) {
	var contributionID int
	err := database.DB.QueryRow(ctx,
		"INSERT INTO contributions (derive_id, image_url, image_lqip, user_name, user_city, user_comment) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		deriveID, imageURL, imageLqip, userName, userCity, userComment).Scan(&contributionID)
	return contributionID, err
}

// InsertUploadLog logs an upload to the upload_logs table
func InsertUploadLog(ctx context.Context, tokenID, deriveNumber int, playerName string, sessionNumber, contributionID int) error {
	_, err := database.DB.Exec(ctx,
		`INSERT INTO upload_logs (token_id, derive_number, player_name, session_number, contribution_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		tokenID, deriveNumber, playerName, sessionNumber, contributionID)
	return err
}

// IncrementTokenUploadCount increments the total_uploads counter for a token
func IncrementTokenUploadCount(ctx context.Context, tokenID int) error {
	_, err := database.DB.Exec(ctx,
		"UPDATE upload_tokens SET total_uploads = total_uploads + 1 WHERE id = $1",
		tokenID)
	return err
}

// InsertBagRequest inserts a new bag request
func InsertBagRequest(ctx context.Context, email string) error {
	_, err := database.DB.Exec(ctx, "INSERT INTO bag_requests (email) VALUES ($1)", email)
	return err
}

// GetBagNameByToken retrieves the bag name for a token
func GetBagNameByToken(ctx context.Context, token string) (string, error) {
	var bagName string
	err := database.DB.QueryRow(ctx, "SELECT COALESCE(bag_name,'') FROM upload_tokens WHERE token = $1", token).Scan(&bagName)
	return bagName, err
}

// UpdatePlayerNameAndCity updates the current player and city for a token
func UpdatePlayerNameAndCity(ctx context.Context, playerName, playerCity, token string) error {
	_, err := database.DB.Exec(ctx,
		"UPDATE upload_tokens SET current_player = $1, current_player_city = $2, session_started_at = NOW() WHERE token = $3",
		playerName, playerCity, token)
	return err
}

// GetContributionForDeletion retrieves contribution info and verifies ownership for deletion
func GetContributionForDeletion(ctx context.Context, contributionID, tokenID, sessionNumber int) (imageURL string, err error) {
	var uploadLogTokenID, uploadLogSessionNumber int
	err = database.DB.QueryRow(ctx, `
		SELECT c.image_url, ul.token_id, ul.session_number
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		WHERE c.id = $1
	`, contributionID).Scan(&imageURL, &uploadLogTokenID, &uploadLogSessionNumber)

	if err != nil {
		return "", err
	}

	// Check if the contribution belongs to the current session
	if uploadLogTokenID != tokenID || uploadLogSessionNumber != sessionNumber {
		return "", pgx.ErrNoRows // Use standard error to indicate ownership mismatch
	}

	return imageURL, nil
}

// DeleteUploadLog deletes an upload log entry
func DeleteUploadLog(ctx context.Context, contributionID int) error {
	_, err := database.DB.Exec(ctx, "DELETE FROM upload_logs WHERE contribution_id = $1", contributionID)
	return err
}

// DeleteContribution deletes a contribution
func DeleteContribution(ctx context.Context, contributionID int) (int64, error) {
	result, err := database.DB.Exec(ctx, "DELETE FROM contributions WHERE id = $1", contributionID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// DecrementTokenUploadCount decrements the total_uploads counter for a token
func DecrementTokenUploadCount(ctx context.Context, tokenID int) error {
	_, err := database.DB.Exec(ctx,
		"UPDATE upload_tokens SET total_uploads = total_uploads - 1 WHERE id = $1 AND total_uploads > 0",
		tokenID)
	return err
}

// Admin-specific database queries

// GetAllTokens retrieves all upload tokens
func GetAllTokens(ctx context.Context) ([]models.TokenInfo, error) {
	rows, err := database.DB.Query(ctx, `
		SELECT id, token, COALESCE(bag_name, ''), COALESCE(current_player, ''), COALESCE(current_player_city, ''),
		       is_active, max_uploads, total_uploads, total_sessions,
		       COALESCE(session_started_at, created_at), created_at
		FROM upload_tokens
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
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

	return tokens, nil
}

// GetRecentContributions retrieves recent contributions for the admin dashboard
func GetRecentContributions(ctx context.Context, limit int) ([]models.RecentContrib, error) {
	contribRows, err := database.DB.Query(ctx, `
		SELECT c.id, c.image_url, COALESCE(ul.player_name, 'Anonym'), ul.derive_number
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		ORDER BY c.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer contribRows.Close()

	var recentContribs []models.RecentContrib
	for contribRows.Next() {
		var rc models.RecentContrib
		if err := contribRows.Scan(&rc.ID, &rc.ImageUrl, &rc.PlayerName, &rc.DeriveNumber); err != nil {
			continue
		}
		recentContribs = append(recentContribs, rc)
	}

	return recentContribs, nil
}

// GetBagRequestCounts retrieves counts of open and handled bag requests
func GetBagRequestCounts(ctx context.Context) (openCount, handledCount int, err error) {
	err = database.DB.QueryRow(ctx, "SELECT COUNT(*) FROM bag_requests WHERE handled = FALSE").Scan(&openCount)
	if err != nil {
		return 0, 0, err
	}

	err = database.DB.QueryRow(ctx, "SELECT COUNT(*) FROM bag_requests WHERE handled = TRUE").Scan(&handledCount)
	if err != nil {
		return 0, 0, err
	}

	return openCount, handledCount, nil
}

// GetBagRequests retrieves bag requests optionally filtered by status
func GetBagRequests(ctx context.Context, status string, limit int) ([]models.BagRequest, error) {
	var query string
	switch status {
	case "open":
		query = "SELECT id, email, created_at, handled FROM bag_requests WHERE handled = FALSE ORDER BY created_at DESC LIMIT $1"
	case "handled":
		query = "SELECT id, email, created_at, handled FROM bag_requests WHERE handled = TRUE ORDER BY created_at DESC LIMIT $1"
	default:
		query = "SELECT id, email, created_at, handled FROM bag_requests ORDER BY created_at DESC LIMIT $1"
	}

	reqRows, err := database.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer reqRows.Close()

	var bagRequests []models.BagRequest
	for reqRows.Next() {
		var br models.BagRequest
		if err := reqRows.Scan(&br.ID, &br.Email, &br.CreatedAt, &br.Handled); err == nil {
			bagRequests = append(bagRequests, br)
		}
	}

	return bagRequests, nil
}

// MarkBagRequestHandled marks a bag request as handled
func MarkBagRequestHandled(ctx context.Context, id int) (int64, error) {
	res, err := database.DB.Exec(ctx, "UPDATE bag_requests SET handled = TRUE WHERE id = $1", id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// ResetToken resets a token for the next player
func ResetToken(ctx context.Context, tokenID string) (int64, error) {
	result, err := database.DB.Exec(ctx,
		`UPDATE upload_tokens 
		 SET total_uploads = 0, 
		     total_sessions = total_sessions + 1,
		     session_started_at = NOW(),
		     current_player = NULL,
		     is_active = true
		 WHERE id = $1`,
		tokenID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// DeactivateToken deactivates a token
func DeactivateToken(ctx context.Context, tokenID string) (int64, error) {
	result, err := database.DB.Exec(ctx,
		"UPDATE upload_tokens SET is_active = false WHERE id = $1",
		tokenID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// AssignTokenToPlayer assigns a token to a specific player
func AssignTokenToPlayer(ctx context.Context, tokenID, playerName string) (int64, error) {
	result, err := database.DB.Exec(ctx,
		`UPDATE upload_tokens 
		 SET current_player = $1,
		     session_started_at = NOW(),
		     is_active = true
		 WHERE id = $2`,
		playerName, tokenID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// CreateToken creates a new token/bag
func CreateToken(ctx context.Context, token, bagName string, maxUploads int) (int, error) {
	var tokenID int
	err := database.DB.QueryRow(ctx,
		`INSERT INTO upload_tokens (token, bag_name, max_uploads, total_sessions) 
		 VALUES ($1, $2, $3, 1) RETURNING id`,
		token, bagName, maxUploads).Scan(&tokenID)
	return tokenID, err
}

// GetTokenByID retrieves a token and bag name by ID
func GetTokenByID(ctx context.Context, tokenID string) (token, bagName string, err error) {
	err = database.DB.QueryRow(ctx,
		"SELECT token, COALESCE(bag_name, '') FROM upload_tokens WHERE id = $1",
		tokenID).Scan(&token, &bagName)
	return token, bagName, err
}

// UpdateTokenQuota updates the max_uploads quota for a token
func UpdateTokenQuota(ctx context.Context, tokenID string, maxUploads int) (int64, error) {
	result, err := database.DB.Exec(ctx,
		"UPDATE upload_tokens SET max_uploads = $1 WHERE id = $2",
		maxUploads, tokenID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// GetContributionForAdminDeletion retrieves contribution info for admin deletion
func GetContributionForAdminDeletion(ctx context.Context, contributionID int) (imageURL string, tokenID int, err error) {
	err = database.DB.QueryRow(ctx, `
		SELECT c.image_url, ul.token_id
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		WHERE c.id = $1
	`, contributionID).Scan(&imageURL, &tokenID)

	if err != nil {
		log.Printf("Failed to fetch contribution: %v", err)
		return "", 0, err
	}

	return imageURL, tokenID, nil
}
