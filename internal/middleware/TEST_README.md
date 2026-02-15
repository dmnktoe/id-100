# Middleware Test Suite

Comprehensive testing suite for authentication, authorization, and session management middleware.

## Test Coverage

### 1. Auth Middleware (`auth_extended_test.go`)

Tests for session and authentication management:

- **TestGetOrCreateSessionUUID**: Session UUID creation and retrieval
  - Creates new UUID when none exists (44-character unique identifier)
  - Returns existing UUID from session
  
- **TestGetOrCreateCSRFToken**: CSRF token generation and management
  - Creates new token when none exists (32-character token)
  - Returns existing token from session
  
- **TestBasicAuth_ValidCredentials**: Admin authentication with valid credentials
- **TestBasicAuth_InvalidUsername**: Rejection of invalid usernames
- **TestBasicAuth_InvalidPassword**: Rejection of invalid passwords
- **TestBasicAuth_NoCredentials**: Handling of missing credentials
- **TestBasicAuth_MissingEnvVars**: Server error when env vars not configured
- **TestBasicAuth_ConstantTimeComparison**: Security test for timing attack prevention
- **TestInitSessionStore_GobRegistration**: Verification that time.Time is registered for serialization

### 2. CSRF Middleware (`csrf_test.go`)

Tests for Cross-Site Request Forgery protection:

- **TestCSRFProtection_GET**: GET requests bypass CSRF check
- **TestCSRFProtection_POST_ValidToken**: POST with valid CSRF token succeeds
- **TestCSRFProtection_POST_MissingToken**: POST without token is rejected
- **TestCSRFProtection_POST_InvalidToken**: POST with wrong token is rejected
- **TestCSRFProtection_HeaderToken**: CSRF token accepted from X-CSRF-Token header
- **TestCSRFProtection_SkipPaths**: Certain paths skip CSRF validation
  - `/upload/invitations/accept`
  - `/werkzeug-anfordern`
- **TestCSRFProtection_PUT**: PUT requests require CSRF token
- **TestCSRFProtection_DELETE**: DELETE requests require CSRF token
- **TestCSRFProtection_MultipartForm**: CSRF token validation in multipart forms

### 3. Token Middleware (`token_test.go`)

Tests for upload token validation and session binding:

- **TestTokenWithSession_NoToken**: Forbidden response when no token provided
- **TestTokenWithSession_TokenFromQuery**: Token validation from query parameter
- **TestTokenWithSession_TokenFromForm**: Token validation from POST form
- **TestTokenWithSession_InvalidToken**: Rejection of invalid tokens
- **TestTokenWithSession_InactiveToken**: Rejection of deactivated tokens
- **TestTokenWithSession_SessionUUIDCreation**: Automatic session UUID generation
- **TestTokenWithSession_SessionConflict**: Detection of concurrent session conflicts
- **TestTokenWithSession_SessionPersistence**: Session data persists across requests
- **TestTokenWithSession_UploadLimitReached**: Enforcement of upload limits

### 4. Session Helpers (`session_helpers_test.go`)

Tests for session data type conversions:

- **TestGetSessionNumber**: Converts various types to session numbers
  - int, int64, float64, string
  - Edge cases: NaN, Inf, overflow
  
- **TestGetSessionTime**: Converts various types to time.Time
  - time.Time, RFC3339 strings, Unix timestamps
  - Edge cases: invalid strings, out-of-range values

### 5. Session Store (`auth_test.go`)

Tests for session store initialization:

- **TestInitSessionStore**: Verifies cookie store configuration
  - Development vs production mode
  - Path, MaxAge, HttpOnly, Secure flags
  - SameSite=Strict for CSRF protection

## Running Tests

### Run all tests (requires database):
```bash
go test ./internal/middleware -v
```

### Run without database tests (CI-friendly):
```bash
go test ./internal/middleware -v -short
```

### Run with coverage:
```bash
go test ./internal/middleware -cover
```

### Run specific test:
```bash
go test ./internal/middleware -v -run TestCSRFProtection_POST_ValidToken
```

## Test Results

**Total Tests: 32**
- ✅ 24 tests pass in short mode (without database)
- ✅ 8 database tests skip in short mode
- ✅ All tests pass with database available

```
PASS
ok  	id-100/internal/middleware	0.008s
```

## Test Infrastructure

### Mock Renderer (`test_helpers.go`)
Provides a simple Echo renderer for testing HTTP responses without requiring full template rendering.

### Database Test Helpers (`token_test.go`)
- `setupTestDB()`: Initializes test database connection
- `createTestToken()`: Creates test tokens with specific configurations
- Tests automatically skip when run with `-short` flag

## Key Features Tested

### Security Features
- ✅ Constant-time comparison for credentials
- ✅ CSRF token validation
- ✅ Session UUID binding
- ✅ BasicAuth authentication
- ✅ Concurrent session conflict detection

### Session Management
- ✅ Session creation and persistence
- ✅ Session UUID generation (44 characters)
- ✅ CSRF token generation (32 characters)
- ✅ Gob serialization for time.Time
- ✅ Cookie security flags (HttpOnly, Secure, SameSite)

### Token Validation
- ✅ Token from query params
- ✅ Token from POST forms
- ✅ Token from session storage
- ✅ Invalid/missing token handling
- ✅ Inactive token rejection
- ✅ Upload limit enforcement

### Request Flow
- ✅ First-time user flow
- ✅ Returning user flow
- ✅ Session conflict detection
- ✅ Invitation system bypass
- ✅ Multi-session support

## Test Patterns

### Table-Driven Tests
```go
tests := []struct {
    name     string
    input    interface{}
    expected bool
}{
    {"creates new UUID", "", true},
    {"returns existing", "uuid123", false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test implementation
    })
}
```

### Skip Pattern for Database Tests
```go
func TestTokenValidation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database test in short mode")
    }
    // database test implementation
}
```

### Mock Pattern for HTTP Responses
```go
e := echo.New()
e.Renderer = &mockRenderer{}
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)
```

## Continuous Integration

The test suite is designed to work in CI environments:

1. **Short mode** (`-short` flag) skips database tests
2. **No external dependencies** for core tests
3. **Fast execution** (< 10ms for most tests)
4. **Clear error messages** for debugging
5. **Isolated tests** (no shared state)

## Future Enhancements

Potential areas for additional testing:

- [ ] Integration tests with real database
- [ ] Load testing for concurrent sessions
- [ ] Fuzzing for token validation
- [ ] End-to-end flow tests
- [ ] Performance benchmarks
- [ ] Race condition detection

## Contributing

When adding new middleware functionality:

1. Add corresponding tests to appropriate file
2. Follow existing test patterns
3. Use table-driven tests for multiple scenarios
4. Add `-short` skip for database tests
5. Update this README with new test descriptions
6. Ensure all tests pass before committing

## Related Documentation

- [Middleware Implementation](./README.md)
- [Session Management](../docs/sessions.md)
- [CSRF Protection](../docs/csrf.md)
- [Token Authentication](../docs/tokens.md)
