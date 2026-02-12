package admin

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"

	"id-100/internal/models"
	"id-100/internal/repository"
	"id-100/internal/templates"
	"id-100/internal/utils"
)

// AdminDashboardHandler shows the admin dashboard
func AdminDashboardHandler(c echo.Context) error {
	// Get all tokens
	tokens, err := repository.GetAllTokens(context.Background())
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Get recent contributions
	recentContribs, err := repository.GetRecentContributions(context.Background(), 20)
	if err != nil {
		log.Printf("Failed to fetch recent contributions: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelWarning)
				hub.CaptureException(err)
			})
		}
		recentContribs = []models.RecentContrib{}
	}

	// Normalize image URLs
	for i := range recentContribs {
		recentContribs[i].ImageUrl = utils.EnsureFullImageURL(recentContribs[i].ImageUrl)
	}

	// Fetch bag requests (with optional status filter)
	status := c.QueryParam("bag_status")
	tab := c.QueryParam("tab")
	if tab != "tokens" && tab != "requests" && tab != "contribs" {
		tab = "tokens"
	}

	// Get counts for filter badges
	openCount, handledCount, err := repository.GetBagRequestCounts(context.Background())
	if err != nil {
		log.Printf("Failed to fetch bag request counts: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelWarning)
				hub.CaptureException(err)
			})
		}
		openCount, handledCount = 0, 0
	}

	// Get bag requests
	bagRequests, err := repository.GetBagRequests(context.Background(), status, 50)
	if err != nil {
		log.Printf("Failed to fetch bag requests: %v", err)
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelWarning)
				hub.CaptureException(err)
			})
		}
		bagRequests = []models.BagRequest{}
	}

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           "Admin Dashboard",
		"ContentTemplate": "admin_dashboard.content",
		"AdditionalCSS":   "admin.styles.css",
		"Tokens":          tokens,
		"RecentContribs":  recentContribs,
		"BagRequests":     bagRequests,
		"BagStatus":       status,
		"OpenCount":       openCount,
		"HandledCount":    handledCount,
		"Tab":             tab,
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
	}))
}
