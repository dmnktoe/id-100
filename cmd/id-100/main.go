package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var store *sessions.CookieStore
var baseURL string

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
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

type Derive struct {
	ID           int    `json:"id"`
	Number       int    `json:"number"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageUrl     string `json:"image_url"`
	ImageLqip    string `json:"image_lqip"`
	ContribCount int    `json:"contrib_count"`
	// Points assigned to the derive (used for badges and overlay selection)
	Points int `json:"points"`
	// PointsTier maps points to 1..3 for styling purposes
	PointsTier int `json:"points_tier"`
}

// ensureFullImageURL is implemented in cmd/id-100/utils.go to keep main.go smaller.

func main() {
	initDatabase()
	defer db.Close()

	// Load environment variables
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	isProduction := os.Getenv("ENVIRONMENT") == "production"

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		if isProduction {
			log.Fatal("SESSION_SECRET must be set in production. Generate one with: openssl rand -base64 32")
		}
		log.Println("WARNING: Using insecure default SESSION_SECRET. Set SESSION_SECRET environment variable.")
		sessionSecret = "id-100-secret-key-change-in-production"
	}

	// Initialize session store with secure settings
	store = sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,           // 30 days
		HttpOnly: true,                 // Prevent JavaScript access
		Secure:   isProduction,         // HTTPS only in production
		SameSite: http.SameSiteLaxMode, // CSRF protection
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// load templates (moved to helper)
	t := LoadTemplates()
	e.Renderer = t

	// register routes in routes.go
	registerRoutes(e)
	// routes are registered in routes.go

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
