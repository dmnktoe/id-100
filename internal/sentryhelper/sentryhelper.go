package sentryhelper

import (
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
)

// CaptureError captures an error to Sentry with the specified level
func CaptureError(c echo.Context, err error, level sentry.Level) {
	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			hub.CaptureException(err)
		})
	}
}

// CaptureException captures an error to Sentry with default error level
func CaptureException(c echo.Context, err error) {
	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
	}
}
