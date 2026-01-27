package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// FooterStats holds database statistics for the footer
type FooterStats struct {
	TotalDeriven       int
	TotalContributions int
	ActiveUsers        int
	LastActivity       time.Time
}

// getFooterStats fetches creative database statistics
var getFooterStats = func() FooterStats {
	stats := FooterStats{}

	// Count total deriven
	db.QueryRow(context.Background(), "SELECT COUNT(*) FROM deriven").Scan(&stats.TotalDeriven)

	// Count total contributions
	db.QueryRow(context.Background(), "SELECT COUNT(*) FROM contributions").Scan(&stats.TotalContributions)

	// Count active users (users who contributed)
	db.QueryRow(context.Background(), "SELECT COUNT(DISTINCT user_name) FROM contributions WHERE user_name != ''").Scan(&stats.ActiveUsers)

	// Get last activity timestamp
	var last sql.NullTime
	err := db.QueryRow(context.Background(), "SELECT MAX(created_at) FROM contributions").Scan(&last)
	if err != nil {
		log.Printf("Error fetching last activity: %v", err)
		stats.LastActivity = time.Now()
	} else if last.Valid {
		stats.LastActivity = last.Time
	} else {
		stats.LastActivity = time.Now()
	}

	return stats
}

// ensureFullImageURL makes sure stored image URLs are usable in templates.
// Moved to utils.go to keep main.go small.
func ensureFullImageURL(raw string) string {
	if raw == "" {
		return ""
	}
	// already absolute (including data URIs)
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "data:") {
		return raw
	}
	base := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	bucket := strings.Trim(os.Getenv("S3_BUCKET"), "/")

	// If the path already starts with /storage, just prefix the base
	if strings.HasPrefix(raw, "/storage/") {
		return base + raw
	}
	if strings.HasPrefix(raw, "storage/") {
		return base + "/" + raw
	}

	// If it starts with a slash (other absolute path), prefix base
	if strings.HasPrefix(raw, "/") {
		return base + raw
	}

	// If it already contains the storage object path, be safe
	if strings.Contains(raw, "storage/v1/object/public") {
		if strings.HasPrefix(raw, "/") {
			return base + raw
		}
		return base + "/" + raw
	}

	// If it begins with the bucket name (e.g. "contributions/derive_...")
	if bucket != "" && (strings.HasPrefix(raw, bucket+"/") || strings.HasPrefix(raw, bucket)) {
		return fmt.Sprintf("%s/storage/v1/object/public/%s", base, strings.TrimLeft(raw, "/"))
	}

	// If it's just a filename, assume bucket and build the public url
	if bucket != "" && !strings.Contains(raw, "/") {
		return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", base, bucket, raw)
	}

	// Fallback: prefix base
	return base + "/" + raw
}
