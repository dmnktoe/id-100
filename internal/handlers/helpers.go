package handlers

import "os"

// GetGlobalTemplateData returns global data that should be available in all templates
func GetGlobalTemplateData() map[string]interface{} {
	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "http://localhost:8081" // Default fallback
	}

	return map[string]interface{}{
		"NominatimURL": nominatimURL,
	}
}

// MergeTemplateData merges global template data with page-specific data
func MergeTemplateData(data map[string]interface{}) map[string]interface{} {
	global := GetGlobalTemplateData()
	for k, v := range data {
		global[k] = v
	}
	return global
}
