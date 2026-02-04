package utils

import (
	"strings"
	"testing"
)

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short token", 10},
		{"medium token", 32},
		{"long token", 64},
		{"very long token", 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateSecureToken(tt.length)
			if err != nil {
				t.Fatalf("GenerateSecureToken(%d) returned error: %v", tt.length, err)
			}
			
			// Check length
			if len(token) != tt.length {
				t.Errorf("GenerateSecureToken(%d) length = %d, want %d", tt.length, len(token), tt.length)
			}
			
			// Check that token is URL-safe base64 (no +, /, or =)
			if strings.ContainsAny(token, "+/=") {
				t.Errorf("Token contains non-URL-safe characters: %s", token)
			}
		})
	}
}

func TestGenerateSecureTokenUniqueness(t *testing.T) {
	length := 32
	iterations := 100
	tokens := make(map[string]bool)
	
	for i := 0; i < iterations; i++ {
		token, err := GenerateSecureToken(length)
		if err != nil {
			t.Fatalf("GenerateSecureToken failed: %v", err)
		}
		
		if tokens[token] {
			t.Errorf("Generated duplicate token: %s", token)
		}
		tokens[token] = true
	}
	
	if len(tokens) != iterations {
		t.Errorf("Expected %d unique tokens, got %d", iterations, len(tokens))
	}
}

func TestGenerateSecureTokenEdgeCases(t *testing.T) {
	// Test with length 1
	token, err := GenerateSecureToken(1)
	if err != nil {
		t.Fatalf("GenerateSecureToken(1) failed: %v", err)
	}
	if len(token) != 1 {
		t.Errorf("GenerateSecureToken(1) length = %d, want 1", len(token))
	}
	
	// Test with length 0 - should still work
	token, err = GenerateSecureToken(0)
	if err != nil {
		t.Fatalf("GenerateSecureToken(0) failed: %v", err)
	}
	if len(token) != 0 {
		t.Errorf("GenerateSecureToken(0) length = %d, want 0", len(token))
	}
}
