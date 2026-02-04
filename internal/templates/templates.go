package templates

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// Renderer is the template renderer for Echo
type Renderer struct {
	templates *template.Template
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

	return t.templates.ExecuteTemplate(w, name, data)
}

// Load loads all templates from the web/templates directory
func Load() *Renderer {
	// Load all template files from various directories
	files, err := filepath.Glob("web/templates/*.html")
	if err != nil {
		log.Fatal(err)
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
	return &Renderer{templates: tmpls}
}
