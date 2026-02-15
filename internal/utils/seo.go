package utils

import (
	"regexp"
	"strings"
)

// SEOMetadata holds all SEO-related metadata for a page
type SEOMetadata struct {
	Title       string
	Description string
	ImageURL    string
	URL         string
	Type        string // "website" or "article"
}

var (
	// Regex to match emoji characters
	emojiRegex = regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA00}-\x{1FA6F}]|[\x{1FA70}-\x{1FAFF}]|[\x{1F004}-\x{1F0CF}]|[\x{1F170}-\x{1F251}]`)
	
	// Regex patterns to match derive-related terminology
	derivePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bderive[sn]?\b`),
		regexp.MustCompile(`(?i)\bderived?\b`),
	}
)

// RemoveEmojis removes all emoji characters from the input string
func RemoveEmojis(s string) string {
	// First pass: remove emoji using regex
	cleaned := emojiRegex.ReplaceAllString(s, "")
	
	// Second pass: clean up extra whitespace and parentheses artifacts
	cleaned = regexp.MustCompile(`\(\s*\)`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// RemoveDeriveReferences removes all references to "derive" terminology
func RemoveDeriveReferences(s string) string {
	result := s
	for _, pattern := range derivePatterns {
		result = pattern.ReplaceAllString(result, "ID")
	}
	return result
}

// CleanSEOText cleans text for SEO by removing emojis and derive references
func CleanSEOText(s string) string {
	cleaned := RemoveEmojis(s)
	cleaned = RemoveDeriveReferences(cleaned)
	return cleaned
}

// NewSEOMetadata creates a new SEOMetadata instance with cleaned text
func NewSEOMetadata(title, description, imageURL, url, pageType string) *SEOMetadata {
	return &SEOMetadata{
		Title:       CleanSEOText(title),
		Description: CleanSEOText(description),
		ImageURL:    imageURL,
		URL:         url,
		Type:        pageType,
	}
}

// GetDefaultSEOMetadata returns default SEO metadata for the site
func GetDefaultSEOMetadata(baseURL string) *SEOMetadata {
	return &SEOMetadata{
		Title:       "Innenstadt ID - 100",
		Description: "Eine urbane Stadtrallye zur Dokumentation und Wahrnehmung des Stadtraums. Entdecke 100 IDs und teile deine Perspektive auf die Innenstadt.",
		ImageURL:    baseURL + "/static/assets/images/og-image.png", // We'll need to add this
		URL:         baseURL,
		Type:        "website",
	}
}

// GetPageSEOMetadata generates SEO metadata for specific pages
func GetPageSEOMetadata(pageName, baseURL string) *SEOMetadata {
	defaultMeta := GetDefaultSEOMetadata(baseURL)
	
	switch pageName {
	case "leitfaden":
		return &SEOMetadata{
			Title:       "Leitfaden - Innenstadt ID - 100",
			Description: "Anleitung und Leitfaden zur urbanen Stadtrallye. Erfahre, wie du an der Dokumentation des Stadtraums teilnehmen kannst.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/leitfaden",
			Type:        "website",
		}
	case "impressum":
		return &SEOMetadata{
			Title:       "Impressum - Innenstadt ID - 100",
			Description: "Impressum und rechtliche Informationen zum Projekt Innenstadt ID - 100.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/impressum",
			Type:        "website",
		}
	case "datenschutz":
		return &SEOMetadata{
			Title:       "Datenschutzerklärung - Innenstadt ID - 100",
			Description: "Datenschutzerklärung für das Projekt Innenstadt ID - 100. Informationen zum Umgang mit personenbezogenen Daten.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/datenschutz",
			Type:        "website",
		}
	case "upload":
		return &SEOMetadata{
			Title:       "Beweis hochladen - Innenstadt ID - 100",
			Description: "Lade deine Fotos zur urbanen Stadtrallye hoch und dokumentiere deine Wahrnehmung des Stadtraums.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/upload",
			Type:        "website",
		}
	case "request_bag":
		return &SEOMetadata{
			Title:       "Werkzeug anfordern - Innenstadt ID - 100",
			Description: "Fordere dein Werkzeug für die urbane Stadtrallye an und starte deine Entdeckungsreise.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/request-bag",
			Type:        "website",
		}
	default:
		return defaultMeta
	}
}
