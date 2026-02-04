package utils

import (
	"os"
	"testing"
)

func TestEnsureFullImageURL(t *testing.T) {
	os.Setenv("SUPABASE_URL", "https://abjrgsgtwtaqfunrdpig.supabase.co")
	os.Setenv("S3_BUCKET", "contributions")

	tests := []struct{ raw, want string }{
		{"", ""},
		{"https://example.com/image.png", "https://example.com/image.png"},
		{"/storage/v1/object/public/contributions/derive_5_1.webp", "https://abjrgsgtwtaqfunrdpig.supabase.co/storage/v1/object/public/contributions/derive_5_1.webp"},
		{"storage/v1/object/public/contributions/derive_5_1.webp", "https://abjrgsgtwtaqfunrdpig.supabase.co/storage/v1/object/public/contributions/derive_5_1.webp"},
		{"contributions/derive_5_1.webp", "https://abjrgsgtwtaqfunrdpig.supabase.co/storage/v1/object/public/contributions/derive_5_1.webp"},
		{"derive_5_1.webp", "https://abjrgsgtwtaqfunrdpig.supabase.co/storage/v1/object/public/contributions/derive_5_1.webp"},
	}

	for _, tc := range tests {
		got := EnsureFullImageURL(tc.raw)
		if got != tc.want {
			t.Fatalf("raw=%q want=%q got=%q", tc.raw, tc.want, got)
		}
	}
}
