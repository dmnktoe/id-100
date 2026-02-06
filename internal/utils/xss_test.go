package utils

import (
	"strings"
	"testing"
)

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "with HTML tags",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "with special chars",
			input:    `<>&"'`,
			expected: "&lt;&gt;&amp;&#34;&#39;",
		},
		{
			name:     "mixed content",
			input:    "Hello <b>World</b> & Friends",
			expected: "Hello &lt;b&gt;World&lt;/b&gt; &amp; Friends",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizePlayerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean name",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "with whitespace",
			input:    "  Alice  ",
			expected: "Alice",
		},
		{
			name:     "with script tags",
			input:    "Bob<script>alert('xss')</script>",
			expected: "Bob&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "very long name",
			input:    strings.Repeat("A", 150),
			expected: strings.Repeat("A", 100),
		},
		{
			name:     "with special HTML chars",
			input:    "Name & <Company>",
			expected: "Name &amp; &lt;Company&gt;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePlayerName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizePlayerName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizePlayerNameLength(t *testing.T) {
	// Test with multibyte characters (e.g., emoji)
	input := strings.Repeat("👨‍💻", 60) // Each emoji is multiple bytes
	result := SanitizePlayerName(input)

	// Count runes, not bytes
	runes := []rune(result)
	if len(runes) > 100 {
		t.Errorf("SanitizePlayerName should limit to 100 runes, got %d", len(runes))
	}
}

func TestSanitizePlayerNameXSSPrevention(t *testing.T) {
	xssAttempts := []string{
		"<img src=x onerror=alert('XSS')>",
		"<svg/onload=alert('XSS')>",
		"javascript:alert('XSS')",
		"<iframe src='javascript:alert(\"XSS\")'></iframe>",
		"<body onload=alert('XSS')>",
	}

	for _, attempt := range xssAttempts {
		result := SanitizePlayerName(attempt)
		// Check that dangerous characters are escaped
		if strings.Contains(result, "<") && !strings.Contains(result, "&lt;") {
			t.Errorf("XSS attempt not properly escaped: %q resulted in %q", attempt, result)
		}
		if strings.Contains(result, ">") && !strings.Contains(result, "&gt;") {
			t.Errorf("XSS attempt not properly escaped: %q resulted in %q", attempt, result)
		}
	}
}
