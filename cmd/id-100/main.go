package main

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Add Sentry middleware if configured
	if cfg.SentryDSN != "" {
		e.Use(sentryecho.New(sentryecho.Options{
			Repanic: true,
		}))
	}

	// Load and set up templates
	t := templates.New(cfg)
	e.Renderer = t

	// Register routes
	handlers.RegisterRoutes(e, cfg.BaseURL)

	// Log app version
	log.Printf("Starting server on port %s (version=%s)", cfg.Port, version.Version)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
