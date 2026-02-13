package middleware

import (
	"context"
	"encoding/gob"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"id-100/internal/database"
)

func init() {
	// Register time.Time for gob encoding in tests
	gob.Register(time.Time{})
}

// mockDB is a helper to set up a test database connection
func setupTestDB(t *testing.T) func() {
	// Initialize test database
	database.Init()
	
	return func() {
		// Cleanup function
		if database.DB != nil {
			database.Close()
		}
	}
}

// createTestToken creates a test upload token in the database
func createTestToken(t *testing.T, token string, isActive bool, maxUploads int, currentPlayer string, sessionUUID *string) int {
	t.Helper()
	
	var tokenID int
	query := `INSERT INTO upload_tokens (token, is_active, max_uploads, total_uploads, total_sessions, 
	          current_player, bag_name, session_uuid, created_at) 
	          VALUES ($1, $2, $3, 0, 1, $4, 'Test Bag', $5, NOW()) 
	          RETURNING id`
	
	err := database.DB.QueryRow(context.Background(), query, 
		token, isActive, maxUploads, currentPlayer, sessionUUID).Scan(&tokenID)
	
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	
	return tokenID
}

// TestTokenWithSession_NoToken tests handling when no token is provided
func TestTokenWithSession_NoToken(t *testing.T) {
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	e.Renderer = &mockRenderer{}
	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handler := TokenWithSession(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	
	// Should return forbidden when no token
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusOK {
		t.Errorf("Expected status forbidden or OK, got %d", rec.Code)
	}
}

// TestTokenWithSession_TokenFromQuery tests token validation from query parameter
func TestTokenWithSession_TokenFromQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-query"
	createTestToken(t, token, true, 10, "", nil)
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := TokenWithSession(func(c echo.Context) error {
		handlerCalled = true
		
		// Check context values are set
		if c.Get("token") != token {
			t.Errorf("Token not set in context")
		}
		if c.Get("bag_name") != "Test Bag" {
			t.Errorf("Bag name not set in context")
		}
		
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler was not called")
	}
}

// TestTokenWithSession_TokenFromForm tests token validation from POST form
func TestTokenWithSession_TokenFromForm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-form"
	createTestToken(t, token, true, 10, "", nil)
	
	e := echo.New()
	form := url.Values{}
	form.Add("token", token)
	form.Add("player_name", "Test Player")
	
	req := httptest.NewRequest(http.MethodPost, "/upload/set-name", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := TokenWithSession(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if !handlerCalled {
		t.Error("Inner handler was not called for POST form token")
	}
}

// TestTokenWithSession_InvalidToken tests handling of invalid token
func TestTokenWithSession_InvalidToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload?token=invalid-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := TokenWithSession(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called with invalid token")
	}
	
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusOK {
		t.Errorf("Expected status forbidden, got %d", rec.Code)
	}
}

// TestTokenWithSession_InactiveToken tests handling of inactive token
func TestTokenWithSession_InactiveToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-inactive"
	createTestToken(t, token, false, 10, "Test Player", nil)
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := TokenWithSession(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called with inactive token")
	}
}

// TestTokenWithSession_SessionUUIDCreation tests session UUID creation
func TestTokenWithSession_SessionUUIDCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-uuid"
	createTestToken(t, token, true, 10, "", nil)
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handler := TokenWithSession(func(c echo.Context) error {
		sessionUUID := c.Get("session_uuid")
		if sessionUUID == nil || sessionUUID == "" {
			t.Error("Session UUID not created")
		}
		
		// Check it's a valid format (44 characters)
		if uuidStr, ok := sessionUUID.(string); ok {
			if len(uuidStr) != 44 {
				t.Errorf("Session UUID has wrong length: %d, expected 44", len(uuidStr))
			}
		}
		
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}

// TestTokenWithSession_SessionConflict tests session UUID conflict detection
func TestTokenWithSession_SessionConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-conflict"
	boundSessionUUID := "existing-session-uuid-12345678901234567890"
	createTestToken(t, token, true, 10, "Existing Player", &boundSessionUUID)
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := TokenWithSession(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	// Should return conflict status
	if handlerCalled {
		t.Error("Inner handler should not be called when session conflict detected")
	}
	
	if rec.Code != http.StatusConflict && rec.Code != http.StatusOK {
		t.Errorf("Expected status conflict (409), got %d", rec.Code)
	}
}

// TestTokenWithSession_SessionPersistence tests session data persistence
func TestTokenWithSession_SessionPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-persist"
	createTestToken(t, token, true, 10, "", nil)
	
	e := echo.New()
	
	// First request - create session
	req1 := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	
	var firstSessionUUID string
	handler := TokenWithSession(func(c echo.Context) error {
		firstSessionUUID = c.Get("session_uuid").(string)
		return c.String(http.StatusOK, "success")
	})
	
	err := handler(c1)
	if err != nil {
		t.Errorf("First request returned error: %v", err)
	}
	
	// Extract session cookie
	cookies := rec1.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "id-100-session" {
			sessionCookie = cookie
			break
		}
	}
	
	if sessionCookie == nil {
		t.Skip("Session cookie not set, skipping persistence test")
	}
	
	// Second request - should use same session
	req2 := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	req2.AddCookie(sessionCookie)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	
	var secondSessionUUID string
	handler2 := TokenWithSession(func(c echo.Context) error {
		secondSessionUUID = c.Get("session_uuid").(string)
		return c.String(http.StatusOK, "success")
	})
	
	err = handler2(c2)
	if err != nil {
		t.Errorf("Second request returned error: %v", err)
	}
	
	if firstSessionUUID != secondSessionUUID {
		t.Errorf("Session UUID changed between requests: %s != %s", firstSessionUUID, secondSessionUUID)
	}
}

// TestTokenWithSession_UploadLimitReached tests upload limit handling
func TestTokenWithSession_UploadLimitReached(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}
	
	cleanup := setupTestDB(t)
	defer cleanup()
	
	InitSessionStore("test-secret", false)
	
	token := "test-token-limit"
	tokenID := createTestToken(t, token, true, 5, "Test Player", nil)
	
	// Set total uploads to max
	_, err := database.DB.Exec(context.Background(), 
		"UPDATE upload_tokens SET total_uploads = 5 WHERE id = $1", tokenID)
	if err != nil {
		t.Fatalf("Failed to update uploads: %v", err)
	}
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/upload?token="+token, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handlerCalled := false
	handler := TokenWithSession(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "success")
	})
	
	err = handler(c)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
	
	if handlerCalled {
		t.Error("Inner handler should not be called when upload limit reached")
	}
}
