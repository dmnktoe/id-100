package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
	"id-100/internal/sentryhelper"
	"id-100/internal/seo"
	"id-100/internal/templates"
	"id-100/internal/utils"
)

// MapHandler renders the interactive map page at /karte.
func MapHandler(c echo.Context) error {
	stats := utils.GetFooterStats()

	cities, err := repository.GetCityContribCounts(context.Background())
	if err != nil {
		log.Printf("MapHandler cities error: %v", err)
		sentryhelper.CaptureException(c, err)
		cities = []repository.CityContrib{}
	}

	baseURL := seo.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
	builder := seo.NewBuilder(baseURL)
	seoMeta := builder.Custom(
		"Interaktive Karte | Innenstadt ID-100",
		"Entdecke alle teilnehmenden Städte auf der interaktiven Karte.",
		"",
		baseURL+"/karte",
		"website",
	)

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           seoMeta.Title,
		"SEO":             seoMeta,
		"ContentTemplate": "map.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
		"Cities":          cities,
	}))
}

// MapDataHandler returns city contribution counts as JSON for the client-side map.
func MapDataHandler(c echo.Context) error {
	cities, err := repository.GetCityContribCounts(context.Background())
	if err != nil {
		sentryhelper.CaptureException(c, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if cities == nil {
		cities = []repository.CityContrib{}
	}
	return c.JSON(http.StatusOK, cities)
}
