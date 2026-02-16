package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"id-100/internal/handlers/admin"
	"id-100/internal/handlers/app"
	"id-100/internal/middleware"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(e *echo.Echo, baseURL string) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "id-100",
		})
	})

	e.GET("/api/stats", StatsHandler)

	// Sitemap for SEO
	e.GET("/sitemap.xml", SitemapHandler)

	e.Static("/static", "web/static")

	e.GET("/", app.DerivenHandler)
	e.GET("/id/:number", app.DeriveHandler)

	// Upload routes - protected by token middleware with session support
	e.GET("/upload", app.UploadGetHandler, middleware.TokenWithSession)
	e.POST("/upload", app.UploadPostHandler, middleware.TokenWithSession)
	e.POST("/upload/set-name", app.SetPlayerNameHandler, middleware.TokenWithSession)
	e.POST("/upload/end-session", app.EndSessionHandler, middleware.TokenWithSession)
	e.POST("/upload/contributions/:id/delete", app.UserDeleteContributionHandler, middleware.TokenWithSession)

	e.GET("/leitfaden", app.RulesHandler)
	e.GET("/impressum", app.ImpressumHandler)
	e.GET("/datenschutz", app.DatenschutzHandler)
	e.GET("/werkzeug-anfordern", app.RequestBagHandler)
	e.POST("/werkzeug-anfordern", app.RequestBagPostHandler)

	// Admin routes for token management
	adminGroup := e.Group("/admin", middleware.BasicAuth)
	adminGroup.GET("", admin.AdminDashboardHandler)
	adminGroup.GET("/tokens", admin.AdminTokenListHandler)
	adminGroup.POST("/tokens", func(c echo.Context) error {
		return admin.AdminCreateTokenHandler(c, baseURL)
	})
	adminGroup.POST("/tokens/:id/deactivate", admin.AdminTokenDeactivateHandler)
	adminGroup.POST("/tokens/:id/reset", admin.AdminTokenResetHandler)
	adminGroup.POST("/tokens/:id/assign", admin.AdminTokenAssignHandler)
	adminGroup.POST("/tokens/:id/quota", admin.AdminUpdateQuotaHandler)
	adminGroup.GET("/tokens/:id/qr", func(c echo.Context) error {
		return admin.AdminDownloadQRHandler(c, baseURL)
	})

	// Werkzeug request management
	adminGroup.POST("/werkzeug-anfragen/:id/complete", admin.AdminBagRequestCompleteHandler)

	// Contribution deletion
	adminGroup.POST("/contributions/:id/delete", admin.AdminDeleteContributionHandler)
}
