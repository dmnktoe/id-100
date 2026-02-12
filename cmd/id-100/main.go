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
	"id-100/internal/version"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	database.Init()
	defer database.Close()

	// Initialize session store
	appMiddleware.InitSessionStore(cfg.SessionSecret, cfg.IsProduction)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Load and set up templates
	t := templates.New()
	e.Renderer = t

	// Register routes
	handlers.RegisterRoutes(e, cfg.BaseURL)

	// Log app version
	log.Printf("Starting server on port %s (version=%s)", cfg.Port, version.Version)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
