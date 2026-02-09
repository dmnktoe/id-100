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
			name:    "MinIO full URL with bucket and file",
			url:     "http://localhost:9000/id100-images/derive_5_1.webp",
			want:    "derive_5_1.webp",
			wantErr: false,
		},
		{
			name:    "MinIO URL with nested path",
			url:     "http://localhost:9000/id100-images/subfolder/image.jpg",
			want:    "subfolder/image.jpg",
			wantErr: false,
		},
		{
			name:    "just filename (no URL)",
			url:     "derive_5_1.webp",
			want:    "derive_5_1.webp",
			wantErr: false,
		},
		{
			name:    "nested path without domain",
			url:     "subfolder/image.jpg",
			want:    "subfolder/image.jpg",
			wantErr: false,
		},
		{
			name:    "legacy Supabase full URL (backward compat)",
			url:     "https://example.supabase.co/storage/v1/object/public/contributions/derive_5_1.webp",
			want:    "derive_5_1.webp",
			wantErr: false,
		},
		{
			name:    "legacy Supabase relative URL (backward compat)",
			url:     "/storage/v1/object/public/contributions/derive_5_1.webp",
			want:    "derive_5_1.webp",
			wantErr: false,
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
