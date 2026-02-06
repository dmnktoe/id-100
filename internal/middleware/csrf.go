package middleware

import (
	"id-100/internal/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// CSRFMiddleware provides CSRF protection for state-changing requests
func CSRFMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip CSRF check for GET, HEAD, OPTIONS requests (safe methods)
		if c.Request().Method == "GET" || c.Request().Method == "HEAD" || c.Request().Method == "OPTIONS" {
			return next(c)
		}

		// Get session
		session, err := Store.Get(c.Request(), "id-100-session")
		if err != nil {
			log.Printf("CSRF middleware: session error: %v", err)
			return c.String(http.StatusForbidden, "Session error")
		}

		// Get or generate CSRF token in session
		var csrfToken string
		if token, ok := session.Values["csrf_token"].(string); ok && token != "" {
			csrfToken = token
		} else {
			// Generate new CSRF token
			csrfToken, err = utils.GenerateCSRFToken()
			if err != nil {
				log.Printf("Failed to generate CSRF token: %v", err)
				return c.String(http.StatusInternalServerError, "Server error")
			}
			session.Values["csrf_token"] = csrfToken
			session.Save(c.Request(), c.Response())
		}

		// Make CSRF token available to templates and context
		c.Set("csrf_token", csrfToken)

		// For state-changing requests, validate CSRF token
		if c.Request().Method == "POST" || c.Request().Method == "PUT" || c.Request().Method == "DELETE" || c.Request().Method == "PATCH" {
			// Get token from form, header, or query param
			submittedToken := c.FormValue("csrf_token")
			if submittedToken == "" {
				submittedToken = c.Request().Header.Get("X-CSRF-Token")
			}
			if submittedToken == "" {
				submittedToken = c.QueryParam("csrf_token")
			}

			// Check if JSON request
			if strings.Contains(c.Request().Header.Get("Content-Type"), "application/json") {
				// For JSON requests, read from header or body
				if submittedToken == "" {
					var body map[string]interface{}
					if err := c.Bind(&body); err == nil {
						if token, ok := body["csrf_token"].(string); ok {
							submittedToken = token
						}
					}
					// Reset body for further processing
					c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, int64(1024*1024))
				}
			}

			// Validate token
			if submittedToken != csrfToken {
				log.Printf("CSRF token mismatch: expected %s, got %s", utils.MaskToken(csrfToken), utils.MaskToken(submittedToken))
				return c.String(http.StatusForbidden, "CSRF token validation failed")
			}
		}

		return next(c)
	}
}

// InjectCSRFToken injects CSRF token into all GET requests for templates
func InjectCSRFToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Only for GET requests
		if c.Request().Method != "GET" {
			return next(c)
		}

		// Get session
		session, err := Store.Get(c.Request(), "id-100-session")
		if err != nil {
			log.Printf("CSRF injection: session error: %v", err)
			// Don't fail the request, just skip CSRF token injection
			return next(c)
		}

		// Get or generate CSRF token
		var csrfToken string
		if token, ok := session.Values["csrf_token"].(string); ok && token != "" {
			csrfToken = token
		} else {
			// Generate new CSRF token
			csrfToken, err = utils.GenerateCSRFToken()
			if err != nil {
				log.Printf("Failed to generate CSRF token: %v", err)
				// Don't fail the request
				return next(c)
			}
			session.Values["csrf_token"] = csrfToken
			session.Save(c.Request(), c.Response())
		}

		// Make CSRF token available to context
		c.Set("csrf_token", csrfToken)

		return next(c)
	}
}
