package handlers

import (
	"bytes"
	"context"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chai2010/webp"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"id-100/internal/database"
	"id-100/internal/imgutil"
	"id-100/internal/middleware"
	"id-100/internal/models"
	"id-100/internal/utils"
)

// DerivenHandler displays the list of deriven with pagination and optional city filter
func DerivenHandler(c echo.Context) error {
	stats := utils.GetFooterStats()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	cityFilter := c.QueryParam("city") // Get city filter from query params
	limit := 20
	offset := (page - 1) * limit

	// Get list of all distinct cities from contributions for the filter dropdown
	citiesQuery := `SELECT DISTINCT user_city FROM contributions WHERE user_city IS NOT NULL AND user_city != '' ORDER BY user_city ASC`
	citiesRows, err := database.DB.Query(context.Background(), citiesQuery)
	if err != nil {
		log.Printf("Cities Query Error: %v", err)
	}
	defer citiesRows.Close()

	var cities []string
	for citiesRows.Next() {
		var city string
		if err := citiesRows.Scan(&city); err == nil {
			cities = append(cities, city)
		}
	}

	// Build the count and main query based on city filter
	var totalCount int
	var countQuery string
	var query string

	if cityFilter != "" {
		// Filter by city: only show deriven that have contributions from this city
		countQuery = `SELECT COUNT(DISTINCT d.id) FROM deriven d 
		              INNER JOIN contributions c ON c.derive_id = d.id 
		              WHERE c.user_city = $1`
		err = database.DB.QueryRow(context.Background(), countQuery, cityFilter).Scan(&totalCount)

		query = `
            SELECT 
                d.id, d.number, d.title, d.description, 
                COALESCE(c.image_url, ''), COALESCE(c.image_lqip, ''),
                (SELECT COUNT(*) FROM contributions WHERE derive_id = d.id) as contrib_count,
                d.points
            FROM deriven d
            INNER JOIN contributions city_contrib ON city_contrib.derive_id = d.id AND city_contrib.user_city = $1
            LEFT JOIN LATERAL (
                SELECT image_url, image_lqip FROM contributions 
                WHERE derive_id = d.id 
                ORDER BY created_at DESC LIMIT 1
            ) c ON true
            GROUP BY d.id, d.number, d.title, d.description, c.image_url, c.image_lqip, d.points
            ORDER BY d.number ASC 
            LIMIT $2 OFFSET $3`
	} else {
		// No filter: show all deriven
		countQuery = "SELECT COUNT(*) FROM deriven"
		err = database.DB.QueryRow(context.Background(), countQuery).Scan(&totalCount)

		query = `
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
	}

	if err != nil {
		log.Printf("Count Error: %v", err)
		totalCount = 100 // fallback
	}
	totalPages := (totalCount + limit - 1) / limit // ceiling division

	// Execute the main query
	var rows pgx.Rows
	if cityFilter != "" {
		rows, err = database.DB.Query(context.Background(), query, cityFilter, limit, offset)
	} else {
		rows, err = database.DB.Query(context.Background(), query, limit, offset)
	}

	if err != nil {
		log.Printf("Query Error: %v", err)
		return c.String(http.StatusInternalServerError, "Datenbankfehler")
	}
	defer rows.Close()

	var deriven []models.Derive
	for rows.Next() {
		var d models.Derive
		if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.ImageLqip, &d.ContribCount, &d.Points); err != nil {
			log.Printf("Scan Error: %v", err)
			return err
		}
		// Normalize image URL
		d.ImageUrl = utils.EnsureFullImageURL(d.ImageUrl)
		// map points to a simple tier (1..3) for badge + overlay selection
		if d.Points <= 1 {
			d.PointsTier = 1
		} else if d.Points == 2 {
			d.PointsTier = 2
		} else {
			d.PointsTier = 3
		}
		deriven = append(deriven, d)
	}

	// Build pagination pages for template
	var pages []models.PageNumber

	// Always show first page
	pages = append(pages, models.PageNumber{Number: 1, IsCurrent: page == 1})

	// Show dots if current page > 3
	if page > 3 {
		pages = append(pages, models.PageNumber{IsDots: true})
	}

	// Show page before current (if exists and not page 1 or 2)
	if page > 2 {
		pages = append(pages, models.PageNumber{Number: page - 1, IsCurrent: false})
	}

	// Show current page (if not first or last)
	if page > 1 && page < totalPages {
		pages = append(pages, models.PageNumber{Number: page, IsCurrent: true})
	}

	// Show page after current (if exists and not last page or second to last)
	if page < totalPages-1 {
		pages = append(pages, models.PageNumber{Number: page + 1, IsCurrent: false})
	}

	// Show dots if there's a gap to last page
	if page < totalPages-2 {
		pages = append(pages, models.PageNumber{IsDots: true})
	}

	// Always show last page (if more than 1 page)
	if totalPages > 1 {
		pages = append(pages, models.PageNumber{Number: totalPages, IsCurrent: page == totalPages})
	}

	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Innenstadt (🏠) ID (🆔) - 100 (💯)",
		"Deriven":         deriven,
		"CurrentPage":     page,
		"TotalPages":      totalPages,
		"Pages":           pages,
		"HasNext":         page < totalPages,
		"HasPrev":         page > 1,
		"NextPage":        page + 1,
		"PrevPage":        page - 1,
		"Cities":          cities,
		"SelectedCity":    cityFilter,
		"ContentTemplate": "ids.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// DeriveHandler displays a single derive with its contributions
func DeriveHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	num := c.Param("number")
	pageParam := c.QueryParam("page") // Capture page parameter for back navigation

	var d models.Derive
	query := `
            SELECT d.id, d.number, d.title, d.description, COALESCE(c.image_url, ''), d.points
            FROM deriven d
            LEFT JOIN LATERAL (
                SELECT image_url FROM contributions WHERE derive_id = d.id ORDER BY created_at DESC LIMIT 1
            ) c ON true
            WHERE d.number = $1`

	err := database.DB.QueryRow(context.Background(), query, num).Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.Points)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	// Normalize derive image URL
	d.ImageUrl = utils.EnsureFullImageURL(d.ImageUrl)
	// compute PointsTier for styling
	if d.Points <= 1 {
		d.PointsTier = 1
	} else if d.Points == 2 {
		d.PointsTier = 2
	} else {
		d.PointsTier = 3
	}

	rows, _ := database.DB.Query(context.Background(),
		"SELECT image_url, COALESCE(image_lqip,''), user_name, COALESCE(user_city,''), COALESCE(user_comment,''), created_at FROM contributions WHERE derive_id = $1 ORDER BY created_at DESC", d.ID)
	defer rows.Close()

	var contribs []models.Contribution
	for rows.Next() {
		var ct models.Contribution
		rows.Scan(&ct.ImageUrl, &ct.ImageLqip, &ct.UserName, &ct.UserCity, &ct.UserComment, &ct.CreatedAt)
		// Normalize contribution image URL
		ct.ImageUrl = utils.EnsureFullImageURL(ct.ImageUrl)
		contribs = append(contribs, ct)
	}

	// If requested as a partial (AJAX), return only the detail fragment
	if c.QueryParam("partial") == "1" {
		return c.Render(http.StatusOK, "id_detail.content", map[string]interface{}{
			"Derive":        d,
			"Contributions": contribs,
			"PageParam":     pageParam,
			"IsPartial":     true,
		})
	}

	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           fmt.Sprintf("#%d %s", d.Number, d.Title),
		"Derive":          d,
		"Contributions":   contribs,
		"PageParam":       pageParam,
		"IsPartial":       false,
		"ContentTemplate": "id_detail.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// UploadGetHandler displays the upload form
func UploadGetHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	rows, err := database.DB.Query(context.Background(), `
SELECT d.number, d.title, COALESCE(d.points, 0) as points, COALESCE((SELECT COUNT(*) FROM contributions WHERE derive_id = d.id),0) as contrib_count
FROM deriven d
ORDER BY d.number ASC`)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Datenbankfehler")
	}
	defer rows.Close()

	var list []models.Derive
	for rows.Next() {
		var d models.Derive
		if err := rows.Scan(&d.Number, &d.Title, &d.Points, &d.ContribCount); err != nil {
			return err
		}
		list = append(list, d)
	}

	// Fetch session uploads for this token/session so we can display them under the upload form
	tokenID, _ := c.Get("token_id").(int)
	sessionNumber, _ := c.Get("session_number").(int)
	uRows, err := database.DB.Query(context.Background(), `
		SELECT c.id, d.number, c.image_url, COALESCE(c.image_lqip, '')
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		JOIN deriven d ON d.id = c.derive_id
		WHERE ul.token_id = $1 AND ul.session_number = $2
		ORDER BY ul.uploaded_at DESC
	`, tokenID, sessionNumber)
	if err != nil {
		log.Printf("Failed to fetch session uploads: %v", err)
	}
	defer func() {
		if uRows != nil {
			uRows.Close()
		}
	}()

	var sessionContribs []map[string]interface{}
	for uRows != nil && uRows.Next() {
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
			"image_url":  utils.EnsureFullImageURL(imageUrl),
			"image_lqip": imageLqip,
		})
	}

	token, _ := c.Get("token").(string)
	currentPlayer, _ := c.Get("current_player").(string)

	// Build a map[string]bool of derive numbers that were uploaded in THIS session/token
	uploadedNumbers := make(map[string]bool)
	totalPoints := 0
	for _, sc := range sessionContribs {
		if num, ok := sc["number"].(int); ok {
			uploadedNumbers[strconv.Itoa(num)] = true
			// Find the points for this derive number
			for _, d := range list {
				if d.Number == num {
					totalPoints += d.Points
					break
				}
			}
		}
	}

	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Beweis hochladen - 🏠🆔💯",
		"Deriven":         list,
		"ContentTemplate": "upload.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
		"SessionContribs": sessionContribs,
		"SelectedNumber":  c.QueryParam("number"),
		"Token":           token,
		"CurrentPlayer":   currentPlayer,
		"UploadedNumbers": uploadedNumbers,
		"TotalPoints":     totalPoints,
	}))
}

// UploadPostHandler handles image upload
func UploadPostHandler(c echo.Context) error {
	// Get token info from middleware context
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.String(http.StatusForbidden, "Token nicht gefunden")
	}

	currentPlayer, _ := c.Get("current_player").(string)
	sessionNumber, _ := c.Get("session_number").(int)

	deriveNumberStr := c.FormValue("derive_number")
	deriveNumber, err := strconv.Atoi(deriveNumberStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Ungültige Aufgabennummer")
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.String(http.StatusBadRequest, "Kein Bild gefunden")
	}

	src, _ := file.Open()
	defer src.Close()
	// Decode and auto-orient based on EXIF so mobile uploads keep the correct rotation
	img, err := imgutil.DecodeAutoOriented(src)
	if err != nil {
		return c.String(http.StatusBadRequest, "Ungültiges Bildformat oder Korrektur fehlgeschlagen")
	}
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 75}); err != nil {
		return c.String(http.StatusInternalServerError, "WebP-Kodierung fehlgeschlagen")
	}

	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("S3_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			""),
		),
	)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	fileName := fmt.Sprintf("derive_%s_%d.webp", deriveNumberStr, time.Now().Unix())

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET")),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/webp"),
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "S3 Fehler: "+err.Error())
	}

	// Store relative path in DB, EnsureFullImageURL will add the base URL when reading
	relativePath := fmt.Sprintf("/storage/v1/object/public/%s/%s", os.Getenv("S3_BUCKET"), fileName)

	// generate tiny LQIP (data-uri) and store it
	lqip, lqipErr := utils.GenerateLQIP(img, 24)
	if lqipErr != nil {
		log.Printf("LQIP generation failed: %v", lqipErr)
		lqip = ""
	}

	var internalID int
	err = database.DB.QueryRow(context.Background(),
		"SELECT id FROM deriven WHERE number = $1", deriveNumberStr).Scan(&internalID)
	if err != nil {
		return c.String(http.StatusNotFound, "Aufgabe nicht gefunden")
	}

	// Get optional user comment (max 100 chars)
	userComment := c.FormValue("comment")
	runes := []rune(userComment)
	if len(runes) > 100 {
		userComment = string(runes[:100])
	}

	// Insert contribution and get ID
	var contributionID int
	currentPlayerCity, _ := c.Get("current_player_city").(string)
	err = database.DB.QueryRow(context.Background(),
		"INSERT INTO contributions (derive_id, image_url, image_lqip, user_name, user_city, user_comment) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		internalID, relativePath, lqip, currentPlayer, currentPlayerCity, userComment).Scan(&contributionID)

	if err != nil {
		log.Printf("DB Error inserting contribution: %v", err)
		return c.String(http.StatusInternalServerError, "DB Error")
	}

	// Log upload in upload_logs table
	_, err = database.DB.Exec(context.Background(),
		`INSERT INTO upload_logs (token_id, derive_number, player_name, session_number, contribution_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		tokenID, deriveNumber, currentPlayer, sessionNumber, contributionID)

	if err != nil {
		log.Printf("Failed to log upload: %v", err)
		// Don't fail the request, contribution is already saved
	}

	// Increment total_uploads counter for token
	_, err = database.DB.Exec(context.Background(),
		"UPDATE upload_tokens SET total_uploads = total_uploads + 1 WHERE id = $1",
		tokenID)

	if err != nil {
		log.Printf("Failed to increment upload counter: %v", err)
		// Don't fail the request, contribution is already saved
	}

	// Redirect back to the upload page with an uploaded flag so the client can
	// clear the derive selection and show a success message.
	// Only propagate a token if the original client request actually provided one
	// (either via query string or a form-encoded body). Do NOT leak cookie/session tokens.
	redirectURL := "/upload?uploaded=1"
	// Prefer raw query param (avoids parsing body/multipart)
	originalToken := c.Request().URL.Query().Get("token")
	if originalToken == "" && c.Request().Method == "POST" {
		contentType := c.Request().Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// Limit body size before parsing, consistent with middleware
			const maxFormSize = int64(2 * 1024 * 1024) // 2 MiB
			c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxFormSize)
			if formToken := c.FormValue("token"); formToken != "" {
				originalToken = formToken
			}
		}
	}
	if originalToken != "" {
		redirectURL = fmt.Sprintf("%s&token=%s", redirectURL, url.QueryEscape(originalToken))
	}
	return c.Redirect(http.StatusSeeOther, redirectURL)
}

// RulesHandler displays the rules page
func RulesHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Leitfaden - 🏠🆔💯",
		"ContentTemplate": "leitfaden.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// ImpressumHandler displays the impressum page
func ImpressumHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Impressum - 🏠🆔💯",
		"ContentTemplate": "impressum.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// DatenschutzHandler displays the privacy policy page
func DatenschutzHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Datenschutzerklärung - 🏠🆔💯",
		"ContentTemplate": "datenschutz.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// RequestBagHandler displays the bag request form
func RequestBagHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	if c.QueryParam("partial") == "1" {
		return c.Render(http.StatusOK, "request_bag.content", map[string]interface{}{
			"CurrentPath": c.Request().URL.Path,
			"CurrentYear": time.Now().Year(),
			"FooterStats": stats,
			"IsPartial":   true,
		})
	}
	return c.Render(http.StatusOK, "layout", MergeTemplateData(map[string]interface{}{
		"Title":           "Werkzeug anfordern - 🏠🆔💯",
		"ContentTemplate": "request_bag.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// RequestBagPostHandler handles bag request submissions
func RequestBagPostHandler(c echo.Context) error {
	type payload struct {
		Email string `json:"email"`
	}
	var p payload
	if strings.Contains(c.Request().Header.Get("Content-Type"), "application/json") {
		if err := c.Bind(&p); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ungültiger Request"})
		}
	} else {
		p.Email = c.FormValue("email")
	}
	email := strings.TrimSpace(p.Email)
	if email == "" || !strings.Contains(email, "@") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ungültige E-Mail"})
	}

	_, err := database.DB.Exec(context.Background(), "INSERT INTO bag_requests (email) VALUES ($1)", email)
	if err != nil {
		log.Printf("Failed to insert bag request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Serverfehler"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// SetPlayerNameHandler handles the name entry form submission
func SetPlayerNameHandler(c echo.Context) error {
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
		_ = database.DB.QueryRow(context.Background(), "SELECT COALESCE(bag_name,'') FROM upload_tokens WHERE token = $1", token).Scan(&bagName)
		return c.Render(http.StatusBadRequest, "layout", MergeTemplateData(map[string]interface{}{
			"Title":           "Willkommen bei ID-100!",
			"ContentTemplate": "enter_name.content",
			"Token":           token,
			"BagName":         bagName,
			"FormError":       "Bitte bestätige die Datenschutzerklärung und dass du keine erkennbaren Personen ohne Einwilligung hochlädst.",
		}))
	}

	playerCity := strings.TrimSpace(c.FormValue("player_city"))

	// Save name and city in session
	session, _ := middleware.Store.Get(c.Request(), "id-100-session")
	session.Values["player_name"] = playerName
	session.Values["player_city"] = playerCity
	session.Save(c.Request(), c.Response())

	// Update database with city
	_, err := database.DB.Exec(context.Background(),
		"UPDATE upload_tokens SET current_player = $1, current_player_city = $2, session_started_at = NOW() WHERE token = $3",
		playerName, playerCity, token)

	if err != nil {
		log.Printf("Error setting player name: %v", err)
	}

	// Redirect to upload page
	return c.Redirect(http.StatusSeeOther, "/upload?token="+token)
}

// UserDeleteContributionHandler allows users to delete their own contributions from the current session
func UserDeleteContributionHandler(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.Atoi(contributionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}

	// Get token info from middleware context
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Token not found"})
	}

	sessionNumber, _ := c.Get("session_number").(int)

	// Verify that this contribution belongs to the current user's session
	var imageURL string
	var uploadLogTokenID, uploadLogSessionNumber int
	err = database.DB.QueryRow(context.Background(), `
		SELECT c.image_url, ul.token_id, ul.session_number
		FROM contributions c
		JOIN upload_logs ul ON ul.contribution_id = c.id
		WHERE c.id = $1
	`, contributionID).Scan(&imageURL, &uploadLogTokenID, &uploadLogSessionNumber)

	if err != nil {
		log.Printf("Failed to fetch contribution: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	// Check if the contribution belongs to the current session
	if uploadLogTokenID != tokenID || uploadLogSessionNumber != sessionNumber {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You can only delete your own uploads from this session"})
	}

	// Delete from upload_logs first
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
		"message": "Upload deleted successfully",
	})
}
