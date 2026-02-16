package seo

import "fmt"

// Metadata holds all SEO-related metadata for a page
type Metadata struct {
	Title       string
	Description string
	ImageURL    string
	URL         string
	Type        string // "website" or "article"
}

// Builder helps construct SEO metadata
type Builder struct {
	config  *Config
	baseURL string
}

// NewBuilder creates a new SEO metadata builder
func NewBuilder(baseURL string) *Builder {
	return &Builder{
		config:  GetConfig(),
		baseURL: baseURL,
	}
}

// ForPage returns metadata for a named static page
func (b *Builder) ForPage(pageName string) *Metadata {
	page, exists := b.config.Pages[pageName]
	if !exists {
		// Return default metadata
		return b.Default()
	}

	return &Metadata{
		Title:       page.Title,
		Description: page.Description,
		ImageURL:    b.baseURL + b.config.DefaultImage,
		URL:         b.baseURL + page.Path,
		Type:        page.Type,
	}
}

// ForID returns metadata for a specific ID page
func (b *Builder) ForID(idNumber int, idTitle, idDescription, idImageURL string) *Metadata {
	title := fmt.Sprintf("ID #%d - Innenstadt ID - 100", idNumber)
	description := idDescription
	if description == "" {
		description = fmt.Sprintf("Entdecke ID #%d aus der urbanen Stadtrallye und sieh dir die Beiträge der Teilnehmer*innen an.", idNumber)
	}

	return &Metadata{
		Title:       title,
		Description: description,
		ImageURL:    idImageURL,
		URL:         fmt.Sprintf("%s/id/%d", b.baseURL, idNumber),
		Type:        "article",
	}
}

// Default returns the default metadata
func (b *Builder) Default() *Metadata {
	return &Metadata{
		Title:       b.config.DefaultTitle,
		Description: b.config.DefaultDescription,
		ImageURL:    b.baseURL + b.config.DefaultImage,
		URL:         b.baseURL,
		Type:        "website",
	}
}

// Custom creates custom metadata
func (b *Builder) Custom(title, description, imageURL, url, pageType string) *Metadata {
	return &Metadata{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		URL:         url,
		Type:        pageType,
	}
}
