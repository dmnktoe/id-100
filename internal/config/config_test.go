package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		// Empty values are treated as unset by Load.
		t.Setenv("BASE_URL", "")
		t.Setenv("ENVIRONMENT", "")
		t.Setenv("SESSION_SECRET", "")
		t.Setenv("PORT", "")
		t.Setenv("ADMIN_USERNAME", "")
		t.Setenv("ADMIN_PASSWORD", "")

		cfg := Load()

		if cfg.BaseURL != "http://localhost:8080" {
			t.Errorf("Default BaseURL = %q, want %q", cfg.BaseURL, "http://localhost:8080")
		}

		if cfg.IsProduction {
			t.Error("IsProduction should be false by default")
		}

		if cfg.Port != "8080" {
			t.Errorf("Default Port = %q, want %q", cfg.Port, "8080")
		}

		if cfg.SessionSecret == "" {
			t.Error("SessionSecret should have a default value in dev")
		}

		if cfg.AdminUsername != "" {
			t.Error("AdminUsername should be empty when not set")
		}

		if cfg.AdminPassword != "" {
			t.Error("AdminPassword should be empty when not set")
		}
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv("BASE_URL", "https://example.com")
		t.Setenv("ENVIRONMENT", "production")
		t.Setenv("SESSION_SECRET", "my-secret-key")
		t.Setenv("PORT", "3000")
		t.Setenv("ADMIN_USERNAME", "admin")
		t.Setenv("ADMIN_PASSWORD", "pass123")

		cfg := Load()

		if cfg.BaseURL != "https://example.com" {
			t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://example.com")
		}

		if !cfg.IsProduction {
			t.Error("IsProduction should be true when ENVIRONMENT=production")
		}

		if cfg.Port != "3000" {
			t.Errorf("Port = %q, want %q", cfg.Port, "3000")
		}

		if cfg.SessionSecret != "my-secret-key" {
			t.Errorf("SessionSecret = %q, want %q", cfg.SessionSecret, "my-secret-key")
		}

		if cfg.AdminUsername != "admin" {
			t.Errorf("AdminUsername = %q, want %q", cfg.AdminUsername, "admin")
		}

		if cfg.AdminPassword != "pass123" {
			t.Errorf("AdminPassword = %q, want %q", cfg.AdminPassword, "pass123")
		}
	})

	t.Run("production with SESSION_SECRET", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		// A secret must be set in production, otherwise Load calls log.Fatal.
		t.Setenv("SESSION_SECRET", "required-in-prod")

		cfg := Load()
		if !cfg.IsProduction {
			t.Error("IsProduction should be true")
		}
	})
}
