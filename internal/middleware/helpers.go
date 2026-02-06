package middleware

import "os"

// getNominatimURL returns the Nominatim URL from environment or default
func getNominatimURL() string {
	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "http://localhost:8081" // Default fallback
	}
	return nominatimURL
}
