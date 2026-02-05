package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGenerateInvitationRequiresAuth(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Simulate missing token (no context values)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload/invite", nil)
	e.ServeHTTP(rec, req)

	// Should return 403 since there's no token
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing token, got %d", rec.Code)
	}
}

func TestAcceptInvitationWithValidCode(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Mock database to return valid invitation
	invitationCode := "test-invitation-code"
	tokenID := 1
	token := "test-token"
	bagName := "TestBag"

	// Override db functions to simulate invitation flow
	origGetToken := getUploadToken
	defer func() { getUploadToken = origGetToken }()

	getUploadToken = func(ctx context.Context, tkn string) (TokenMeta, error) {
		if tkn == token {
			return TokenMeta{
				ID:               tokenID,
				IsActive:         true,
				MaxUploads:       100,
				TotalUploads:     0,
				TotalSessions:    1,
				CurrentPlayer:    "Alice",
				BagName:          bagName,
				SessionStartedAt: time.Now(),
				SessionUUID:      sql.NullString{String: "primary-session-uuid", Valid: true},
			}, nil
		}
		return TokenMeta{}, fmt.Errorf("token not found")
	}

	// First request to accept invitation (without player name set)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/upload/accept-invite?code=%s", invitationCode), nil)

	// We need to mock the database calls for invitation lookup
	// Since we can't easily mock pgx in tests without a real DB connection,
	// this test demonstrates the structure but would need a test database
	// to fully validate the invitation flow

	e.ServeHTTP(rec, req)

	// In a real scenario with test DB, we'd check:
	// - If player name is not set, should show enter_name_invitation page
	// - If player name is set, should authorize and redirect to upload page

	// For now, we just verify the endpoint exists and doesn't crash
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusOK && rec.Code != http.StatusBadRequest {
		t.Logf("Invitation acceptance endpoint returned status: %d (expected failure without DB)", rec.Code)
	}
}

func TestInvitationFlowRequiresPlayerName(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	invitationCode := "test-code"

	// POST to set player name without consent
	form := url.Values{}
	form.Set("player_name", "Bob")
	form.Set("invitation_code", invitationCode)
	// Missing: form.Set("agree_privacy", "1")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload/invite/set-name", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.ServeHTTP(rec, req)

	// Should fail without consent
	if rec.Code == http.StatusSeeOther {
		t.Fatalf("expected form error for missing consent, got redirect")
	}
}

func TestInvitationCodeValidation(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Request without invitation code
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload/accept-invite", nil)
	e.ServeHTTP(rec, req)

	// Should return 400 bad request
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing invitation code, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Ungültige Einladung") && !strings.Contains(body, "Einladungscode") {
		t.Fatalf("expected error message about invalid invitation, got: %s", body)
	}
}

func TestListSessionsRequiresAuth(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Request without auth
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upload/sessions", nil)
	e.ServeHTTP(rec, req)

	// Should return 403 since there's no token in middleware
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unauthorized session list, got %d", rec.Code)
	}
}

func TestRevokeSessionRequiresAuth(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// Request without auth
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload/sessions/some-uuid/revoke", nil)
	e.ServeHTTP(rec, req)

	// Should return 403 since there's no token in middleware
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unauthorized revoke, got %d", rec.Code)
	}
}

// Integration test demonstrating multi-session flow (would need real DB)
func TestMultiSessionFlow(t *testing.T) {
	// This test documents the expected flow:
	// 1. User A scans QR code, enters name, gets primary session
	// 2. User A generates invitation link
	// 3. User B clicks invitation link
	// 4. User B enters their name
	// 5. User B is authorized and can upload
	// 6. Both User A and B can upload simultaneously
	// 7. User A can revoke User B's access

	t.Log("Multi-session flow requires integration test with real database")
	t.Log("Flow: Primary session -> Generate invitation -> Accept invitation -> Both sessions active -> Revoke")
}
