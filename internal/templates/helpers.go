package templates

import "id-100/internal/config"

// GetGlobalTemplateData returns global data that should be available in all templates
func GetGlobalTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"NominatimURL": config.GetNominatimURL(),
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
