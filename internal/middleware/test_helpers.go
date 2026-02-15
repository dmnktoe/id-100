package middleware

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// mockRenderer is a simple renderer for testing
type mockRenderer struct{}

func (m *mockRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Write forbidden status
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.WriteHeader(http.StatusForbidden)
	}
	return nil
}
