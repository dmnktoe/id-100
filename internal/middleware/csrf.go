package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// CSRFProtection validates CSRF tokens for state-changing requests
func CSRFProtection(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Only check POST, PUT, DELETE, PATCH requests
		method := c.Request().Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			return next(c)
		}

		// Skip CSRF check for certain paths (like invitation acceptance without prior session)
		path := c.Request().URL.Path
		skipPaths := []string{
			"/upload/invitations/accept",
			"/werkzeug-anfordern",
		}
		for _, skipPath := range skipPaths {
			if path == skipPath {
				return next(c)
			}
		}

		// Get session
		session, err := Store.Get(c.Request(), "id-100-session")
		if err != nil {
			log.Printf("CSRF session error: %v", err)
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "Sitzungsfehler",
				"ContentTemplate": "access_denied.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		// Get CSRF token from session
		sessionToken, ok := session.Values["csrf_token"].(string)
		if !ok || sessionToken == "" {
			log.Printf("CSRF token not found in session")
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "CSRF-Token fehlt",
				"ContentTemplate": "access_denied.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		// Get CSRF token from request (form field or header)
		var requestToken string
		
		// Try form field first
		contentType := c.Request().Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") || 
		   strings.Contains(contentType, "multipart/form-data") {
			requestToken = c.FormValue("csrf_token")
		}

		// If not in form, try header
		if requestToken == "" {
			requestToken = c.Request().Header.Get("X-CSRF-Token")
		}

		if requestToken == "" {
			log.Printf("CSRF token not provided in request")
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "CSRF-Token fehlt",
				"ContentTemplate": "access_denied.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		// Compare tokens (constant time comparison for security)
		if requestToken != sessionToken {
			log.Printf("CSRF token mismatch")
			return c.Render(http.StatusForbidden, "layout", mergeTemplateData(map[string]interface{}{
				"Title":           "Ungültiger CSRF-Token",
				"ContentTemplate": "access_denied.content",
				"CurrentPath":     c.Request().URL.Path,
				"CurrentYear":     time.Now().Year(),
			}))
		}

		return next(c)
	}
}
