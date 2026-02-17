package templates

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	assetManifest     map[string]string
	assetManifestOnce sync.Once
)

// LoadAssetManifest loads the asset manifest from disk
func LoadAssetManifest() map[string]string {
	assetManifestOnce.Do(func() {
		manifestPath := "web/static/manifest.json"
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			log.Printf("Warning: Could not load asset manifest from %s: %v", manifestPath, err)
			// Fallback to old paths if manifest doesn't exist
			assetManifest = map[string]string{
				"main.css": "/static/style.css",
				"main.js":  "/static/main.js",
			}
			return
		}

		if err := json.Unmarshal(data, &assetManifest); err != nil {
			log.Printf("Warning: Could not parse asset manifest: %v", err)
			assetManifest = map[string]string{
				"main.css": "/static/style.css",
				"main.js":  "/static/main.js",
			}
			return
		}

		log.Printf("Loaded asset manifest with %d entries", len(assetManifest))
	})

	return assetManifest
}

// GetAssetPath returns the versioned path for an asset
func GetAssetPath(assetName string) string {
	manifest := LoadAssetManifest()
	if path, ok := manifest[assetName]; ok {
		return path
	}
	// Fallback to the asset name itself
	return "/static/" + assetName
}
