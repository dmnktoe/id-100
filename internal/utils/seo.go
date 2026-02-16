package utils

import "id-100/internal/seo"

// SEOMetadata is an alias for backward compatibility
type SEOMetadata = seo.Metadata

// GetBaseURLFromRequest is kept for backward compatibility
func GetBaseURLFromRequest(scheme, host, forwardedHost string) string {
	return seo.GetBaseURLFromRequest(scheme, host, forwardedHost)
}

// NewSEOMetadata is kept for backward compatibility
func NewSEOMetadata(title, description, imageURL, url, pageType string) *SEOMetadata {
	return &SEOMetadata{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		URL:         url,
		Type:        pageType,
	}
}

// GetDefaultSEOMetadata is kept for backward compatibility
func GetDefaultSEOMetadata(baseURL string) *SEOMetadata {
	builder := seo.NewBuilder(baseURL)
	return builder.Default()
}

// GetPageSEOMetadata is kept for backward compatibility
func GetPageSEOMetadata(pageName, baseURL string) *SEOMetadata {
	builder := seo.NewBuilder(baseURL)
	return builder.ForPage(pageName)
}
