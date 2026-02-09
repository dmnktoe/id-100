package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"id-100/internal/database"
	"id-100/internal/models"
)

// EnsureFullImageURL makes sure stored image URLs are usable in templates
// Constructs MinIO URLs for images stored in S3-compatible storage
func EnsureFullImageURL(raw string) string {
	if raw == "" {
		return ""
	}
	// already absolute (including data URIs)
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "data:") {
		return raw
	}
	
	// Get MinIO configuration
	s3Endpoint := strings.TrimRight(os.Getenv("S3_ENDPOINT"), "/")
	bucket := os.Getenv("S3_BUCKET")
	
	// Default to minio:9000 if not set (for Docker internal)
	if s3Endpoint == "" {
		s3Endpoint = "http://minio:9000"
	}
	if bucket == "" {
		bucket = "id100-images"
	}

	// If it's just a filename or path, construct MinIO URL
	// MinIO public URLs: http://minio:9000/bucket-name/object-key
	fileName := strings.TrimLeft(raw, "/")
	
	// Remove bucket name if it's already in the path
	if strings.HasPrefix(fileName, bucket+"/") {
		fileName = strings.TrimPrefix(fileName, bucket+"/")
	}
	
	return fmt.Sprintf("%s/%s/%s", s3Endpoint, bucket, fileName)
}

// GetFooterStats wraps the database function and returns a FooterStats model
func GetFooterStats() models.FooterStats {
	stats := models.FooterStats{}
	totalDeriven, totalContribs, activeUsers, lastActivity := database.GetFooterStats()

	stats.TotalDeriven = totalDeriven
	stats.TotalContributions = totalContribs
	stats.ActiveUsers = activeUsers

	if lastActivity.Valid {
		stats.LastActivity = lastActivity.Time
	} else {
		stats.LastActivity = time.Now()
	}

	return stats
}

// SanitizeFilename removes characters that could cause header injection
func SanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, name)
}
