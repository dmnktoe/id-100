package utils

import (
	"testing"
)

func TestExtractFileNameFromURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "full URL with bucket and file",
			url:     "https://example.supabase.co/storage/v1/object/public/contributions/derive_5_1.webp",
			want:    "derive_5_1.webp",
			wantErr: false,
		},
		{
			name:    "relative URL with bucket and file",
			url:     "/storage/v1/object/public/contributions/derive_5_1.webp",
			want:    "derive_5_1.webp",
			wantErr: false,
		},
		{
			name:    "nested path in bucket",
			url:     "https://example.supabase.co/storage/v1/object/public/id100-images/subfolder/image.jpg",
			want:    "subfolder/image.jpg",
			wantErr: false,
		},
		{
			name:    "URL without storage path",
			url:     "https://example.com/image.jpg",
			want:    "",
			wantErr: true,
		},
		{
			name:    "malformed storage URL - no bucket",
			url:     "/storage/v1/object/public/",
			want:    "",
			wantErr: true,
		},
		{
			name:    "malformed storage URL - bucket only",
			url:     "/storage/v1/object/public/bucket",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFileNameFromURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFileNameFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractFileNameFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
