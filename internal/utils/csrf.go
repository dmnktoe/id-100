package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateCSRFToken generates a secure CSRF token
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
