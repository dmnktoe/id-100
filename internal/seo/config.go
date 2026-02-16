package seo

// PageConfig holds SEO configuration for a single page
type PageConfig struct {
Path        string
Title       string
Description string
Type        string // "website" or "article"
}

// Config holds all SEO configurations for the application
type Config struct {
DefaultTitle       string
DefaultDescription string
DefaultImage       string
Pages              map[string]PageConfig
}

// GetConfig returns the centralized SEO configuration
func GetConfig() *Config {
return &Config{
DefaultTitle:       "Innenstadt (🏠) ID (🆔) - 100 (💯)",
DefaultDescription: "Eine urbane Stadtrallye zur Dokumentation und Wahrnehmung des Stadtraums. Entdecke 100 IDs und teile deine Perspektive auf die Innenstadt.",
DefaultImage:       "/static/assets/images/og-image.png",
Pages: map[string]PageConfig{
"home": {
Path:        "/",
Title:       "Innenstadt (🏠) ID (🆔) - 100 (💯)",
Description: "Eine urbane Stadtrallye zur Dokumentation und Wahrnehmung des Stadtraums. Entdecke 100 IDs und teile deine Perspektive auf die Innenstadt.",
Type:        "website",
},
"leitfaden": {
Path:        "/leitfaden",
Title:       "Leitfaden - 🏠🆔💯",
Description: "Anleitung und Leitfaden zur urbanen Stadtrallye. Erfahre, wie du an der Dokumentation des Stadtraums teilnehmen kannst.",
Type:        "website",
},
"impressum": {
Path:        "/impressum",
Title:       "Impressum - 🏠🆔💯",
Description: "Impressum und rechtliche Informationen zum Projekt Innenstadt ID - 100.",
Type:        "website",
},
"datenschutz": {
Path:        "/datenschutz",
Title:       "Datenschutzerklärung - 🏠🆔💯",
Description: "Datenschutzerklärung für das Projekt Innenstadt ID - 100. Informationen zum Umgang mit personenbezogenen Daten.",
Type:        "website",
},
"upload": {
Path:        "/upload",
Title:       "Beweis hochladen - 🏠🆔💯",
Description: "Lade deine Fotos zur urbanen Stadtrallye hoch und dokumentiere deine Wahrnehmung des Stadtraums.",
Type:        "website",
},
"request_bag": {
Path:        "/werkzeug-anfordern",
Title:       "Werkzeug anfordern - 🏠🆔💯",
Description: "Fordere dein Werkzeug für die urbane Stadtrallye an und starte deine Entdeckungsreise.",
Type:        "website",
},
},
}
}

// GetStaticPages returns a list of all static page keys for sitemap generation
func (c *Config) GetStaticPages() []string {
return []string{"home", "leitfaden", "impressum", "datenschutz", "upload", "request_bag"}
}
