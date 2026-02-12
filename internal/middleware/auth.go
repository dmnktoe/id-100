package middleware

import (
	"crypto/subtle"
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"id-100/internal/utils"
)

// Store is the session store
var Store *sessions.CookieStore

// InitSessionStore initializes the session store with the provided secret
func InitSessionStore(secret string, isProduction bool) {
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   isProduction, // Enable in production with HTTPS
		SameSite: 4,            // Strict mode for CSRF protection (4 = http.SameSiteStrictMode)
	}
}

// GetOrCreateSessionUUID gets or creates a session UUID for this browser
func GetOrCreateSessionUUID(session *sessions.Session) (string, error) {
	if uuid, ok := session.Values["session_uuid"].(string); ok && uuid != "" {
		return uuid, nil
	}

	// Generate new session UUID
	uuid, err := utils.GenerateSessionUUID()
	if err != nil {
		return "", err
	}

	session.Values["session_uuid"] = uuid
	return uuid, nil
}

// GetOrCreateCSRFToken gets or creates a CSRF token for this session
func GetOrCreateCSRFToken(session *sessions.Session) (string, error) {
	if token, ok := session.Values["csrf_token"].(string); ok && token != "" {
		return token, nil
	}

	// Generate new CSRF token
	token, err := utils.GenerateCSRFToken()
	if err != nil {
		return "", err
	}

	session.Values["csrf_token"] = token
	return token, nil
}

// BasicAuth provides basic authentication middleware for admin routes
func BasicAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		adminUser := os.Getenv("ADMIN_USERNAME")
		adminPass := os.Getenv("ADMIN_PASSWORD")

		if adminUser == "" || adminPass == "" {
			log.Printf("ADMIN_USERNAME or ADMIN_PASSWORD not set")
			return c.String(500, "Server misconfiguration")
		}

		username, password, ok := c.Request().BasicAuth()
		if !ok {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			return c.String(401, "Unauthorized")
		}

		// Use constant-time comparison to prevent timing attacks
		userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(adminUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(adminPass)) == 1

		if !userMatch || !passMatch {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			return c.String(401, "Unauthorized")
		}
		return next(c)
	}
}
