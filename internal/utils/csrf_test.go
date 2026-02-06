package utils

import (
	"testing"
)

func TestGenerateCSRFToken(t *testing.T) {
	token, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated CSRF token is empty")
	}

	// Test uniqueness
	token2, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken failed: %v", err)
	}

	if token == token2 {
		t.Error("Generated CSRF tokens should be unique")
	}
}

func TestCSRFTokenLength(t *testing.T) {
	token, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken failed: %v", err)
	}

	// CSRF token should be at least 32 characters (base64 encoded 32 bytes)
	if len(token) < 32 {
		t.Errorf("Expected CSRF token length >= 32, got %d", len(token))
	}
}
