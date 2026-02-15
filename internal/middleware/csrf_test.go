package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestCSRFProtection_GET tests that GET requests bypass CSRF check
func TestCSRFProtection_GET(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler should be called for GET requests")
	}
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestCSRFProtection_POST_ValidToken tests POST with valid CSRF token
func TestCSRFProtection_POST_ValidToken(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session with CSRF token
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	
	// Set up session with CSRF token
	session, _ := Store.New(req, "id-100-session")
	csrfToken := "test-csrf-token-12345678901234567890"
	session.Values["csrf_token"] = csrfToken
	session.Save(req, rec)
	
	// Get the session cookie
	cookies := rec.Result().Cookies()
	
	// Make POST request with CSRF token
	form := url.Values{}
	form.Add("csrf_token", csrfToken)
	
	req2 := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler should be called with valid CSRF token")
	}
	
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec2.Code)
	}
}

// TestCSRFProtection_POST_MissingToken tests POST without CSRF token
func TestCSRFProtection_POST_MissingToken(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session with CSRF token
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	
	session, _ := Store.New(req, "id-100-session")
	session.Values["csrf_token"] = "test-csrf-token-12345678901234567890"
	session.Save(req, rec)
	
	cookies := rec.Result().Cookies()
	
	// Make POST request WITHOUT CSRF token
	req2 := httptest.NewRequest(http.MethodPost, "/upload", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called without CSRF token")
	}
	
	if rec2.Code != http.StatusForbidden && rec2.Code != http.StatusOK {
		t.Errorf("Expected status 403, got %d", rec2.Code)
	}
}

// TestCSRFProtection_POST_InvalidToken tests POST with invalid CSRF token
func TestCSRFProtection_POST_InvalidToken(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session with CSRF token
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	
	session, _ := Store.New(req, "id-100-session")
	session.Values["csrf_token"] = "correct-token-123456789012345678901"
	session.Save(req, rec)
	
	cookies := rec.Result().Cookies()
	
	// Make POST request with WRONG CSRF token
	form := url.Values{}
	form.Add("csrf_token", "wrong-token-123456789012345678901234")
	
	req2 := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called with invalid CSRF token")
	}
	
	if rec2.Code != http.StatusForbidden && rec2.Code != http.StatusOK {
		t.Errorf("Expected status 403, got %d", rec2.Code)
	}
}

// TestCSRFProtection_HeaderToken tests CSRF token from X-CSRF-Token header
func TestCSRFProtection_HeaderToken(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session with CSRF token
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	
	session, _ := Store.New(req, "id-100-session")
	csrfToken := "test-csrf-token-12345678901234567890"
	session.Values["csrf_token"] = csrfToken
	session.Save(req, rec)
	
	cookies := rec.Result().Cookies()
	
	// Make POST request with CSRF token in header
	req2 := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req2.Header.Set("X-CSRF-Token", csrfToken)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler should be called with valid CSRF token in header")
	}
	
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec2.Code)
	}
}

// TestCSRFProtection_SkipPaths tests that certain paths skip CSRF check
func TestCSRFProtection_SkipPaths(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	skipPaths := []string{
		"/upload/invitations/accept",
		"/werkzeug-anfordern",
	}
	
	for _, path := range skipPaths {
		t.Run(path, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			
			handlerCalled := false
			handler := CSRFProtection(func(c echo.Context) error {
				handlerCalled = true
				return c.String(http.StatusOK, "success")
			})
			
			err := handler(c)
			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}
			
			if !handlerCalled {
				t.Errorf("Inner handler should be called for skip path %s", path)
			}
			
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200 for skip path %s, got %d", path, rec.Code)
			}
		})
	}
}

// TestCSRFProtection_PUT tests that PUT requests require CSRF token
func TestCSRFProtection_PUT(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session without CSRF token
	req := httptest.NewRequest(http.MethodPut, "/upload", nil)
	rec := httptest.NewRecorder()
	
	session, _ := Store.New(req, "id-100-session")
	session.Save(req, rec)
	
	cookies := rec.Result().Cookies()
	
	// Make PUT request without CSRF token
	req2 := httptest.NewRequest(http.MethodPut, "/upload", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called for PUT without CSRF token")
	}
}

// TestCSRFProtection_DELETE tests that DELETE requests require CSRF token
func TestCSRFProtection_DELETE(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session without CSRF token
	req := httptest.NewRequest(http.MethodDelete, "/upload", nil)
	rec := httptest.NewRecorder()
	
	session, _ := Store.New(req, "id-100-session")
	session.Save(req, rec)
	
	cookies := rec.Result().Cookies()
	
	// Make DELETE request without CSRF token
	req2 := httptest.NewRequest(http.MethodDelete, "/upload", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called for DELETE without CSRF token")
	}
}

// TestCSRFProtection_MultipartForm tests CSRF token from multipart form
func TestCSRFProtection_MultipartForm(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	
	// Create session with CSRF token
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	
	session, _ := Store.New(req, "id-100-session")
	csrfToken := "test-csrf-token-12345678901234567890"
	session.Values["csrf_token"] = csrfToken
	session.Save(req, rec)
	
	cookies := rec.Result().Cookies()
	
	// Create multipart form with CSRF token
	body := strings.NewReader("--boundary\r\n" +
		"Content-Disposition: form-data; name=\"csrf_token\"\r\n\r\n" +
		csrfToken + "\r\n" +
		"--boundary--\r\n")
	
	req2 := httptest.NewRequest(http.MethodPost, "/upload", body)
	req2.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)
	
	handlerCalled := false
	handler := CSRFProtection(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler should be called with valid CSRF token in multipart form")
	}
}

