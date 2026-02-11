package config

import "os"

const (
	// DefaultGeocodingURL is the default Meilisearch API URL if not configured
	DefaultGeocodingURL = "http://localhost:8081"
)

// GetGeocodingURL returns the Meilisearch/geocoding API URL from environment or default
func GetGeocodingURL() string {
	geocodingURL := os.Getenv("GEOCODING_API_URL")
	if geocodingURL == "" {
		geocodingURL = DefaultGeocodingURL
	}
	return geocodingURL
}

// GetMeiliMasterKey returns the Meilisearch master key (for backend operations only)
func GetMeiliMasterKey() string {
	return os.Getenv("MEILI_MASTER_KEY")
}

// GetMeiliSearchKey returns the Meilisearch search key (for frontend/read-only access)
func GetMeiliSearchKey() string {
	return os.Getenv("MEILI_SEARCH_KEY")
}
