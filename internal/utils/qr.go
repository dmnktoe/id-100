package utils

import (
	"fmt"
	"html"

	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCodeSVG generates a simple SVG QR code with label
func GenerateQRCodeSVG(url, label string) string {
	// Generate QR code
	qr, _ := qrcode.New(url, qrcode.High)

	// Get bitmap
	bitmap := qr.Bitmap()
	size := len(bitmap)
	scale := 10 // pixels per module
	padding := 40
	labelHeight := 60

	svgWidth := size*scale + 2*padding
	svgHeight := size*scale + 2*padding + labelHeight

	// Escape label to prevent XSS
	escapedLabel := html.EscapeString(label)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
	<rect width="%d" height="%d" fill="white"/>
	<text x="%d" y="30" text-anchor="middle" font-family="Arial" font-size="24" font-weight="bold" fill="black">%s</text>
	<text x="%d" y="%d" text-anchor="middle" font-family="Arial" font-size="16" fill="#666">Scanne für Upload</text>
	<g transform="translate(%d, %d)">`,
		svgWidth, svgHeight, svgWidth, svgHeight,
		svgWidth, svgHeight,
		svgWidth/2, escapedLabel,
		svgWidth/2, svgHeight-20,
		padding, padding+40)

	// Draw QR modules
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if bitmap[y][x] {
				svg += fmt.Sprintf(`
		<rect x="%d" y="%d" width="%d" height="%d" fill="black"/>`,
					x*scale, y*scale, scale, scale)
			}
		}
	}

	svg += `
	</g>
</svg>`

	return svg
}
