package templates

import (
	"id-100/internal/config"
	"id-100/internal/version"
)

// GetGlobalTemplateData returns global data that should be available in all templates
func GetGlobalTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"GeocodingURL":        config.GetGeocodingURL(),
		"MeiliKey":            config.GetMeiliSearchKey(),
		"SentryDSN":           config.GetSentryDSN(),
		"Environment":         config.GetEnvironment(),
		"DatadogAppID":        config.GetDatadogAppID(),
		"DatadogClientToken":  config.GetDatadogClientToken(),
		"AppVersion":          version.Version,
		"BaseURL":             config.GetBaseURL(),
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
