package utils

import (
	"strings"
	"testing"
)

func TestGenerateSessionUUID(t *testing.T) {
	uuid, err := GenerateSessionUUID()
	if err != nil {
		t.Fatalf("GenerateSessionUUID failed: %v", err)
	}

	if len(uuid) != 44 {
		t.Errorf("Expected session UUID length 44, got %d", len(uuid))
	}

	// Test uniqueness
	uuid2, err := GenerateSessionUUID()
	if err != nil {
		t.Fatalf("GenerateSessionUUID failed: %v", err)
	}

	if uuid == uuid2 {
		t.Error("Generated UUIDs should be unique")
	}
}

func TestGenerateInvitationCode(t *testing.T) {
	code, err := GenerateInvitationCode()
	if err != nil {
		t.Fatalf("GenerateInvitationCode failed: %v", err)
	}

	if len(code) != 12 {
		t.Errorf("Expected invitation code length 12, got %d", len(code))
	}

	// Test uniqueness
	code2, err := GenerateInvitationCode()
	if err != nil {
		t.Fatalf("GenerateInvitationCode failed: %v", err)
	}

	if code == code2 {
		t.Error("Generated invitation codes should be unique")
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "normal token",
			token:    "abcdefghij1234567890",
			expected: "abcdef...7890",
		},
		{
			name:     "short token",
			token:    "short",
			expected: "***",
		},
		{
			name:     "exact 10 chars",
			token:    "1234567890",
			expected: "***",
		},
		{
			name:     "11 chars",
			token:    "12345678901",
			expected: "123456...8901",
		},
		{
			name:     "long token",
			token:    "this-is-a-very-long-secure-token-with-many-characters",
			expected: "this-i...ters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskToken(tt.token)
			if result != tt.expected {
				t.Errorf("MaskToken(%q) = %q, want %q", tt.token, result, tt.expected)
			}
		})
	}
}

func TestMaskTokenNoLeakage(t *testing.T) {
	token := "secret-token-12345678901234567890"
	masked := MaskToken(token)

	// Ensure masked version doesn't contain the full token
	if strings.Contains(masked, "secret-token-12345678901234567890") {
		t.Error("Masked token contains full token")
	}

	// Ensure it's significantly shorter
	if len(masked) >= len(token) {
		t.Error("Masked token should be shorter than original")
	}
}
