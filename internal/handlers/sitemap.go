package handlers

import (
"context"
"log"
"net/http"

"github.com/labstack/echo/v4"

"id-100/internal/repository"
"id-100/internal/seo"
)

// SitemapHandler generates and serves the sitemap.xml
func SitemapHandler(c echo.Context) error {
// Get base URL from request
baseURL := seo.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))

// Get all ID numbers from the database
deriven, err := repository.GetDerivenList(context.Background(), "", 1000, 0) // Get up to 1000 IDs
if err != nil {
log.Printf("Failed to fetch IDs for sitemap: %v", err)
return c.String(http.StatusInternalServerError, "Failed to generate sitemap")
}

// Extract ID numbers
idNumbers := make([]int, len(deriven))
for i, d := range deriven {
idNumbers[i] = d.Number
}

// Generate sitemap
sitemapXML, err := seo.GenerateSitemap(baseURL, idNumbers)
if err != nil {
log.Printf("Failed to generate sitemap: %v", err)
return c.String(http.StatusInternalServerError, "Failed to generate sitemap")
}

// Set appropriate headers and return XML
return c.XMLBlob(http.StatusOK, sitemapXML)
}
