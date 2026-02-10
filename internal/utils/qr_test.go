package utils

import (
	"strings"
	"testing"
)

func TestGenerateQRCodeSVG(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		label string
	}{
		{
			name:  "simple url",
			url:   "https://example.com",
			label: "ID 100",
		},
		{
			name:  "url with parameters",
			url:   "https://example.com/upload?id=123&token=abc",
			label: "Werkzeug #42",
		},
		{
			name:  "local url",
			url:   "http://localhost:3000/upload",
			label: "Test Upload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg := GenerateQRCodeSVG(tt.url, tt.label)

			// Check that it's a valid SVG
			if !strings.HasPrefix(svg, "<?xml") {
				t.Error("Generated SVG should start with XML declaration")
			}

			if !strings.Contains(svg, "<svg") {
				t.Error("Generated SVG should contain <svg> tag")
			}

			if !strings.Contains(svg, "</svg>") {
				t.Error("Generated SVG should contain closing </svg> tag")
			}

			// Check that label is included and escaped
			if !strings.Contains(svg, tt.label) {
				t.Errorf("Generated SVG should contain label %q", tt.label)
			}

			// Check for expected elements
			if !strings.Contains(svg, "<rect") {
				t.Error("Generated SVG should contain rectangles for QR modules")
			}

			if !strings.Contains(svg, "<text") {
				t.Error("Generated SVG should contain text elements for labels")
			}

			if !strings.Contains(svg, "Scanne für Upload") {
				t.Error("Generated SVG should contain instruction text")
			}
		})
	}
}

func TestGenerateQRCodeSVGXSSProtection(t *testing.T) {
	// Test that HTML special characters are properly escaped
	dangerousLabel := `<script>alert("xss")</script>`
	svg := GenerateQRCodeSVG("https://example.com", dangerousLabel)

	// The label should be escaped
	if strings.Contains(svg, "<script>") {
		t.Error("SVG should escape HTML tags in label")
	}

	// Check that escaped version is present
	if !strings.Contains(svg, "&lt;script&gt;") {
		t.Error("SVG should contain escaped version of HTML tags")
	}
}

func TestGenerateQRCodeSVGDimensions(t *testing.T) {
	svg := GenerateQRCodeSVG("https://example.com", "Test")

	// Check that width and height are specified
	if !strings.Contains(svg, "width=") {
		t.Error("SVG should have width attribute")
	}

	if !strings.Contains(svg, "height=") {
		t.Error("SVG should have height attribute")
	}

	if !strings.Contains(svg, "viewBox=") {
		t.Error("SVG should have viewBox attribute")
	}
}
