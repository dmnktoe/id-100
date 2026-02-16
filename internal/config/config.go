package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	BaseURL       string
	SessionSecret string
	IsProduction  bool
	Port          string
	AdminUsername string
	AdminPassword string
}

// Load loads configuration from environment variables
func Load() *Config {
	godotenv.Load()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	isProduction := os.Getenv("ENVIRONMENT") == "production"

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		if isProduction {
			log.Fatal("SESSION_SECRET must be set in production. Generate one with: openssl rand -base64 32")
		}
		log.Println("WARNING: Using insecure default SESSION_SECRET. Set SESSION_SECRET environment variable.")
		sessionSecret = "id-100-secret-key-change-in-production"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		BaseURL:       baseURL,
		SessionSecret: sessionSecret,
		IsProduction:  isProduction,
		Port:          port,
		AdminUsername: os.Getenv("ADMIN_USERNAME"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}
}

// GetBaseURL returns the base URL from environment or default
func GetBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL
}
