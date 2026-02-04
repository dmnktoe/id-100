package handlers

import (
	"github.com/labstack/echo/v4"
	"id-100/internal/middleware"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(e *echo.Echo, baseURL string) {
	e.Static("/static", "web/static")

	e.GET("/", DerivenHandler)
	e.GET("/id/:number", DeriveHandler)

	// Upload routes - protected by token middleware with session support
	e.GET("/upload", UploadGetHandler, middleware.TokenWithSession)
	e.POST("/upload", UploadPostHandler, middleware.TokenWithSession)
	e.POST("/upload/set-name", SetPlayerNameHandler, middleware.TokenWithSession)

	e.GET("/leitfaden", RulesHandler)
	e.GET("/impressum", ImpressumHandler)
	e.GET("/datenschutz", DatenschutzHandler)
	e.GET("/werkzeug-anfordern", RequestBagHandler)
	e.POST("/werkzeug-anfordern", RequestBagPostHandler)

	// Admin routes for token management
	adminGroup := e.Group("/admin", middleware.BasicAuth)
	adminGroup.GET("", AdminDashboardHandler)
	adminGroup.GET("/tokens", AdminTokenListHandler)
	adminGroup.POST("/tokens", func(c echo.Context) error {
		return AdminCreateTokenHandler(c, baseURL)
	})
	adminGroup.POST("/tokens/:id/deactivate", AdminTokenDeactivateHandler)
	adminGroup.POST("/tokens/:id/reset", AdminTokenResetHandler)
	adminGroup.POST("/tokens/:id/assign", AdminTokenAssignHandler)
	adminGroup.POST("/tokens/:id/quota", AdminUpdateQuotaHandler)
	adminGroup.GET("/tokens/:id/qr", func(c echo.Context) error {
		return AdminDownloadQRHandler(c, baseURL)
	})

	// Werkzeug request management
	adminGroup.POST("/werkzeug-anfragen/:id/complete", AdminBagRequestCompleteHandler)
}
