package utils

import "testing"

func TestNewSEOMetadata(t *testing.T) {
	title := "Test Title 🏠"
	description := "Test Description 🆔"
	imageURL := "https://example.com/image.png"
	url := "https://example.com/page"
	pageType := "article"

	meta := NewSEOMetadata(title, description, imageURL, url, pageType)

	if meta.Title != title {
		t.Errorf("NewSEOMetadata Title = %q; want %q", meta.Title, title)
	}
	if meta.Description != description {
		t.Errorf("NewSEOMetadata Description = %q; want %q", meta.Description, description)
	}
	if meta.ImageURL != imageURL {
		t.Errorf("NewSEOMetadata ImageURL = %q; want %q", meta.ImageURL, imageURL)
	}
	if meta.URL != url {
		t.Errorf("NewSEOMetadata URL = %q; want %q", meta.URL, url)
	}
	if meta.Type != pageType {
		t.Errorf("NewSEOMetadata Type = %q; want %q", meta.Type, pageType)
	}
}

func TestGetDefaultSEOMetadata(t *testing.T) {
	baseURL := "https://example.com"
	meta := GetDefaultSEOMetadata(baseURL)

	if meta.Title == "" {
		t.Error("GetDefaultSEOMetadata Title is empty")
	}
	if meta.Description == "" {
		t.Error("GetDefaultSEOMetadata Description is empty")
	}
	if meta.URL != baseURL {
		t.Errorf("GetDefaultSEOMetadata URL = %q; want %q", meta.URL, baseURL)
	}
	if meta.Type != "website" {
		t.Errorf("GetDefaultSEOMetadata Type = %q; want 'website'", meta.Type)
	}
	// Verify emojis are preserved in title
	if meta.Title != "Innenstadt (🏠) ID (🆔) - 100 (💯)" {
		t.Errorf("GetDefaultSEOMetadata Title = %q; want 'Innenstadt (🏠) ID (🆔) - 100 (💯)'", meta.Title)
	}
}

func TestGetPageSEOMetadata(t *testing.T) {
	baseURL := "https://example.com"

	tests := []struct {
		pageName      string
		expectedURL   string
		expectedType  string
		expectedTitle string
	}{
		{"leitfaden", baseURL + "/leitfaden", "website", "Leitfaden - 🏠🆔💯"},
		{"impressum", baseURL + "/impressum", "website", "Impressum - 🏠🆔💯"},
		{"datenschutz", baseURL + "/datenschutz", "website", "Datenschutzerklärung - 🏠🆔💯"},
		{"upload", baseURL + "/upload", "website", "Beweis hochladen - 🏠🆔💯"},
		{"request_bag", baseURL + "/werkzeug-anfordern", "website", "Werkzeug anfordern - 🏠🆔💯"},
		{"unknown", baseURL, "website", "Innenstadt (🏠) ID (🆔) - 100 (💯)"}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.pageName, func(t *testing.T) {
			meta := GetPageSEOMetadata(tt.pageName, baseURL)
			
			if meta.Title != tt.expectedTitle {
				t.Errorf("GetPageSEOMetadata(%q) Title = %q; want %q", tt.pageName, meta.Title, tt.expectedTitle)
			}
			if meta.Description == "" {
				t.Errorf("GetPageSEOMetadata(%q) Description is empty", tt.pageName)
			}
			if meta.URL != tt.expectedURL {
				t.Errorf("GetPageSEOMetadata(%q) URL = %q; want %q", tt.pageName, meta.URL, tt.expectedURL)
			}
			if meta.Type != tt.expectedType {
				t.Errorf("GetPageSEOMetadata(%q) Type = %q; want %q", tt.pageName, meta.Type, tt.expectedType)
			}
		})
	}
}

func TestGetBaseURLFromRequest(t *testing.T) {
	tests := []struct {
		name          string
		scheme        string
		host          string
		forwardedHost string
		expected      string
	}{
		{
			name:          "Direct request without proxy",
			scheme:        "http",
			host:          "localhost:8080",
			forwardedHost: "",
			expected:      "http://localhost:8080",
		},
		{
			name:          "HTTPS request without proxy",
			scheme:        "https",
			host:          "example.com",
			forwardedHost: "",
			expected:      "https://example.com",
		},
		{
			name:          "Request behind proxy with X-Forwarded-Host",
			scheme:        "http",
			host:          "internal-server:8080",
			forwardedHost: "example.com",
			expected:      "https://example.com",
		},
		{
			name:          "Request behind proxy with subdomain",
			scheme:        "http",
			host:          "backend",
			forwardedHost: "subdomain.example.com",
			expected:      "https://subdomain.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBaseURLFromRequest(tt.scheme, tt.host, tt.forwardedHost)
			if result != tt.expected {
				t.Errorf("GetBaseURLFromRequest(%q, %q, %q) = %q; want %q", 
					tt.scheme, tt.host, tt.forwardedHost, result, tt.expected)
			}
		})
	}
}
