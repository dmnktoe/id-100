package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"id-100/internal/models"
	"id-100/internal/repository"
	"id-100/internal/templates"
	"id-100/internal/utils"
)

// DerivenHandler displays the list of deriven with pagination and optional city filter
func DerivenHandler(c echo.Context) error {
	stats := utils.GetFooterStats()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	cityFilter := c.QueryParam("city")
	limit := 20
	offset := (page - 1) * limit

	// Get list of all distinct cities from contributions for the filter dropdown
	cities, err := repository.GetDistinctCities(context.Background())
	if err != nil {
		log.Printf("Cities Query Error: %v", err)
		cities = []string{}
	}

	// Get total count
	totalCount, err := repository.GetDerivenCount(context.Background(), cityFilter)
	if err != nil {
		log.Printf("Count Error: %v", err)
		totalCount = 100 // fallback
	}
	totalPages := (totalCount + limit - 1) / limit

	// Get deriven list
	deriven, err := repository.GetDerivenList(context.Background(), cityFilter, limit, offset)
	if err != nil {
		log.Printf("Query Error: %v", err)
		return c.String(http.StatusInternalServerError, "Datenbankfehler")
	}

	// Normalize image URLs and calculate points tier
	for i := range deriven {
		deriven[i].ImageUrl = utils.EnsureFullImageURL(deriven[i].ImageUrl)
		if deriven[i].Points <= 1 {
			deriven[i].PointsTier = 1
		} else if deriven[i].Points == 2 {
			deriven[i].PointsTier = 2
		} else {
			deriven[i].PointsTier = 3
		}
	}

	// Build pagination pages for template
	var pages []models.PageNumber

	// Always show first page
	pages = append(pages, models.PageNumber{Number: 1, IsCurrent: page == 1})

	// Show dots if current page > 3
	if page > 3 {
		pages = append(pages, models.PageNumber{IsDots: true})
	}

	// Show page before current (if exists and not page 1 or 2)
	if page > 2 {
		pages = append(pages, models.PageNumber{Number: page - 1, IsCurrent: false})
	}

	// Show current page (if not first or last)
	if page > 1 && page < totalPages {
		pages = append(pages, models.PageNumber{Number: page, IsCurrent: true})
	}

	// Show page after current (if exists and not last page or second to last)
	if page < totalPages-1 {
		pages = append(pages, models.PageNumber{Number: page + 1, IsCurrent: false})
	}

	// Show dots if there's a gap to last page
	if page < totalPages-2 {
		pages = append(pages, models.PageNumber{IsDots: true})
	}

	// Always show last page (if more than 1 page)
	if totalPages > 1 {
		pages = append(pages, models.PageNumber{Number: totalPages, IsCurrent: page == totalPages})
	}

	// Generate SEO metadata
	baseURL := utils.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
	seoMeta := utils.GetDefaultSEOMetadata(baseURL)

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           seoMeta.Title,
		"SEO":             seoMeta,
		"Deriven":         deriven,
		"CurrentPage":     page,
		"TotalPages":      totalPages,
		"Pages":           pages,
		"HasNext":         page < totalPages,
		"HasPrev":         page > 1,
		"NextPage":        page + 1,
		"PrevPage":        page - 1,
		"Cities":          cities,
		"SelectedCity":    cityFilter,
		"ContentTemplate": "ids.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}

// DeriveHandler displays a single derive with its contributions
func DeriveHandler(c echo.Context) error {
	stats := utils.GetFooterStats()
	num := c.Param("number")
	pageParam := c.QueryParam("page")
	cityFilter := c.QueryParam("city")

	// Get derive by number
	d, err := repository.GetDeriveByNumber(context.Background(), num)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// Normalize derive image URL
	d.ImageUrl = utils.EnsureFullImageURL(d.ImageUrl)
	// compute PointsTier for styling
	if d.Points <= 1 {
		d.PointsTier = 1
	} else if d.Points == 2 {
		d.PointsTier = 2
	} else {
		d.PointsTier = 3
	}

	// Query contributions
	contribs, err := repository.GetDeriveContributions(context.Background(), d.ID, cityFilter)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Fehler beim Laden der Beiträge")
	}

	// Normalize contribution image URLs
	for i := range contribs {
		contribs[i].ImageUrl = utils.EnsureFullImageURL(contribs[i].ImageUrl)
	}

	// If requested as a partial (AJAX), return only the detail fragment
	if c.QueryParam("partial") == "1" {
		return c.Render(http.StatusOK, "id_detail.content", map[string]interface{}{
			"Derive":        d,
			"Contributions": contribs,
			"PageParam":     pageParam,
			"CityFilter":    cityFilter,
			"IsPartial":     true,
		})
	}

	// Generate SEO metadata for the ID detail page
	baseURL := utils.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
	
	// Create custom SEO metadata for this specific ID
	pageURL := fmt.Sprintf("%s/id/%s", baseURL, num)
	pageTitle := fmt.Sprintf("ID #%d - Innenstadt ID - 100", d.Number)
	pageDescription := d.Description
	if pageDescription == "" {
		pageDescription = fmt.Sprintf("Entdecke ID #%d aus der urbanen Stadtrallye und sieh dir die Beiträge der Teilnehmer*innen an.", d.Number)
	}
	
	seoMeta := utils.NewSEOMetadata(
		pageTitle,
		pageDescription,
		utils.EnsureFullImageURL(d.ImageUrl),
		pageURL,
		"article",
	)

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           seoMeta.Title,
		"SEO":             seoMeta,
		"Derive":          d,
		"Contributions":   contribs,
		"PageParam":       pageParam,
		"CityFilter":      cityFilter,
		"IsPartial":       false,
		"ContentTemplate": "id_detail.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
	}))
}
