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
	// Wenn ein Content-Template angegeben ist, rendern wir es zuerst und injizieren
	// das Ergebnis als unescaped HTML in `ContentHTML`, bevor wir das Layout ausgeben.
	if m, ok := data.(map[string]interface{}); ok {
		if ct, ok := m["ContentTemplate"].(string); ok && ct != "" {
			var buf bytes.Buffer
			if err := t.templates.ExecuteTemplate(&buf, ct, m); err != nil {
				return err
			}
			m["ContentHTML"] = template.HTML(buf.String())
		}
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

// Datenstruktur
type Derive struct {
	ID           int    `json:"id"`
	Number       int    `json:"number"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageUrl     string `json:"image_url"`
	ContribCount int    `json:"contrib_count"`
}

func main() {
	// 1. DB Verbindung initialisieren (Funktion muss in deiner database.go stehen)
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

	// Parse templates with a small FuncMap (eq) used for active nav highlighting
	funcs := template.FuncMap{
		"eq": func(a, b string) bool { return a == b },
	}
	tmpl := template.New("").Funcs(funcs)
	tmpls, err := tmpl.ParseFiles(files...)
	if err != nil {
		log.Fatalf("failed to parse templates %v: %v", files, err)
	}
	t := &Template{templates: tmpls}
	e.Renderer = t

	// Statische Dateien
	e.Static("/static", "web/static")

	// --- ROUTES ---

	// HOME
	e.GET("/", func(c echo.Context) error {
		assignments := make([]int, 100)
		for i := range assignments {
			assignments[i] = i + 1
		}

		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "🏠🆔💯 DÉRIVE 100",
			"Assignments":     assignments,
			"ContentTemplate": "index.content",
			"CurrentPath":      c.Request().URL.Path,
		})
	})

	// LISTE ALLER DERIVEN
	e.GET("/deriven", func(c echo.Context) error {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page < 1 {
			page = 1
		}
		limit := 20
		offset := (page - 1) * limit

		query := `
            SELECT 
                d.id, d.number, d.title, d.description, 
                COALESCE(c.image_url, ''),
                (SELECT COUNT(*) FROM contributions WHERE derive_id = d.id) as contrib_count
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
			// Scan von 6 Feldern passend zum SELECT
			if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImageUrl, &d.ContribCount); err != nil {
				log.Printf("Scan Error: %v", err)
				return err
			}
			deriven = append(deriven, d)
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
			"CurrentPath":      c.Request().URL.Path,
		})
	})

	// DETAILSEITE
	e.GET("/derive/:number", func(c echo.Context) error {
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
			"SELECT image_url, user_name, created_at FROM contributions WHERE derive_id = $1 ORDER BY created_at DESC", d.ID)
		defer rows.Close()

		type Contribution struct {
			ImageUrl  string
			UserName  string
			CreatedAt time.Time
		}
		var contribs []Contribution
		for rows.Next() {
			var ct Contribution
			rows.Scan(&ct.ImageUrl, &ct.UserName, &ct.CreatedAt)
			contribs = append(contribs, ct)
		}

		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           fmt.Sprintf("#%d %s", d.Number, d.Title),
			"Derive":          d,
			"Contributions":   contribs,
			"ContentTemplate": "derive_detail.content",
			"CurrentPath":      c.Request().URL.Path,
		})
	})

	// UPLOAD FORMULAR
	e.GET("/upload", func(c echo.Context) error {
		rows, err := db.Query(context.Background(), "SELECT number, title FROM deriven ORDER BY number ASC")
		if err != nil {
			return c.String(http.StatusInternalServerError, "Datenbankfehler")
		}
		defer rows.Close()

		var list []Derive
		for rows.Next() {
			var d Derive
			// Hier nur 2 Felder scannen, da SELECT nur number, title holt
			if err := rows.Scan(&d.Number, &d.Title); err != nil {
				return err
			}
			list = append(list, d)
		}
		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "Submit Evidence - DÉRIVE 100",
			"Deriven":         list,
			"ContentTemplate": "upload.content",
			"CurrentPath":      c.Request().URL.Path,
		})
	})

	// POST UPLOAD LOGIK
	e.POST("/upload", func(c echo.Context) error {
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

		var internalID int
		err = db.QueryRow(context.Background(),
			"SELECT id FROM deriven WHERE number = $1", deriveNumberStr).Scan(&internalID)
		if err != nil {
			return c.String(http.StatusNotFound, "Aufgabe nicht gefunden")
		}

		_, err = db.Exec(context.Background(),
			"INSERT INTO contributions (derive_id, image_url, user_name) VALUES ($1, $2, $3)",
			internalID, publicURL, "Anonym")

		if err != nil {
			log.Printf("DB Error: %v", err)
			return c.String(http.StatusInternalServerError, "DB Error")
		}

		return c.Redirect(http.StatusSeeOther, "/deriven")
	})

	e.GET("/spielregeln", func(c echo.Context) error {
		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "Regeln - DÉRIVE 100",
			"ContentTemplate": "spielregeln.content",
			"CurrentPath":      c.Request().URL.Path,
		})
	})

	// ABOUT
	e.GET("/about", func(c echo.Context) error {
		return c.Render(http.StatusOK, "layout", map[string]interface{}{
			"Title":           "Über - DÉRIVE 100",
			"ContentTemplate": "about.content",
			"CurrentPath":      c.Request().URL.Path,
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
