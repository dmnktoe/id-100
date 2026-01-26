package imgutil

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestDecodeAutoOriented_PNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	got, err := DecodeAutoOriented(&buf)
	if err != nil {
		t.Fatalf("DecodeAutoOriented failed: %v", err)
	}
	if got.Bounds().Dx() != 80 || got.Bounds().Dy() != 40 {
		t.Fatalf("unexpected bounds: %v", got.Bounds())
	}
}

func TestDecodeAutoOriented_JPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpg: %v", err)
	}
	got, err := DecodeAutoOriented(&buf)
	if err != nil {
		t.Fatalf("DecodeAutoOriented failed: %v", err)
	}
	if got.Bounds().Dx() != 50 || got.Bounds().Dy() != 100 {
		t.Fatalf("unexpected bounds: %v", got.Bounds())
	}
}
