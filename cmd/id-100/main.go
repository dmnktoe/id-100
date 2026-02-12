package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"id-100/internal/config"
	"id-100/internal/database"
	"id-100/internal/handlers"
	appMiddleware "id-100/internal/middleware"
	"id-100/internal/templates"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Datadog tracing
	ddCfg := config.LoadDatadogConfig()
	appMiddleware.InitDatadog(ddCfg)
	defer appMiddleware.StopDatadog()

	// Initialize database
	database.Init()
	defer database.Close()

	// Initialize session store
	appMiddleware.InitSessionStore(cfg.SessionSecret, cfg.IsProduction)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Add Datadog tracing middleware if enabled
	if ddCfg.Enabled {
		e.Use(appMiddleware.DatadogMiddleware(ddCfg.ServiceName))
	}

	// Load and set up templates
	t := templates.New()
	e.Renderer = t

	// Register routes
	handlers.RegisterRoutes(e, cfg.BaseURL)

	log.Printf("Starting server on port %s", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
