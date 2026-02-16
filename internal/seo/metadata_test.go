package seo

import "testing"

func TestBuilder_ForPage(t *testing.T) {
baseURL := "https://example.com"
builder := NewBuilder(baseURL)

tests := []struct {
pageName    string
expectedURL string
}{
{"home", "https://example.com/"},
{"leitfaden", "https://example.com/leitfaden"},
{"impressum", "https://example.com/impressum"},
{"datenschutz", "https://example.com/datenschutz"},
{"upload", "https://example.com/upload"},
{"request_bag", "https://example.com/werkzeug-anfordern"},
}

for _, tt := range tests {
t.Run(tt.pageName, func(t *testing.T) {
meta := builder.ForPage(tt.pageName)

if meta.Title == "" {
t.Errorf("ForPage(%q) Title is empty", tt.pageName)
}
if meta.URL != tt.expectedURL {
t.Errorf("ForPage(%q) URL = %q; want %q", tt.pageName, meta.URL, tt.expectedURL)
}
if meta.Description == "" {
t.Errorf("ForPage(%q) Description is empty", tt.pageName)
}
})
}

// Test unknown page returns default
t.Run("unknown_page", func(t *testing.T) {
meta := builder.ForPage("unknown")
defaultMeta := builder.Default()

if meta.Title != defaultMeta.Title {
t.Error("ForPage(unknown) should return default title")
}
})
}

func TestBuilder_ForID(t *testing.T) {
baseURL := "https://example.com"
builder := NewBuilder(baseURL)

// Test with provided description
meta := builder.ForID(42, "Test Title", "Custom description", "https://example.com/image.jpg")

expectedTitle := "ID #42 - Innenstadt ID - 100"
if meta.Title != expectedTitle {
t.Errorf("ForID Title = %q; want %q", meta.Title, expectedTitle)
}
if meta.Description != "Custom description" {
t.Errorf("ForID Description = %q; want 'Custom description'", meta.Description)
}
if meta.URL != "https://example.com/id/42" {
t.Errorf("ForID URL = %q; want 'https://example.com/id/42'", meta.URL)
}
if meta.Type != "article" {
t.Errorf("ForID Type = %q; want 'article'", meta.Type)
}

// Test with empty description (should use default)
meta2 := builder.ForID(99, "Test", "", "")
if meta2.Description == "" {
t.Error("ForID with empty description should generate default description")
}
}

func TestBuilder_Default(t *testing.T) {
baseURL := "https://example.com"
builder := NewBuilder(baseURL)

meta := builder.Default()

if meta.Title == "" {
t.Error("Default metadata Title is empty")
}
if meta.Description == "" {
t.Error("Default metadata Description is empty")
}
if meta.URL != baseURL {
t.Errorf("Default metadata URL = %q; want %q", meta.URL, baseURL)
}
if meta.Type != "website" {
t.Errorf("Default metadata Type = %q; want 'website'", meta.Type)
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
