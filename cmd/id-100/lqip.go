package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	xdraw "golang.org/x/image/draw"
)

// generateLQIP creates a small low-quality WebP thumbnail and returns a data URI.
func generateLQIP(src image.Image, maxWidth int) (string, error) {
	if src == nil {
		return "", nil
	}
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	if w <= maxWidth {
		// encode original at low quality
		return encodeWebPDataURI(src, 30)
	}
	// calculate new size
	ratio := float64(maxWidth) / float64(w)
	nw := maxWidth
	nh := int(float64(h) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	// use high quality scaler
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return encodeWebPDataURI(dst, 30)
}

func encodeWebPDataURI(img image.Image, quality int) (string, error) {
	var buf bytes.Buffer
	options, _ := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(quality))
	if err := webp.Encode(&buf, img, options); err != nil {
		// fallback to jpeg
		buf.Reset()
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 30}); err != nil {
			return "", err
		}
		b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
		return fmt.Sprintf("data:image/jpeg;base64,%s", b64), nil
	}
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/webp;base64,%s", b64), nil
}
