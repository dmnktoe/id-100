package config

import "os"

const (
	// DefaultNominatimURL is the default Nominatim API URL if not configured
	DefaultNominatimURL = "http://localhost:8081"
)

// GetNominatimURL returns the Nominatim URL from environment or default
func GetNominatimURL() string {
	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = DefaultNominatimURL
	}
	return nominatimURL
}
