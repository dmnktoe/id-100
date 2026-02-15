package utils

import "testing"

func TestRemoveEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove house, ID, and 100 emojis",
			input:    "Innenstadt (🏠) ID (🆔) - 100 (💯)",
			expected: "Innenstadt ID - 100",
		},
		{
			name:     "Remove single emoji from title",
			input:    "Leitfaden - 🏠🆔💯",
			expected: "Leitfaden -",
		},
		{
			name:     "Text without emojis",
			input:    "This is plain text",
			expected: "This is plain text",
		},
		{
			name:     "Multiple emojis in sequence",
			input:    "Test 🏠🆔💯 text",
			expected: "Test text",
		},
		{
			name:     "Emoji at start",
			input:    "🆔 Test",
			expected: "Test",
		},
		{
			name:     "Emoji at end",
			input:    "Test 🆔",
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveEmojis(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveEmojis(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveDeriveReferences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Replace single derive",
			input:    "This is a derive",
			expected: "This is a ID",
		},
		{
			name:     "Replace derives (plural)",
			input:    "These are derives",
			expected: "These are ID",
		},
		{
			name:     "Replace derived",
			input:    "This is derived from",
			expected: "This is ID from",
		},
		{
			name:     "Case insensitive replacement",
			input:    "Derive and DERIVE",
			expected: "ID and ID",
		},
		{
			name:     "No derive references",
			input:    "This is a test",
			expected: "This is a test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDeriveReferences(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveDeriveReferences(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanSEOText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean title with emojis and derive",
			input:    "Innenstadt (🏠) ID (🆔) - 100 (💯)",
			expected: "Innenstadt ID - 100",
		},
		{
			name:     "Clean text with both emojis and derive references",
			input:    "This derive 🏠 is cool",
			expected: "This ID is cool",
		},
		{
			name:     "Clean already clean text",
			input:    "Clean text",
			expected: "Clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanSEOText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanSEOText(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewSEOMetadata(t *testing.T) {
	title := "Test (🏠) derive"
	description := "This is a derive 🆔 test"
	imageURL := "https://example.com/image.png"
	url := "https://example.com/page"
	pageType := "article"

	meta := NewSEOMetadata(title, description, imageURL, url, pageType)

	expectedTitle := "Test ID"
	expectedDescription := "This is a ID test"

	if meta.Title != expectedTitle {
		t.Errorf("NewSEOMetadata Title = %q; want %q", meta.Title, expectedTitle)
	}
	if meta.Description != expectedDescription {
		t.Errorf("NewSEOMetadata Description = %q; want %q", meta.Description, expectedDescription)
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
}

func TestGetPageSEOMetadata(t *testing.T) {
	baseURL := "https://example.com"

	tests := []struct {
		pageName     string
		expectedURL  string
		expectedType string
	}{
		{"leitfaden", baseURL + "/leitfaden", "website"},
		{"impressum", baseURL + "/impressum", "website"},
		{"datenschutz", baseURL + "/datenschutz", "website"},
		{"upload", baseURL + "/upload", "website"},
		{"request_bag", baseURL + "/request-bag", "website"},
		{"unknown", baseURL, "website"}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.pageName, func(t *testing.T) {
			meta := GetPageSEOMetadata(tt.pageName, baseURL)
			
			if meta.Title == "" {
				t.Errorf("GetPageSEOMetadata(%q) Title is empty", tt.pageName)
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
