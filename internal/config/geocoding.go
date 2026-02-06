package config

import "os"

const (
	// DefaultGeocodingURL is the default Meilisearch API URL if not configured
	DefaultGeocodingURL = "http://localhost:8081"
)

// GetGeocodingURL returns the Meilisearch/geocoding API URL from environment or default
func GetGeocodingURL() string {
	// Try new variable name first
	geocodingURL := os.Getenv("GEOCODING_API_URL")
	if geocodingURL == "" {
		// Fall back to old variable name for backwards compatibility
		geocodingURL = os.Getenv("NOMINATIM_URL")
	}
	if geocodingURL == "" {
		geocodingURL = DefaultGeocodingURL
	}
	return geocodingURL
}
