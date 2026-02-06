package middleware

import "id-100/internal/templates"

// mergeTemplateData is a convenience wrapper around templates.MergeTemplateData
func mergeTemplateData(data map[string]interface{}) map[string]interface{} {
	return templates.MergeTemplateData(data)
}
