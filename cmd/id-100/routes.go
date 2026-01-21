package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chai2010/webp"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo) {
	e.Static("/static", "web/static")

	e.GET("/", homeHandler)
	e.GET("/deriven", derivenHandler)
	e.GET("/derive/:number", deriveHandler)
	
	// Upload routes - protected by token middleware with session support
	e.GET("/upload", uploadGetHandler, tokenMiddlewareWithSession)
	e.POST("/upload", uploadPostHandler, tokenMiddlewareWithSession)
	e.POST("/upload/set-name", setPlayerNameHandler, tokenMiddlewareWithSession)
	
	e.GET("/spielregeln", rulesHandler)
	e.GET("/about", aboutHandler)
	
	// Admin routes for token management
	adminGroup := e.Group("/admin", basicAuthMiddleware)
	adminGroup.GET("", adminDashboardHandler)
	adminGroup.GET("/tokens", adminTokenListHandler)
	adminGroup.POST("/tokens", adminCreateTokenHandler)
	adminGroup.POST("/tokens/:id/deactivate", adminTokenDeactivateHandler)
	adminGroup.POST("/tokens/:id/reset", adminTokenResetHandler)
	adminGroup.POST("/tokens/:id/assign", adminTokenAssignHandler)
	adminGroup.POST("/tokens/:id/quota", adminUpdateQuotaHandler)
	adminGroup.GET("/tokens/:id/qr", adminDownloadQRHandler)
}

func homeHandler(c echo.Context) error {
	stats := getFooterStats()
	
	// fetch latest contributions with derive meta
	rows, err := db.Query(context.Background(), `
		SELECT c.image_url, COALESCE(c.image_lqip,''), c.user_name, c.created_at, d.number, d.title
		FROM contributions c
		JOIN deriven d ON d.id = c.derive_id
		ORDER BY c.created_at DESC
		LIMIT $1`, 5)
	if err != nil {
		log.Printf("Query Error (recent contributions): %v", err)
		// fallback render with empty list
		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "🏠🆔💯 DÉRIVE 100",
			"RecentContribs":  []interface{}{},
			"ContentTemplate": "index.content",
			"CurrentPath":     c.Request().URL.Path,
			"CurrentYear":     time.Now().Year(),
			"FooterStats":     stats,
		})
	}
	defer rows.Close()

	type RecentContribution struct {
		ImageUrl  string
		ImageLqip string
		UserName  string
		CreatedAt time.Time
		Number    int
		Title     string
	}
	var recent []RecentContribution
	for rows.Next() {
		var r RecentContribution
		if err := rows.Scan(&r.ImageUrl, &r.ImageLqip, &r.UserName, &r.CreatedAt, &r.Number, &r.Title); err != nil {
			log.Printf("Scan Error: %v", err)
			continue
		}
		// Normalize image URL so templates can use it as-is
		r.ImageUrl = ensureFullImageURL(r.ImageUrl)
		recent = append(recent, r)
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "🏠🆔💯 DÉRIVE 100",
		"RecentContribs":  recent,
		"ContentTemplate": "index.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	})
}

func derivenHandler(c echo.Context) error {
	stats := getFooterStats()
	
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// Get total count for pagination
	var totalCount int
	err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM deriven").Scan(&totalCount)
	if err != nil {
		log.Printf("Count Error: %v", err)
		totalCount = 100 // fallback
	}
	totalPages := (totalCount + limit - 1) / limit // ceiling division

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

	rows, err := db.Query(context.Background(), query, limit, offset)
	if err != nil {
		log.Printf("Query Error: %v", err)
		return c.String(http.StatusInternalServerError, "Datenbankfehler")
	}
	defer rows.Close()

	var deriven []Derive
	for rows.Next() {
		var d Derive
		if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.ImageLqip, &d.ContribCount, &d.Points); err != nil {
			log.Printf("Scan Error: %v", err)
			return err
		}
		// Normalize image URL
		d.ImageUrl = ensureFullImageURL(d.ImageUrl)
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
	type PageNumber struct {
		Number    int
		IsCurrent bool
		IsDots    bool
	}
	var pages []PageNumber

	// Always show first page
	pages = append(pages, PageNumber{Number: 1, IsCurrent: page == 1})

	// Show dots if current page > 3
	if page > 3 {
		pages = append(pages, PageNumber{IsDots: true})
	}

	// Show page before current (if exists and not page 1 or 2)
	if page > 2 {
		pages = append(pages, PageNumber{Number: page - 1, IsCurrent: false})
	}

	// Show current page (if not first or last)
	if page > 1 && page < totalPages {
		pages = append(pages, PageNumber{Number: page, IsCurrent: true})
	}

	// Show page after current (if exists and not last page or second to last)
	if page < totalPages-1 {
		pages = append(pages, PageNumber{Number: page + 1, IsCurrent: false})
	}

	// Show dots if there's a gap to last page
	if page < totalPages-2 {
		pages = append(pages, PageNumber{IsDots: true})
	}

	// Always show last page (if more than 1 page)
	if totalPages > 1 {
		pages = append(pages, PageNumber{Number: totalPages, IsCurrent: page == totalPages})
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Index - DÉRIVE 100",
		"Deriven":         deriven,
		"CurrentPage":     page,
		"TotalPages":      totalPages,
		"Pages":           pages,
		"HasNext":         page < totalPages,
		"HasPrev":         page > 1,
		"NextPage":        page + 1,
		"PrevPage":        page - 1,
		"ContentTemplate": "deriven.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	})
}

func deriveHandler(c echo.Context) error {
	stats := getFooterStats()
	num := c.Param("number")
	pageParam := c.QueryParam("page") // Capture page parameter for back navigation
	
	var d Derive
	query := `
            SELECT d.id, d.number, d.title, d.description, COALESCE(c.image_url, ''), d.points
            FROM deriven d
            LEFT JOIN LATERAL (
                SELECT image_url FROM contributions WHERE derive_id = d.id ORDER BY created_at DESC LIMIT 1
            ) c ON true
            WHERE d.number = $1`

	err := db.QueryRow(context.Background(), query, num).Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.Points)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/deriven")
	}
	// Normalize derive image URL
	d.ImageUrl = ensureFullImageURL(d.ImageUrl)
	// compute PointsTier for styling
	if d.Points <= 1 {
		d.PointsTier = 1
	} else if d.Points == 2 {
		d.PointsTier = 2
	} else {
		d.PointsTier = 3
	}
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/deriven")
	}

	// Normalize derive image URL
	d.ImageUrl = ensureFullImageURL(d.ImageUrl)

	rows, _ := db.Query(context.Background(),
		"SELECT image_url, COALESCE(image_lqip,''), user_name, created_at FROM contributions WHERE derive_id = $1 ORDER BY created_at DESC", d.ID)
	defer rows.Close()

	type Contribution struct {
		ImageUrl  string
		ImageLqip string
		UserName  string
		CreatedAt time.Time
	}
	var contribs []Contribution
	for rows.Next() {
		var ct Contribution
		rows.Scan(&ct.ImageUrl, &ct.ImageLqip, &ct.UserName, &ct.CreatedAt)
		// Normalize contribution image URL
		ct.ImageUrl = ensureFullImageURL(ct.ImageUrl)
		contribs = append(contribs, ct)
	}

	// If requested as a partial (AJAX), return only the detail fragment
	if c.QueryParam("partial") == "1" {
		return c.Render(http.StatusOK, "derive_detail.content", map[string]interface{}{
			"Derive":        d,
			"Contributions": contribs,
			"PageParam":     pageParam,
			"IsPartial":     true,
		})
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           fmt.Sprintf("#%d %s", d.Number, d.Title),
		"Derive":          d,
		"Contributions":   contribs,
		"PageParam":       pageParam,
		"IsPartial":       false,
		"ContentTemplate": "derive_detail.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	})
}

func uploadGetHandler(c echo.Context) error {
	stats := getFooterStats()
	rows, err := db.Query(context.Background(), "SELECT number, title FROM deriven ORDER BY number ASC")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Datenbankfehler")
	}
	defer rows.Close()

	var list []Derive
	for rows.Next() {
		var d Derive
		if err := rows.Scan(&d.Number, &d.Title); err != nil {
			return err
		}
		list = append(list, d)
	}
	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Submit Evidence - DÉRIVE 100",
		"Deriven":         list,
		"ContentTemplate": "upload.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	})
}

func uploadPostHandler(c echo.Context) error {
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
	img, _, err := image.Decode(src)
	if err != nil {
		return c.String(http.StatusBadRequest, "Ungültiges Bildformat")
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

	// Store relative path in DB, ensureFullImageURL will add the base URL when reading
	relativePath := fmt.Sprintf("/storage/v1/object/public/%s/%s", os.Getenv("S3_BUCKET"), fileName)

	// generate tiny LQIP (data-uri) and store it
	lqip, lqipErr := generateLQIP(img, 24)
	if lqipErr != nil {
		log.Printf("LQIP generation failed: %v", lqipErr)
		lqip = ""
	}

	var internalID int
	err = db.QueryRow(context.Background(),
		"SELECT id FROM deriven WHERE number = $1", deriveNumberStr).Scan(&internalID)
	if err != nil {
		return c.String(http.StatusNotFound, "Aufgabe nicht gefunden")
	}

	// Insert contribution and get ID
	var contributionID int
	err = db.QueryRow(context.Background(),
		"INSERT INTO contributions (derive_id, image_url, image_lqip, user_name) VALUES ($1, $2, $3, $4) RETURNING id",
		internalID, relativePath, lqip, currentPlayer).Scan(&contributionID)

	if err != nil {
		log.Printf("DB Error inserting contribution: %v", err)
		return c.String(http.StatusInternalServerError, "DB Error")
	}

	// Log upload in upload_logs table
	_, err = db.Exec(context.Background(),
		`INSERT INTO upload_logs (token_id, derive_number, player_name, session_number, contribution_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		tokenID, deriveNumber, currentPlayer, sessionNumber, contributionID)
	
	if err != nil {
		log.Printf("Failed to log upload: %v", err)
		// Don't fail the request, contribution is already saved
	}
	
	// Increment total_uploads counter for token
	_, err = db.Exec(context.Background(),
		"UPDATE upload_tokens SET total_uploads = total_uploads + 1 WHERE id = $1",
		tokenID)
	
	if err != nil {
		log.Printf("Failed to increment upload counter: %v", err)
		// Don't fail the request, contribution is already saved
	}

	return c.Redirect(http.StatusSeeOther, "/deriven")
}

func rulesHandler(c echo.Context) error {
	stats := getFooterStats()
	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Spielregeln - DÉRIVE 100",
		"ContentTemplate": "spielregeln.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	})
}

func aboutHandler(c echo.Context) error {
	stats := getFooterStats()
	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "About - DÉRIVE 100",
		"ContentTemplate": "about.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	})
}
