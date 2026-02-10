package utils

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean filename",
			input: "image.jpg",
			want:  "image.jpg",
		},
		{
			name:  "filename with newline",
			input: "file\nname.jpg",
			want:  "file_name.jpg",
		},
		{
			name:  "filename with carriage return",
			input: "file\rname.jpg",
			want:  "file_name.jpg",
		},
		{
			name:  "filename with double quote",
			input: "file\"name.jpg",
			want:  "file_name.jpg",
		},
		{
			name:  "filename with backslash",
			input: "file\\name.jpg",
			want:  "file_name.jpg",
		},
		{
			name:  "filename with multiple dangerous chars",
			input: "file\n\r\"\\name.jpg",
			want:  "file____name.jpg",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "safe special characters preserved",
			input: "file-name_2023.jpg",
			want:  "file-name_2023.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilenameHeaderInjection(t *testing.T) {
	// Test that header injection attempts are neutralized
	malicious := "image.jpg\r\nContent-Type: text/html"
	result := SanitizeFilename(malicious)

	// Check that dangerous characters are replaced
	if result == malicious {
		t.Error("SanitizeFilename should modify strings with CRLF")
	}

	// Ensure no CRLF sequences remain
	for _, char := range result {
		if char == '\n' || char == '\r' || char == '"' || char == '\\' {
			t.Errorf("SanitizeFilename should replace dangerous character, but found %q in %q", char, result)
		}
	}
}
