package utils

import (
	"html"
	"strings"
)

// SanitizeHTML escapes HTML to prevent XSS attacks
func SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// SanitizePlayerName sanitizes player names to prevent XSS
func SanitizePlayerName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)
	// Escape HTML
	name = html.EscapeString(name)
	// Limit length (additional safety)
	runes := []rune(name)
	if len(runes) > 100 {
		name = string(runes[:100])
	}
	return name
}
