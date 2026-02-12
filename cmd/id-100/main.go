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
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Sentry
	if err := appSentry.Init(cfg.SentryDSN); err != nil {
		log.Printf("Sentry initialization error: %v", err)
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
	t := templates.New()
	e.Renderer = t

	// Register routes
	handlers.RegisterRoutes(e, cfg.BaseURL)

	log.Printf("Starting server on port %s", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
