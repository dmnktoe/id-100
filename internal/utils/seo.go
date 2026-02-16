package utils

// SEOMetadata holds all SEO-related metadata for a page
type SEOMetadata struct {
	Title       string
	Description string
	ImageURL    string
	URL         string
	Type        string // "website" or "article"
}

// GetBaseURLFromRequest extracts the base URL from an HTTP request
// It checks for X-Forwarded-Host header (when behind a proxy) and falls back to the request host
func GetBaseURLFromRequest(scheme, host, forwardedHost string) string {
	if forwardedHost != "" {
		// When behind a proxy, use the forwarded host with https
		return "https://" + forwardedHost
	}
	return scheme + "://" + host
}

// NewSEOMetadata creates a new SEOMetadata instance
func NewSEOMetadata(title, description, imageURL, url, pageType string) *SEOMetadata {
	return &SEOMetadata{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		URL:         url,
		Type:        pageType,
	}
}

// GetDefaultSEOMetadata returns default SEO metadata for the site
func GetDefaultSEOMetadata(baseURL string) *SEOMetadata {
	return &SEOMetadata{
		Title:       "Innenstadt (🏠) ID (🆔) - 100 (💯)",
		Description: "Eine urbane Stadtrallye zur Dokumentation und Wahrnehmung des Stadtraums. Entdecke 100 IDs und teile deine Perspektive auf die Innenstadt.",
		ImageURL:    baseURL + "/static/assets/images/og-image.png",
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
			Title:       "Leitfaden - 🏠🆔💯",
			Description: "Anleitung und Leitfaden zur urbanen Stadtrallye. Erfahre, wie du an der Dokumentation des Stadtraums teilnehmen kannst.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/leitfaden",
			Type:        "website",
		}
	case "impressum":
		return &SEOMetadata{
			Title:       "Impressum - 🏠🆔💯",
			Description: "Impressum und rechtliche Informationen zum Projekt Innenstadt ID - 100.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/impressum",
			Type:        "website",
		}
	case "datenschutz":
		return &SEOMetadata{
			Title:       "Datenschutzerklärung - 🏠🆔💯",
			Description: "Datenschutzerklärung für das Projekt Innenstadt ID - 100. Informationen zum Umgang mit personenbezogenen Daten.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/datenschutz",
			Type:        "website",
		}
	case "upload":
		return &SEOMetadata{
			Title:       "Beweis hochladen - 🏠🆔💯",
			Description: "Lade deine Fotos zur urbanen Stadtrallye hoch und dokumentiere deine Wahrnehmung des Stadtraums.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/upload",
			Type:        "website",
		}
	case "request_bag":
		return &SEOMetadata{
			Title:       "Werkzeug anfordern - 🏠🆔💯",
			Description: "Fordere dein Werkzeug für die urbane Stadtrallye an und starte deine Entdeckungsreise.",
			ImageURL:    defaultMeta.ImageURL,
			URL:         baseURL + "/request-bag",
			Type:        "website",
		}
	default:
		return defaultMeta
	}
}
