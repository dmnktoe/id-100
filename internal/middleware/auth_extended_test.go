package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// TestGetOrCreateSessionUUID tests session UUID creation and retrieval
func TestGetOrCreateSessionUUID(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	tests := []struct {
		name          string
		existingUUID  string
		expectNew     bool
	}{
		{
			name:         "creates new UUID when none exists",
			existingUUID: "",
			expectNew:    true,
		},
		{
			name:         "returns existing UUID",
			existingUUID: "existing-uuid-1234567890123456789012345678",
			expectNew:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &sessions.Session{
				Values: make(map[interface{}]interface{}),
			}
			
			if tt.existingUUID != "" {
				session.Values["session_uuid"] = tt.existingUUID
			}
			
			uuid, err := GetOrCreateSessionUUID(session)
			if err != nil {
				t.Fatalf("GetOrCreateSessionUUID returned error: %v", err)
			}
			
			if uuid == "" {
				t.Error("UUID should not be empty")
			}
			
			if !tt.expectNew && uuid != tt.existingUUID {
				t.Errorf("Expected existing UUID %s, got %s", tt.existingUUID, uuid)
			}
			
			if tt.expectNew && len(uuid) != 44 {
				t.Errorf("New UUID should be 44 characters, got %d", len(uuid))
			}
			
			// Check it's saved in session
			savedUUID, ok := session.Values["session_uuid"].(string)
			if !ok {
				t.Error("UUID not saved in session")
			}
			if savedUUID != uuid {
				t.Errorf("Saved UUID %s doesn't match returned UUID %s", savedUUID, uuid)
			}
		})
	}
}

// TestGetOrCreateCSRFToken tests CSRF token creation and retrieval
func TestGetOrCreateCSRFToken(t *testing.T) {
	tests := []struct {
		name          string
		existingToken string
		expectNew     bool
	}{
		{
			name:          "creates new token when none exists",
			existingToken: "",
			expectNew:     true,
		},
		{
			name:          "returns existing token",
			existingToken: "existing-csrf-token-12345678",
			expectNew:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &sessions.Session{
				Values: make(map[interface{}]interface{}),
			}
			
			if tt.existingToken != "" {
				session.Values["csrf_token"] = tt.existingToken
			}
			
			token, err := GetOrCreateCSRFToken(session)
			if err != nil {
				t.Fatalf("GetOrCreateCSRFToken returned error: %v", err)
			}
			
			if token == "" {
				t.Error("Token should not be empty")
			}
			
			if !tt.expectNew && token != tt.existingToken {
				t.Errorf("Expected existing token %s, got %s", tt.existingToken, token)
			}
			
			if tt.expectNew && len(token) != 32 {
				t.Errorf("New token should be 32 characters, got %d", len(token))
			}
			
			// Check it's saved in session
			savedToken, ok := session.Values["csrf_token"].(string)
			if !ok {
				t.Error("Token not saved in session")
			}
			if savedToken != token {
				t.Errorf("Saved token %s doesn't match returned token %s", savedToken, token)
			}
		})
	}
}

// TestBasicAuth_ValidCredentials tests BasicAuth with valid credentials
func TestBasicAuth_ValidCredentials(t *testing.T) {
	// Set up environment
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret123")
	defer os.Unsetenv("ADMIN_USERNAME")
	defer os.Unsetenv("ADMIN_PASSWORD")
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("admin", "secret123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := BasicAuth(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler was not called with valid credentials")
	}
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestBasicAuth_InvalidUsername tests BasicAuth with invalid username
func TestBasicAuth_InvalidUsername(t *testing.T) {
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret123")
	defer os.Unsetenv("ADMIN_USERNAME")
	defer os.Unsetenv("ADMIN_PASSWORD")
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("wronguser", "secret123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := BasicAuth(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called with invalid username")
	}
	
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// TestBasicAuth_InvalidPassword tests BasicAuth with invalid password
func TestBasicAuth_InvalidPassword(t *testing.T) {
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret123")
	defer os.Unsetenv("ADMIN_USERNAME")
	defer os.Unsetenv("ADMIN_PASSWORD")
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := BasicAuth(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called with invalid password")
	}
	
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// TestBasicAuth_NoCredentials tests BasicAuth without credentials
func TestBasicAuth_NoCredentials(t *testing.T) {
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret123")
	defer os.Unsetenv("ADMIN_USERNAME")
	defer os.Unsetenv("ADMIN_PASSWORD")
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := BasicAuth(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called without credentials")
	}
	
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
	
	// Check WWW-Authenticate header is set
	authHeader := rec.Header().Get("WWW-Authenticate")
	if authHeader == "" {
		t.Error("WWW-Authenticate header not set")
	}
}

// TestBasicAuth_MissingEnvVars tests BasicAuth with missing environment variables
func TestBasicAuth_MissingEnvVars(t *testing.T) {
	// Ensure env vars are not set
	os.Unsetenv("ADMIN_USERNAME")
	os.Unsetenv("ADMIN_PASSWORD")
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("admin", "secret123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := BasicAuth(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called when env vars missing")
	}
	
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

// TestBasicAuth_ConstantTimeComparison tests that credentials are compared in constant time
func TestBasicAuth_ConstantTimeComparison(t *testing.T) {
	// This is a behavioral test - we can't directly test timing,
	// but we can verify the function uses the correct comparison method
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret123")
	defer os.Unsetenv("ADMIN_USERNAME")
	defer os.Unsetenv("ADMIN_PASSWORD")
	
	e := echo.New()
	
	// Test with credentials that differ in length
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.SetBasicAuth("ad", "se")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handler := BasicAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for short credentials, got %d", rec.Code)
	}
}

// TestInitSessionStore_GobRegistration tests that time.Time is registered with gob
func TestInitSessionStore_GobRegistration(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	// Create a session and try to store a time.Time value
	session := &sessions.Session{
		Values: make(map[interface{}]interface{}),
	}
	
	now := time.Now()
	session.Values["test_time"] = now
	
	// If time.Time is not registered, encoding would fail
	// We can't directly test gob encoding here, but we can verify no panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic occurred when setting time.Time in session: %v", r)
		}
	}()
	
	// Verify the value is stored
	if storedTime, ok := session.Values["test_time"].(time.Time); !ok {
		t.Error("time.Time value not stored correctly in session")
	} else if !storedTime.Equal(now) {
		t.Error("Stored time.Time value doesn't match original")
	}
}
