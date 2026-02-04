package middleware

import (
	"testing"
)

func TestInitSessionStore(t *testing.T) {
	tests := []struct {
		name         string
		secret       string
		isProduction bool
	}{
		{
			name:         "development mode",
			secret:       "dev-secret-key",
			isProduction: false,
		},
		{
			name:         "production mode",
			secret:       "prod-secret-key",
			isProduction: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitSessionStore(tt.secret, tt.isProduction)

			if Store == nil {
				t.Fatal("Store should not be nil after initialization")
			}

			if Store.Options == nil {
				t.Fatal("Store.Options should not be nil")
			}

			if Store.Options.Path != "/" {
				t.Errorf("Store.Options.Path = %q, want %q", Store.Options.Path, "/")
			}

			if Store.Options.MaxAge != 86400*30 {
				t.Errorf("Store.Options.MaxAge = %d, want %d", Store.Options.MaxAge, 86400*30)
			}

			if !Store.Options.HttpOnly {
				t.Error("Store.Options.HttpOnly should be true")
			}

			if Store.Options.Secure != tt.isProduction {
				t.Errorf("Store.Options.Secure = %v, want %v", Store.Options.Secure, tt.isProduction)
			}
		})
	}
}

func TestInitSessionStoreSecureFlag(t *testing.T) {
	// Test that Secure flag is correctly set based on environment
	InitSessionStore("test-secret", false)
	if Store.Options.Secure {
		t.Error("In development, Secure should be false")
	}

	InitSessionStore("test-secret", true)
	if !Store.Options.Secure {
		t.Error("In production, Secure should be true")
	}
}
