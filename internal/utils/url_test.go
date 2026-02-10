package utils

import (
	"os"
	"testing"
)

func TestEnsureFullImageURL(t *testing.T) {
	// Set MinIO configuration
	os.Setenv("S3_PUBLIC_URL", "http://localhost:9000")
	os.Setenv("S3_BUCKET", "id100-images")

	tests := []struct{ raw, want string }{
		{"", ""},
		{"https://example.com/image.png", "https://example.com/image.png"},
		{"http://localhost:9000/id100-images/derive_5_1.webp", "http://localhost:9000/id100-images/derive_5_1.webp"},
		{"data:image/png;base64,ABC123", "data:image/png;base64,ABC123"},
		{"/id100-images/derive_5_1.webp", "http://localhost:9000/id100-images/derive_5_1.webp"},
		{"id100-images/derive_5_1.webp", "http://localhost:9000/id100-images/derive_5_1.webp"},
		{"derive_5_1.webp", "http://localhost:9000/id100-images/derive_5_1.webp"},
		{"derive_72_1770641182.webp", "http://localhost:9000/id100-images/derive_72_1770641182.webp"},
	}

	for _, tc := range tests {
		got := EnsureFullImageURL(tc.raw)
		if got != tc.want {
			t.Fatalf("raw=%q want=%q got=%q", tc.raw, tc.want, got)
		}
	}
}

func TestEnsureFullImageURLFallback(t *testing.T) {
	// Test fallback when S3_PUBLIC_URL is not set
	os.Unsetenv("S3_PUBLIC_URL")
	os.Setenv("S3_ENDPOINT", "http://minio:9000")
	os.Setenv("S3_BUCKET", "id100-images")

	got := EnsureFullImageURL("derive_1.webp")
	want := "http://minio:9000/id100-images/derive_1.webp"
	
	if got != want {
		t.Fatalf("want=%q got=%q", want, got)
	}
	
	// Test default fallback when neither is set
	os.Unsetenv("S3_PUBLIC_URL")
	os.Unsetenv("S3_ENDPOINT")
	os.Setenv("S3_BUCKET", "id100-images")
	
	got = EnsureFullImageURL("derive_2.webp")
	want = "http://localhost:9000/id100-images/derive_2.webp"
	
	if got != want {
		t.Fatalf("want=%q got=%q", want, got)
	}
}
