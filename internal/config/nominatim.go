package config

import "os"

const (
	// DefaultPhotonURL is the default Photon API URL if not configured
	DefaultPhotonURL = "http://localhost:8081"
)

// GetNominatimURL returns the Photon/Nominatim URL from environment or default
// Note: Variable name kept as GetNominatimURL for backwards compatibility
func GetNominatimURL() string {
	photonURL := os.Getenv("NOMINATIM_URL")
	if photonURL == "" {
		photonURL = DefaultPhotonURL
	}
	return photonURL
}
