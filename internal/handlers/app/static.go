package app

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"id-100/internal/templates"
	"id-100/internal/utils"
)

// RulesHandler displays the rules page
func RulesHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           "Leitfaden - 🏠🆔💯",
		"ContentTemplate": "leitfaden.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// ImpressumHandler displays the impressum page
func ImpressumHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           "Impressum - 🏠🆔💯",
		"ContentTemplate": "impressum.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// DatenschutzHandler displays the privacy policy page
func DatenschutzHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           "Datenschutzerklärung - 🏠🆔💯",
		"ContentTemplate": "datenschutz.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}
