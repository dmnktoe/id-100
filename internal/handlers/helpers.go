package handlers

import "id-100/internal/templates"

// MergeTemplateData is a convenience wrapper for handlers to use shared template data
func MergeTemplateData(data map[string]interface{}) map[string]interface{} {
	return templates.MergeTemplateData(data)
}
