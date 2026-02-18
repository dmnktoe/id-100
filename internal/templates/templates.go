package templates

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"strings"

	"id-100/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// Renderer is the template renderer for Echo
type Renderer struct {
	templates *template.Template
	minifier  *minify.M
	config    *config.Config
}

// Render renders a template with data
func (t *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if m, ok := data.(map[string]interface{}); ok {
		if ct, ok := m["ContentTemplate"].(string); ok && ct != "" {
			var buf bytes.Buffer
			if err := t.templates.ExecuteTemplate(&buf, ct, m); err != nil {
				return err
			}
			m["ContentHTML"] = template.HTML(buf.String())
		}
	}

	// Execute template to buffer first
	var buf bytes.Buffer
	if err := t.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}

	// Minify HTML only in production
	if t.config.IsProduction {
		return t.minifier.Minify("text/html", w, &buf)
	}

	// In development, write unminified HTML
	_, err := w.Write(buf.Bytes())
	return err
}

// New loads all templates and returns a new Renderer
func New(cfg *config.Config) *Renderer {
	return Load(cfg)
}

// Load loads all templates from the web/templates directory
func Load(cfg *config.Config) *Renderer {
	// Initialize minifier
	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	// Load all template files from various directories
	files, err := filepath.Glob("web/templates/*.html")
	if err != nil {
		log.Fatalf("failed to glob web/templates/*.html: %v", err)
	}

	// Load templates from subdirectories
	adminFiles, _ := filepath.Glob("web/templates/admin/*.html")
	errorFiles, _ := filepath.Glob("web/templates/errors/*.html")
	appFiles, _ := filepath.Glob("web/templates/app/*.html")
	compFiles, _ := filepath.Glob("web/templates/components/*.html")

	files = append(files, adminFiles...)
	files = append(files, errorFiles...)
	files = append(files, appFiles...)
	files = append(files, compFiles...)

	// Check if we found any template files
	if len(files) == 0 {
		log.Fatalf("no template files found. Current directory might be wrong. Looking for: web/templates/*.html")
	}

	log.Printf("Found %d template files to load", len(files))

	funcs := template.FuncMap{
		"eq":        func(a, b string) bool { return a == b },
		"or":        func(a, b bool) bool { return a || b },
		"hasprefix": func(s, prefix string) bool { return strings.HasPrefix(s, prefix) },
	}
	tmpl := template.New("").Funcs(funcs)
	tmpls, err := tmpl.ParseFiles(files...)
	if err != nil {
		log.Fatalf("failed to parse templates %v: %v", files, err)
	}

	log.Printf("Successfully loaded templates")
	return &Renderer{
		templates: tmpls,
		minifier:  m,
		config:    cfg,
	}
}
