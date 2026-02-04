package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateSecureToken generates a cryptographically secure token
func GenerateSecureToken(length int) (string, error) {
	// Calculate bytes needed to get desired length after base64 encoding
	// base64 encoding produces 4 chars for every 3 bytes
	bytesNeeded := (length*3 + 3) / 4
	b := make([]byte, bytesNeeded)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	if len(encoded) > length {
		return encoded[:length], nil
	}
	return encoded, nil
}
