package main

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"id-100/internal/config"
	"id-100/internal/database"
	"id-100/internal/handlers"
	appMiddleware "id-100/internal/middleware"
	appSentry "id-100/internal/sentry"
	"id-100/internal/templates"
	"id-100/internal/version"
)

func main() {
	// Load configuration (validates and caches config values)
	cfg := config.Load()

	// Initialize Sentry
	if err := appSentry.InitWithOptions(appSentry.InitOptions{
		DSN:              cfg.SentryDSN,
		Environment:      cfg.Environment,
		Release:          version.Version,
		TracesSampleRate: 0.1,
		EnableLogs:       true, // enable structured logs for backend
	}); err != nil {
		log.Printf("Failed to initialize Sentry error tracking, continuing without it. Please verify SENTRY_DSN configuration: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Initialize database
	database.Init()
	defer database.Close()

	// Initialize session store
	appMiddleware.InitSessionStore(cfg.SessionSecret, cfg.IsProduction)

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Add Sentry middleware if configured
	if cfg.SentryDSN != "" {
		e.Use(echo.WrapMiddleware(sentryhttp.New(sentryhttp.Options{
			Repanic: true,
		}).Handle))
	}

	// Load and set up templates
	t := templates.New(cfg)
	e.Renderer = t

	// Register routes
	handlers.RegisterRoutes(e, cfg.BaseURL)

	// Log app version
	log.Printf("Starting server on port %s (version=%s)", cfg.Port, version.Version)
	log.Fatal(e.Start(":" + cfg.Port))
}
