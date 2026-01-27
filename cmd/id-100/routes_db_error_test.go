package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUploadPostDBErrorReturns500(t *testing.T) {
	e, cleanup := setupEchoWithMockDB(t)
	defer cleanup()

	// override getUploadToken to return ok for middleware, then error for handler
	orig := getUploadToken
	call := 0
	getUploadToken = func(ctx context.Context, token string) (TokenMeta, error) {
		call++
		if call == 1 {
			// first call (middleware): return a valid token so middleware proceeds
			now := time.Now().UTC()
			return TokenMeta{ID: 1, IsActive: true, MaxUploads: 10, TotalUploads: 0, TotalSessions: 1, CurrentPlayer: "Alice", BagName: "MyBag", SessionStartedAt: now, SessionUUID: sql.NullString{Valid: false}}, nil
		}
		// second call (handler): simulate DB error
		return TokenMeta{}, fmt.Errorf("db down")
	}
	defer func() { getUploadToken = orig }()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload?token=goodtoken", strings.NewReader(""))
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when token lookup fails in handler, got %d, body: %s", rec.Code, rec.Body.String())
	}
}
