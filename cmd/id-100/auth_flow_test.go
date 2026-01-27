package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"context"
	"database/sql"
	"html/template"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// helper to setup echo with mocked pgx pool
func setupEchoWithMockDB(t *testing.T) (*echo.Echo, func()) {
	// initialize session store for tests
	store = sessions.NewCookieStore([]byte("test-secret"))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600}

	e := echo.New()
	// Load templates relative to package for tests (repo root is two dirs up)
	files, err := filepath.Glob("../../web/templates/*.html")
	if err != nil {
		t.Fatalf("failed to glob templates: %v", err)
	}
	comps, _ := filepath.Glob("../../web/templates/components/*.html")
	files = append(files, comps...)
	// include single-level subdirs under web/templates (e.g., conflict/*.html)
	subs, _ := filepath.Glob("../../web/templates/*/*.html")
	files = append(files, subs...)
	funcMap := template.FuncMap{
		"eq":        func(a, b string) bool { return a == b },
		"or":        func(a, b bool) bool { return a || b },
		"hasprefix": func(s, prefix string) bool { return strings.HasPrefix(s, prefix) },
	}
	tmpl := template.New("test").Funcs(funcMap)
	tmpl, err = tmpl.ParseFiles(files...)
	if err != nil {
		t.Fatalf("failed to parse templates for tests: %v", err)
	}
	e.Renderer = &Template{templates: tmpl}
	registerRoutes(e)

	origGetFooter := getFooterStats
	cleanup := func() {
		// restore default DB helpers just in case tests override them
		getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
			var m TokenMeta
			var sessStart time.Time
			var sessUUID sql.NullString
			err := db.QueryRow(ctx, `SELECT id, is_active, max_uploads, total_uploads, total_sessions,
			 COALESCE(current_player, ''), COALESCE(bag_name, ''), COALESCE(session_started_at, created_at), COALESCE(session_uuid, '')
			 FROM upload_tokens WHERE token = $1`, token).Scan(&m.ID, &m.IsActive, &m.MaxUploads, &m.TotalUploads, &m.TotalSessions, &m.CurrentPlayer, &m.BagName, &sessStart, &sessUUID)
			if err != nil {
				return m, err
			}
			m.SessionStartedAt = sessStart
			m.SessionUUID = sessUUID
			return m, nil
		}
		updateTokenCurrentPlayer = func(ctx context.Context, token, playerName string, sessionUUID interface{}) error {
			_, err := db.Exec(ctx, "UPDATE upload_tokens SET current_player = $1, session_started_at = NOW(), session_uuid = COALESCE(session_uuid, $3) WHERE token = $2", playerName, token, sessionUUID)
			return err
		}
		getFooterStats = origGetFooter
	}

	// Provide a simple deterministic footer stats for tests to avoid needing a DB
	getFooterStats = func() FooterStats {
		return FooterStats{TotalDeriven: 1, TotalContributions: 0, ActiveUsers: 0, LastActivity: time.Now()}
	}

	return e, cleanup
}

func TestUploadMissingTokenReturnsForbidden(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Zugang verweigert") {
		t.Fatalf("expected access denied page, got body: %s", rec.Body.String())
	}
}

func TestInvalidTokenReturnsInvalidPage(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Override getUploadToken to simulate DB error
	orig := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		return TokenMeta{}, fmt.Errorf("not found")
	}
	defer func() { getUploadToken = orig }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload?token=badtoken", nil)
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Ungültiger Token") {
		t.Fatalf("expected invalid token page, got body: %s", rec.Body.String())
	}
}

func TestDeactivatedTokenIsForbidden(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Simulate a deactivated token
	orig := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		now := time.Now().UTC()
		return TokenMeta{ID: 2, IsActive: false, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "", BagName: "DeactivatedBag", SessionStartedAt: now}, nil
	}
	defer func() { getUploadToken = orig }()

	// GET should be forbidden and show deactivated page
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload?token=deactivated", nil)
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for deactivated token, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Token deaktiviert") {
		t.Fatalf("expected token deactivated page, got body: %s", rec.Body.String())
	}

	// POST to set-name should also be forbidden
	form := url.Values{}
	form.Set("player_name", "Bob")
	form.Set("token", "deactivated")
	form.Set("agree_privacy", "1")

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/upload/set-name", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for set-name on deactivated token, got %d", rec2.Code)
	}
	if !strings.Contains(rec2.Body.String(), "Token deaktiviert") {
		t.Fatalf("expected token deactivated page on set-name, got body: %s", rec2.Body.String())
	}
}

func TestEnterNameFlowAndSetNameRedirectsToUpload(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	now := time.Now().UTC()

	// stub derive list and session uploads to avoid DB access in handler
	origFetch := fetchDerivenList
	fetchDerivenList = func(ctx context.Context) ([]Derive, error) {
		return []Derive{{Number: 1, Title: "T1", ContribCount: 0}}, nil
	}
	defer func() { fetchDerivenList = origFetch }()

	origSess := fetchSessionUploads
	fetchSessionUploads = func(ctx context.Context, tokenID, sessionNumber int) ([]map[string]interface{}, error) {
		return nil, nil
	}
	defer func() { fetchSessionUploads = origSess }()

	// First request: GET /upload?token=goodtoken -> middleware Query returns token row with empty current_player
	orig := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		return TokenMeta{ID: 1, IsActive: true, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "", BagName: "MyBag", SessionStartedAt: now}, nil
	}
	defer func() { getUploadToken = orig }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload?token=goodtoken", nil)
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on name entry, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Willkommen") {
		t.Fatalf("expected name entry page, got body: %s", rec.Body.String())
	}

	// Now POST to /upload/set-name with form data
	// Expect middleware to query token again and return the same minimal row
	orig2 := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		return TokenMeta{ID: 1, IsActive: true, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "", BagName: "MyBag", SessionStartedAt: now}, nil
	}
	defer func() { getUploadToken = orig2 }()

	// Replace the DB update call with a test double that records invocation
	origUpdate := updateTokenCurrentPlayer
	var updated bool
	updateTokenCurrentPlayer = func(ctx context.Context, token, playerName string, sessionUUID interface{}) error {
		if token != "goodtoken" || playerName != "Alice" {
			return fmt.Errorf("unexpected args")
		}
		updated = true
		return nil
	}
	defer func() { updateTokenCurrentPlayer = origUpdate }()
	form := url.Values{}
	form.Set("player_name", "Alice")
	form.Set("token", "goodtoken")
	form.Set("agree_privacy", "1")

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/upload/set-name", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// copy cookies from previous response (session cookie)
	for _, c := range rec.Result().Cookies() {
		req2.AddCookie(c)
	}

	e.ServeHTTP(rec2, req2)

	if !updated {
		t.Fatalf("expected updateTokenCurrentPlayer to be called")
	}

	// Expect redirect to /upload (SeeOther)
	if rec2.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect after set-name, got %d", rec2.Code)
	}

	// Follow-up GET /upload?token=goodtoken should show upload page with player name
	// Expect Query to return a row with current_player set
	orig3 := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		return TokenMeta{ID: 1, IsActive: true, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "Alice", BagName: "MyBag", SessionStartedAt: now}, nil
	}
	defer func() { getUploadToken = orig3 }()
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/upload?token=goodtoken", nil)
	// ensure session cookie from set-name is sent
	for _, c := range rec2.Result().Cookies() {
		req3.AddCookie(c)
	}
	e.ServeHTTP(rec3, req3)

	// debug output
	t.Logf("rec3 status=%d body=%s", rec3.Code, rec3.Body.String())

	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 on upload page after set-name, got %d; body: %s", rec3.Code, rec3.Body.String())
	}
	if !strings.Contains(rec3.Body.String(), "Beweisfoto hochladen") {
		t.Fatalf("expected upload page, got body: %s", rec3.Body.String())
	}
	if !strings.Contains(rec3.Body.String(), "👤 Alice") {
		t.Fatalf("expected player name on upload page, got body: %s", rec3.Body.String())
	}
}

func TestPostUploadConflictWhenSessionUUIDMismatch(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	now := time.Now().UTC()

	// stub derive list and session uploads to avoid DB access in handler
	origFetch := fetchDerivenList
	fetchDerivenList = func(ctx context.Context) ([]Derive, error) {
		return []Derive{{Number: 1, Title: "T1", ContribCount: 0}}, nil
	}
	defer func() { fetchDerivenList = origFetch }()

	origSess := fetchSessionUploads
	fetchSessionUploads = func(ctx context.Context, tokenID, sessionNumber int) ([]map[string]interface{}, error) {
		return nil, nil
	}
	defer func() { fetchSessionUploads = origSess }()

	// Initial GET to establish session UUID (DB has no session_uuid yet)
	orig := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		return TokenMeta{ID: 1, IsActive: true, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "Alice", BagName: "MyBag", SessionStartedAt: now, SessionUUID: sql.NullString{Valid: false}}, nil
	}
	defer func() { getUploadToken = orig }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload?token=goodtoken", nil)
	e.ServeHTTP(rec, req)

	// New test: session UUID generation error is handled
	origGen := generateSecureToken
	generateSecureToken = func(length int) (string, error) {
		return "", fmt.Errorf("boom")
	}
	defer func() { generateSecureToken = origGen }()

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/upload?token=goodtoken", nil)
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when session uuid generation fails, got %d", rec2.Code)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on initial upload page, got %d", rec.Code)
	}

	// Now simulate another request where DB indicates the token is bound to a different session UUID
	orig2 := getUploadToken
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		return TokenMeta{ID: 1, IsActive: true, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "Alice", BagName: "MyBag", SessionStartedAt: now, SessionUUID: sql.NullString{String: "other-session-uuid", Valid: true}}, nil
	}
	defer func() { getUploadToken = orig2 }()
	// Perform POST to /upload (no need to include multipart - middleware will reject before file handling)
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/upload?token=goodtoken", nil)
	// include session cookie from initial GET
	for _, c := range rec.Result().Cookies() {
		req3.AddCookie(c)
	}

	e.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when session UUID mismatch, got %d, body: %s", rec3.Code, rec3.Body.String())
	}
}
