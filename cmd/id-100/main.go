package main

import (
	"bytes"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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
	ContribCount int    `json:"contrib_count"`
}

// ensureFullImageURL is implemented in cmd/id-100/utils.go to keep main.go smaller.

func main() {
	initDatabase()
	defer db.Close()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// load templates (moved to helper)
	t := LoadTemplates()
	e.Renderer = t

	// register routes in routes.go
	registerRoutes(e)
	// routes are registered in routes.go

	e.Logger.Fatal(e.Start(":8080"))
}
