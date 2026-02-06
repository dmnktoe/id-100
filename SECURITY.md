# Security Features - Session Management & Access Control

This document describes the security features implemented to prevent unauthorized access and resource hijacking.

## Overview

The application now implements robust session management with the following features:

1. **Session UUID Binding**: Each browser gets a unique session identifier
2. **Conflict Detection**: Prevents concurrent access from multiple devices
3. **Bag Release**: Users can voluntarily release access
4. **Invitation System**: Secure codes for multi-user collaboration
5. **XSS Prevention**: All user input is sanitized
6. **Token Masking**: Sensitive data is never logged in plain text

## Session UUID Binding

### How It Works

1. **Browser Identification**: When a user first visits the site, a 44-character random session UUID is generated and stored in a secure cookie
2. **Token Binding**: When a user enters their name, their session UUID is bound to the upload token in the database
3. **Validation**: On every request, the middleware checks if the session UUID matches the bound session in the database

### Database Schema

```sql
-- Add session_uuid to upload_tokens
ALTER TABLE upload_tokens ADD COLUMN session_uuid TEXT DEFAULT '';

-- Create session_bindings for multi-session support
CREATE TABLE session_bindings (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
    session_uuid TEXT NOT NULL,
    player_name TEXT NOT NULL,
    player_city TEXT DEFAULT '',
    is_owner BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(token_id, session_uuid)
);
```

## Conflict Detection

### HTTP 409 Conflict Response

When a user tries to access a token that's bound to a different session:

- **Status Code**: 409 Conflict
- **Error Page**: User-friendly explanation with solutions
- **Log Entry**: Security event logged (with masked tokens)

### Allowed Scenarios

Access is granted if:
1. Session UUID matches the primary bound session (owner)
2. Session UUID is in the `session_bindings` table (invited user)

### Example Flow

```
User A (Phone):   Scans QR code → Enters name → Session bound
User B (Desktop): Tries same QR → ❌ 409 Conflict → Access denied
```

## Bag Release Feature

### API Endpoint

```http
POST /session/release
```

### What It Does

1. Clears `session_uuid` from `upload_tokens` table
2. Removes entry from `session_bindings` table
3. Clears session cookies
4. Allows other users to claim the bag

### UI

- Button in "Werkzeug-Verwaltung" section on upload page
- Confirmation dialog before releasing
- Redirects to homepage after successful release

## Invitation System

### Generate Invitation Code

```http
POST /session/invitation/generate
```

**Response:**
```json
{
  "status": "success",
  "code": "ABC123XYZ789",
  "expires_at": "2024-02-07T12:00:00Z"
}
```

### Accept Invitation

```http
POST /session/invitation/accept
Content-Type: application/json

{
  "code": "ABC123XYZ789"
}
```

**Response (Success):**
```json
{
  "status": "success",
  "redirect_url": "/upload?token=...",
  "bag_name": "Werkzeug 1"
}
```

### Database Schema

```sql
CREATE TABLE invitation_codes (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    created_by_session_uuid TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_by_session_uuid TEXT DEFAULT '',
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Features

- **12-character codes**: Secure, easy to share
- **24-hour expiration**: Codes automatically expire
- **Single use**: Each code can only be used once
- **Copy to clipboard**: Easy sharing via UI

## Multi-Session Support

### Active Sessions Management

```http
GET /session/active
```

**Response:**
```json
{
  "status": "success",
  "sessions": [
    {
      "id": 1,
      "session_uuid": "abc123...",
      "player_name": "Alice",
      "player_city": "Berlin",
      "is_owner": true,
      "is_current": true,
      "created_at": "2024-02-06T10:00:00Z",
      "last_active_at": "2024-02-06T11:00:00Z"
    }
  ]
}
```

### Revoke Access

```http
POST /session/revoke/:session_id
```

**Owner Only**: Only the bag owner (primary session) can revoke access for invited users.

## XSS Prevention

### Player Name Sanitization

All player names are sanitized using `utils.SanitizePlayerName()`:

```go
func SanitizePlayerName(name string) string {
    // Trim whitespace
    name = strings.TrimSpace(name)
    // Escape HTML
    name = html.EscapeString(name)
    // Limit length
    runes := []rune(name)
    if len(runes) > 100 {
        name = string(runes[:100])
    }
    return name
}
```

### Example

```
Input:  <script>alert('xss')</script>
Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;
```

## Token Masking

### Logging Best Practices

Raw tokens are never logged. Instead, use `utils.MaskToken()`:

```go
func MaskToken(token string) string {
    if len(token) <= 10 {
        return "***"
    }
    return token[:6] + "..." + token[len(token)-4:]
}
```

### Example

```
Raw:    abcdefghij1234567890
Masked: abcdef...7890
```

## CSRF Protection (Ready for Implementation)

### CSRF Token Generation

```go
func GenerateCSRFToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("failed to generate CSRF token: %w", err)
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

### Middleware (Available but not yet applied)

The CSRF middleware is implemented in `internal/middleware/csrf.go` but not yet applied to routes. To enable:

```go
// In main.go or routes.go
e.Use(middleware.InjectCSRFToken)
e.Use(middleware.CSRFMiddleware)
```

## Cookie Security

### Session Cookie Configuration

```go
Store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   86400 * 30, // 30 days
    HttpOnly: true,        // Prevents JavaScript access
    Secure:   isProduction, // HTTPS only in production
    SameSite: http.SameSiteLaxMode,
}
```

## Testing

### Unit Tests

All security features have comprehensive unit tests:

- `internal/utils/session_test.go`: Session UUID and invitation codes
- `internal/utils/xss_test.go`: XSS prevention
- `internal/utils/csrf_test.go`: CSRF token generation

### Manual Testing Scenarios

1. **Conflict Detection**:
   - Open bag on Device A
   - Try to access same bag on Device B
   - Verify 409 Conflict response

2. **Invitation System**:
   - Generate invitation code on Device A
   - Accept invitation on Device B
   - Verify both devices can access bag

3. **Bag Release**:
   - Bind bag to Device A
   - Release bag from Device A
   - Verify Device B can now claim bag

4. **XSS Prevention**:
   - Enter `<script>alert('xss')</script>` as player name
   - Verify it's escaped in all displays

## Security Considerations

### Strengths

✅ **Cryptographically secure random**: All tokens use `crypto/rand`
✅ **Session isolation**: Each browser maintains separate session state
✅ **Database-backed validation**: All checks verified against database
✅ **Automatic cleanup**: Session bindings deleted on token reset
✅ **Masked logging**: No sensitive data in logs

### Limitations

⚠️ **Cookie-based sessions**: Sessions stored in cookies (consider Redis/database for high-scale)
⚠️ **No rate limiting**: Consider adding rate limits for invitation generation
⚠️ **Manual expiration cleanup**: Old invitation codes should be cleaned up periodically

### Future Enhancements

1. **Redis-based sessions**: Move session storage to Redis for scalability
2. **Rate limiting**: Add rate limits for API endpoints
3. **Audit logging**: Comprehensive audit trail for security events
4. **2FA support**: Optional two-factor authentication for sensitive operations
5. **CSRF enforcement**: Enable CSRF middleware globally

## Deployment Notes

### Environment Variables

No new environment variables required. Existing configuration is sufficient.

### Database Migration

Migrations run automatically on application startup via `database.runMigrations()`.

### Backwards Compatibility

✅ **Fully backwards compatible**: Existing tokens work without modification
✅ **Graceful degradation**: Old sessions automatically upgrade on next access
✅ **No data loss**: All existing data preserved

## Support

For issues or questions, please open a GitHub issue or contact the development team.
