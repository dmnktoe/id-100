package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer für Echo
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	// Middleware für schöneres Logging im Terminal
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Templates laden (Wicked Arts Layout)
	// Wir parsen alle .html Dateien im Ordner 'templates'
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = t

	// Statische Dateien (für dein CSS/Bilder)
	e.Static("/static", "public")

	// ROUTE: Gallery (Landing Page)
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Title": "🏠🆔💯 DÉRIVE 100",
		})
	})

	// Server starten
	e.Logger.Fatal(e.Start(":8080"))
}