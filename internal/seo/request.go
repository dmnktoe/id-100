package seo

// GetBaseURLFromRequest extracts the base URL from an HTTP request
// It checks for X-Forwarded-Host header (when behind a proxy) and falls back to the request host
func GetBaseURLFromRequest(scheme, host, forwardedHost string) string {
	if forwardedHost != "" {
		// When behind a proxy, use the forwarded host with https
		return "https://" + forwardedHost
	}
	return scheme + "://" + host
}
