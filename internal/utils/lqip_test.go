package utils

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestGenerateLQIP(t *testing.T) {
	t.Run("nil image", func(t *testing.T) {
		lqip, err := GenerateLQIP(nil, 50)
		if err != nil {
			t.Errorf("GenerateLQIP(nil) returned error: %v", err)
		}
		if lqip != "" {
			t.Errorf("GenerateLQIP(nil) = %q, want empty string", lqip)
		}
	})

	t.Run("small image - no resize", func(t *testing.T) {
		// Create a small test image (30x20)
		img := image.NewRGBA(image.Rect(0, 0, 30, 20))
		// Fill with a simple pattern
		for y := 0; y < 20; y++ {
			for x := 0; x < 30; x++ {
				img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			}
		}

		lqip, err := GenerateLQIP(img, 50)
		if err != nil {
			t.Fatalf("GenerateLQIP failed: %v", err)
		}

		// Check it's a data URI
		if !strings.HasPrefix(lqip, "data:image/") {
			t.Errorf("LQIP should be a data URI, got: %s", lqip[:50])
		}

		// Check it contains base64 data
		if !strings.Contains(lqip, "base64,") {
			t.Error("LQIP should contain base64 data")
		}

		// Check for webp or jpeg format
		if !strings.HasPrefix(lqip, "data:image/webp") && !strings.HasPrefix(lqip, "data:image/jpeg") {
			t.Errorf("LQIP should be webp or jpeg format, got: %s", lqip[:30])
		}
	})

	t.Run("large image - resize", func(t *testing.T) {
		// Create a larger test image (200x100)
		img := image.NewRGBA(image.Rect(0, 0, 200, 100))
		// Fill with a simple pattern
		for y := 0; y < 100; y++ {
			for x := 0; x < 200; x++ {
				img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
			}
		}

		maxWidth := 50
		lqip, err := GenerateLQIP(img, maxWidth)
		if err != nil {
			t.Fatalf("GenerateLQIP failed: %v", err)
		}

		// Check it's a data URI
		if !strings.HasPrefix(lqip, "data:image/") {
			t.Errorf("LQIP should be a data URI")
		}

		// The result should be smaller than encoding the full image
		// (this is a weak check but validates the function runs)
		if len(lqip) == 0 {
			t.Error("LQIP should not be empty")
		}

		// Check for webp or jpeg format
		if !strings.HasPrefix(lqip, "data:image/webp") && !strings.HasPrefix(lqip, "data:image/jpeg") {
			t.Errorf("LQIP should be webp or jpeg format")
		}
	})

	t.Run("square image", func(t *testing.T) {
		// Create a square image (100x100)
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			}
		}

		lqip, err := GenerateLQIP(img, 40)
		if err != nil {
			t.Fatalf("GenerateLQIP failed: %v", err)
		}

		if !strings.HasPrefix(lqip, "data:image/") {
			t.Error("LQIP should be a data URI")
		}
	})
}

func TestEncodeWebPDataURI(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	dataURI, err := encodeWebPDataURI(img, 30)
	if err != nil {
		t.Fatalf("encodeWebPDataURI failed: %v", err)
	}

	// Should start with data:image/
	if !strings.HasPrefix(dataURI, "data:image/") {
		t.Errorf("Expected data URI prefix, got: %s", dataURI[:30])
	}

	// Should contain base64
	if !strings.Contains(dataURI, "base64,") {
		t.Error("Expected base64 encoding in data URI")
	}

	// Should be either webp or jpeg
	if !strings.HasPrefix(dataURI, "data:image/webp") && !strings.HasPrefix(dataURI, "data:image/jpeg") {
		t.Errorf("Expected webp or jpeg format, got: %s", dataURI[:30])
	}
}
