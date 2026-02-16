package app

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"id-100/internal/repository"
	"id-100/internal/seo"
	"id-100/internal/templates"
	"id-100/internal/utils"
)

// RequestBagHandler displays the bag request form
func RequestBagHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	
	// Generate SEO metadata
	baseURL := seo.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
	builder := seo.NewBuilder(baseURL)
	seoMeta := builder.ForPage("request_bag")
	
	if c.QueryParam("partial") == "1" {
		return c.Render(http.StatusOK, "request_bag.content", map[string]interface{}{
			"CurrentPath": c.Request().URL.Path,
			"CurrentYear": time.Now().Year(),
			"FooterStats": stats,
			"IsPartial":   true,
		})
	}
	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           seoMeta.Title,
		"SEO":             seoMeta,
		"ContentTemplate": "request_bag.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// RequestBagPostHandler handles bag request submissions
func RequestBagPostHandler(c echo.Context) error {
	type payload struct {
		Email string `json:"email"`
	}
	var p payload
	if strings.Contains(c.Request().Header.Get("Content-Type"), "application/json") {
		if err := c.Bind(&p); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ungültiger Request"})
		}
	} else {
		p.Email = c.FormValue("email")
	}
	email := strings.TrimSpace(p.Email)
	if email == "" || !strings.Contains(email, "@") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ungültige E-Mail"})
	}

	err := repository.InsertBagRequest(context.Background(), email)
	if err != nil {
		log.Printf("Failed to insert bag request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Serverfehler"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
