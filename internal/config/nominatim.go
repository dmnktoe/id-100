package config

import "os"

const (
	// DefaultMeilisearchURL is the default Meilisearch API URL if not configured
	DefaultMeilisearchURL = "http://localhost:8081"
)

// GetNominatimURL returns the Meilisearch URL from environment or default
// Note: Variable name kept as GetNominatimURL for backwards compatibility
func GetNominatimURL() string {
	meilisearchURL := os.Getenv("NOMINATIM_URL")
	if meilisearchURL == "" {
		meilisearchURL = DefaultMeilisearchURL
	}
	return meilisearchURL
}
