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
	Environment   string // "production" or "development"
	Port          string
	AdminUsername string
	AdminPassword string
	SentryDSN     string // SentryDSN is the Data Source Name for Sentry error tracking
}

// Load loads configuration from environment variables
func Load() *Config {
	godotenv.Load()

	isProduction := IsProduction()

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
		BaseURL:       GetBaseURL(),
		SessionSecret: sessionSecret,
		IsProduction:  isProduction,
		Environment:   GetEnvironment(),
		Port:          port,
		AdminUsername: os.Getenv("ADMIN_USERNAME"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		SentryDSN:     GetSentryDSN(),
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

// GetSentryDSN returns the Sentry DSN from environment
func GetSentryDSN() string {
	return os.Getenv("SENTRY_DSN")
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}

// GetEnvironment returns the environment string (e.g., "production", "staging", "test", or "development")
func GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}

// GetDatadogAppID returns the Datadog Application ID from environment
func GetDatadogAppID() string {
	return os.Getenv("DATADOG_APP_ID")
}

// GetDatadogClientToken returns the Datadog Client Token from environment
func GetDatadogClientToken() string {
	return os.Getenv("DATADOG_CLIENT_TOKEN")
}
