package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"  // Wichtig: Registriert GIF-Decoder
	_ "image/jpeg" // Wichtig: Registriert JPEG-Decoder
	_ "image/png"  // Wichtig: Registriert PNG-Decoder
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	storage_go "github.com/supabase-community/storage-go"
)

// Template Renderer für Echo
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Datenstruktur für eine Aufgabe (Derive)
type Derive struct {
	ID          int    `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImagePath   string `json:"image_path"`
}

func main() {
	// 1. DB Verbindung initialisieren (Funktion kommt aus deiner database.go)
	initDatabase()
	defer db.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 2. Templates laden (Inklusive Komponenten)
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

	// Statische Dateien (CSS, JS)
	e.Static("/static", "web/static")

	// --- ROUTES ---

	// LANDING PAGE
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Title": "🏠🆔💯 DÉRIVE 100",
		})
	})

	// LISTE ALLER DERIVEN (Paginierung)
	e.GET("/deriven", func(c echo.Context) error {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page < 1 { page = 1 }
		limit := 20
		offset := (page - 1) * limit

		rows, err := db.Query(context.Background(), 
			"SELECT id, number, title, description, COALESCE(image_path, '') FROM deriven ORDER BY number ASC LIMIT $1 OFFSET $2", 
			limit, offset)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Datenbankfehler")
		}
		defer rows.Close()

		var deriven []Derive
		for rows.Next() {
			var d Derive
			if err := rows.Scan(&d.ID, &d.Number, &d.Title, &d.Description, &d.ImagePath); err != nil {
				return err
			}
			deriven = append(deriven, d)
		}

		return c.Render(http.StatusOK, "deriven.html", map[string]interface{}{
			"Title":       "Index - DÉRIVE 100",
			"Deriven":     deriven,
			"CurrentPage": page,
			"HasNext":     page < 5,
			"HasPrev":     page > 1,
			"NextPage":    page + 1,
			"PrevPage":    page - 1,
		})
	})

	// UPLOAD FORMULAR ANZEIGEN
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

	// UPLOAD VERARBEITEN (WebP + Storage)
	e.POST("/upload", func(c echo.Context) error {
		deriveNumber := c.FormValue("derive_number")
		file, err := c.FormFile("image")
		if err != nil {
			return c.String(http.StatusBadRequest, "Kein Bild gefunden")
		}

		// A. Bild dekodieren
		src, _ := file.Open()
		defer src.Close()
		img, _, err := image.Decode(src)
		if err != nil {
			return c.String(http.StatusBadRequest, "Ungültiges Bildformat")
		}

		// B. WebP Konvertierung
		var buf bytes.Buffer
		options, _ := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err := webp.Encode(&buf, img, options); err != nil {
			return c.String(http.StatusInternalServerError, "Optimierung fehlgeschlagen")
		}

		// C. Supabase Storage Upload
		storageClient := storage_go.NewClient(
			os.Getenv("SUPABASE_URL"), 
			os.Getenv("SUPABASE_SERVICE_ROLE_KEY"), 
			nil,
		)

		fileName := fmt.Sprintf("derive_%s.webp", deriveNumber)
		_, err = storageClient.UploadFile("solutions", fileName, bytes.NewReader(buf.Bytes()))
		if err != nil {
			return c.String(http.StatusInternalServerError, "Storage Upload Error")
		}

		// D. Link in DB speichern
		publicURL := fmt.Sprintf("%s/storage/v1/object/public/solutions/%s", os.Getenv("SUPABASE_URL"), fileName)
		_, err = db.Exec(context.Background(), 
			"UPDATE deriven SET image_path = $1 WHERE number = $2", 
			publicURL, deriveNumber)

		if err != nil {
			return c.String(http.StatusInternalServerError, "Datenbank Update fehlgeschlagen")
		}

		return c.Redirect(http.StatusSeeOther, "/deriven")
	})

	// SPIELREGELN
	e.GET("/spielregeln", func(c echo.Context) error {
		return c.Render(http.StatusOK, "spielregeln.html", map[string]interface{}{
			"Title": "Regeln - DÉRIVE 100",
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}