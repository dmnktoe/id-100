package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Template Renderer für Echo
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Datenstruktur angepasst an JOIN-Abfrage
type Derive struct {
	ID          int    `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"` // Jetzt aus der contributions Tabelle
}

func main() {
	// 1. DB Verbindung initialisieren
	initDatabase()
	defer db.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 2. Templates laden
	files, err := filepath.Glob("web/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	comps, _ := filepath.Glob("web/templates/components/*.html")
	files = append(files, comps...)
	
	t := &Template{
		templates: template.Must(template.ParseFiles(files...)),
	}
	e.Renderer = t

	// Statische Dateien
	e.Static("/static", "web/static")

	// --- ROUTES ---

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Title": "🏠🆔💯 DÉRIVE 100",
		})
	})

	// LISTE ALLER DERIVEN (Mit JOIN auf Contributions)
	e.GET("/deriven", func(c echo.Context) error {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page < 1 { page = 1 }
		limit := 20
		offset := (page - 1) * limit

		// Query nutzt einen Lateral Join, um pro Aufgabe das aktuellste Bild zu finden
		query := `
			SELECT d.id, d.number, d.title, d.description, COALESCE(c.image_url, '') 
			FROM deriven d
			LEFT JOIN LATERAL (
				SELECT image_url FROM contributions 
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
			if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl); err != nil {
				return err
			}
			deriven = append(deriven, d)
		}

		return c.Render(http.StatusOK, "deriven.html", map[string]interface{}{
			"Title":       "Index - DÉRIVE 100",
			"Deriven":     deriven,
			"CurrentPage": page,
			"HasNext":     len(deriven) == limit,
			"HasPrev":     page > 1,
			"NextPage":    page + 1,
			"PrevPage":    page - 1,
		})
	})

	e.GET("/upload", func(c echo.Context) error {
		rows, _ := db.Query(context.Background(), "SELECT number, title FROM deriven ORDER BY number ASC")
		var list []Derive
		for rows.Next() {
			var d Derive
			rows.Scan(&d.Number, &d.Title)
			list = append(list, d)
		}
		return c.Render(http.StatusOK, "upload.html", map[string]interface{}{
			"Title":   "Submit Evidence - DÉRIVE 100",
			"Deriven": list,
		})
	})

	e.POST("/upload", func(c echo.Context) error {
		deriveNumberStr := c.FormValue("derive_number")
		file, err := c.FormFile("image")
		if err != nil {
			return c.String(http.StatusBadRequest, "Kein Bild gefunden")
		}

		// A. Bild dekodieren & WebP Konvertierung
		src, _ := file.Open()
		defer src.Close()
		img, _, err := image.Decode(src)
		if err != nil {
			return c.String(http.StatusBadRequest, "Ungültiges Bildformat")
		}
		var buf bytes.Buffer
		options, _ := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err := webp.Encode(&buf, img, options); err != nil {
			return c.String(http.StatusInternalServerError, "WebP Fehler")
		}

		// B. S3 Client Setup aus ENV
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
		
		// C. S3 Upload
		_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(os.Getenv("S3_BUCKET")),
			Key:         aws.String(fileName),
			Body:        bytes.NewReader(buf.Bytes()),
			ContentType: aws.String("image/webp"),
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, "S3 Fehler: "+err.Error())
		}

		// D. DATENBANK LOGIK
		publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", 
			os.Getenv("SUPABASE_URL"), os.Getenv("S3_BUCKET"), fileName)

		// 1. Die interne ID der Derive-Aufgabe anhand der Nummer holen
		var internalID int
		err = db.QueryRow(context.Background(), 
			"SELECT id FROM deriven WHERE number = $1", deriveNumberStr).Scan(&internalID)
		if err != nil {
			return c.String(http.StatusNotFound, "Aufgabe nicht gefunden")
		}

		// 2. Neue Zeile in CONTRIBUTIONS einfügen
		_, err = db.Exec(context.Background(), 
			"INSERT INTO contributions (derive_id, image_url, user_name) VALUES ($1, $2, $3)", 
			internalID, publicURL, "Anonym")

		if err != nil {
			log.Printf("DB Error: %v", err)
			return c.String(http.StatusInternalServerError, "Datenbankfehler beim Speichern der Contribution")
		}

		return c.Redirect(http.StatusSeeOther, "/deriven")
	})

	e.GET("/spielregeln", func(c echo.Context) error {
		return c.Render(http.StatusOK, "spielregeln.html", map[string]interface{}{
			"Title": "Regeln - DÉRIVE 100",
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}