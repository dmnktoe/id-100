package main

import (
	"fmt"
	"os"
	"strings"
)

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
