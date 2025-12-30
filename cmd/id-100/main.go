package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// gather template files from templates and components
	files, err := filepath.Glob("web/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	comps, err := filepath.Glob("web/templates/components/*.html")
	if err != nil {
		log.Fatal(err)
	}
	files = append(files, comps...)
	if len(files) == 0 {
		log.Fatal("no template files found in web/templates")
	}
	templates := template.Must(template.ParseFiles(files...))

	// sanity check that required component templates exist
	if templates.Lookup("footer") == nil {
		log.Fatalf("footer template not found; parsed files: %v", files)
	}
	if templates.Lookup("header") == nil {
		log.Fatalf("header template not found; parsed files: %v", files)
	}

	t := &Template{
		templates: templates,
	}
	e.Renderer = t

	e.Static("/static", "web/static")

	e.GET("/", func(c echo.Context) error {
		assignments := make([]int, 100)
		for i := range assignments {
			assignments[i] = i + 1
		}

		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Title":           "🏠🆔💯 DÉRIVE 100",
			"Assignments":     assignments,
			"ContentTemplate": "index.content",
		})
	})

	e.GET("/spielregeln", func(c echo.Context) error {
		log.Println("/spielregeln handler called")
		return c.Render(http.StatusOK, "spielregeln.html", map[string]interface{}{
			"Title":           "Spielregeln - 🏠🆔💯 DÉRIVE 100",
			"ContentTemplate": "spielregeln.content",
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
