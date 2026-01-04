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
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo) {
	e.Static("/static", "web/static")

	e.GET("/", homeHandler)
	e.GET("/deriven", derivenHandler)
	e.GET("/derive/:number", deriveHandler)
	e.GET("/upload", uploadGetHandler)
	e.POST("/upload", uploadPostHandler)
	e.GET("/spielregeln", rulesHandler)
	e.GET("/about", aboutHandler)
}

func homeHandler(c echo.Context) error {
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
	})
}

func derivenHandler(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	query := `
            SELECT 
                d.id, d.number, d.title, d.description, 
                COALESCE(c.image_url, ''), COALESCE(c.image_lqip, ''),
                (SELECT COUNT(*) FROM contributions WHERE derive_id = d.id) as contrib_count
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
		if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.ImageLqip, &d.ContribCount); err != nil {
			log.Printf("Scan Error: %v", err)
			return err
		}
		// Normalize image URL
		d.ImageUrl = ensureFullImageURL(d.ImageUrl)
		deriven = append(deriven, d)
	}

	// debug: log types/values to debug template field issue
	for i, d := range deriven {
		log.Printf("derive[%d] type=%T ImageLqip=%q", i, d, d.ImageLqip)
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Index - DÉRIVE 100",
		"Deriven":         deriven,
		"CurrentPage":     page,
		"HasNext":         len(deriven) == limit,
		"HasPrev":         page > 1,
		"NextPage":        page + 1,
		"PrevPage":        page - 1,
		"ContentTemplate": "deriven.content",
		"CurrentPath":     c.Request().URL.Path,
	})
}

func deriveHandler(c echo.Context) error {
	num := c.Param("number")
	var d Derive
	query := `
            SELECT d.id, d.number, d.title, d.description, COALESCE(c.image_url, '')
            FROM deriven d
            LEFT JOIN LATERAL (
                SELECT image_url FROM contributions WHERE derive_id = d.id ORDER BY created_at DESC LIMIT 1
            ) c ON true
            WHERE d.number = $1`

	err := db.QueryRow(context.Background(), query, num).Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/deriven")
	}

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
		})
	}

	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           fmt.Sprintf("#%d %s", d.Number, d.Title),
		"Derive":          d,
		"Contributions":   contribs,
		"ContentTemplate": "derive_detail.content",
		"CurrentPath":     c.Request().URL.Path,
	})
}

func uploadGetHandler(c echo.Context) error {
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
	})
}

func uploadPostHandler(c echo.Context) error {
	deriveNumberStr := c.FormValue("derive_number")
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
	options, _ := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	webp.Encode(&buf, img, options)

	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("S3_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			""),
		),
	)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT"))
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

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		os.Getenv("SUPABASE_URL"), os.Getenv("S3_BUCKET"), fileName)

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

	_, err = db.Exec(context.Background(),
		"INSERT INTO contributions (derive_id, image_url, image_lqip, user_name) VALUES ($1, $2, $3, $4)",
		internalID, publicURL, lqip, "Anonym")

	if err != nil {
		log.Printf("DB Error: %v", err)
		return c.String(http.StatusInternalServerError, "DB Error")
	}

	return c.Redirect(http.StatusSeeOther, "/deriven")
}

func rulesHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Regeln - DÉRIVE 100",
		"ContentTemplate": "spielregeln.content",
		"CurrentPath":     c.Request().URL.Path,
	})
}

func aboutHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "layout", map[string]interface{}{
		"Title":           "Über - DÉRIVE 100",
		"ContentTemplate": "about.content",
		"CurrentPath":     c.Request().URL.Path,
	})
}
