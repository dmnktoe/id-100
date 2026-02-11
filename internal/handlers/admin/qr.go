package admin

import (
	"context"
	"fmt"
	"mime"
	"net/http"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"

	"id-100/internal/repository"
	"id-100/internal/utils"
)

// AdminDownloadQRHandler generates and returns QR code as SVG or PNG
func AdminDownloadQRHandler(c echo.Context, baseURL string) error {
	tokenID := c.Param("id")

	// Get token from database
	token, bagName, err := repository.GetTokenByID(context.Background(), tokenID)
	if err != nil {
		return c.String(http.StatusNotFound, "Token not found")
	}

	// Generate upload URL
	uploadURL := fmt.Sprintf("%s/upload?token=%s", baseURL, token)

	// Check format parameter
	format := c.QueryParam("format")
	if format == "" {
		format = "png" // default
	}

	switch format {
	case "svg":
		// Generate SVG QR code using custom SVG generation
		svg := utils.GenerateQRCodeSVG(uploadURL, bagName)
		c.Response().Header().Set("Content-Type", "image/svg+xml")
		// Use mime.FormatMediaType to safely encode filename
		filename := fmt.Sprintf("qr_%s.svg", utils.SanitizeFilename(bagName))
		c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return c.String(http.StatusOK, svg)

	case "png":
		// Generate PNG QR code
		qr, err := qrcode.New(uploadURL, qrcode.High)
		if err != nil {
			return c.String(http.StatusInternalServerError, "QR generation failed")
		}

		pngBytes, err := qr.PNG(512)
		if err != nil {
			return c.String(http.StatusInternalServerError, "PNG generation failed")
		}

		c.Response().Header().Set("Content-Type", "image/png")
		// Use SanitizeFilename to prevent header injection
		filename := fmt.Sprintf("qr_%s.png", utils.SanitizeFilename(bagName))
		c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return c.Blob(http.StatusOK, "image/png", pngBytes)

	default:
		return c.String(http.StatusBadRequest, "Invalid format. Use 'svg' or 'png'")
	}
}
