package sentryhelper

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
)

// CaptureError captures an error to Sentry with the specified level
func CaptureError(c *echo.Context, err error, level sentry.Level) {
	if c == nil || c.Request() == nil {
		return
	}
	if hub := sentry.GetHubFromContext(c.Request().Context()); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			hub.CaptureException(err)
		})
	}
}

// CaptureException captures an error to Sentry with default error level
func CaptureException(c *echo.Context, err error) {
	if c == nil || c.Request() == nil {
		return
	}
	if hub := sentry.GetHubFromContext(c.Request().Context()); hub != nil {
		hub.CaptureException(err)
	}
}

// Logger returns a sentry.Logger bound to the request's context so structured logs
// can be correlated with traces. If the echo context is nil, a background context
// logger is returned.
func Logger(c *echo.Context) sentry.Logger {
	if c == nil || c.Request() == nil {
		return sentry.NewLogger(context.Background())
	}
	return sentry.NewLogger(c.Request().Context())
}
