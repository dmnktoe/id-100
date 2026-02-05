# Session Invitation System

## Overview

The session invitation system allows multiple players to collaborate on the same token/bag simultaneously. This enables friends to play together while maintaining security and access control.

## How It Works

### Primary Session
When a user scans a QR code and enters their name, they become the **primary session holder**. This session has full control over:
- Uploading photos
- Generating invitations
- Managing other sessions (revoking access)
- Releasing the bag

### Invited Sessions
The primary session holder can invite others by:
1. Clicking "Spieler einladen" (Invite Player)
2. Generating an invitation link
3. Sharing the link with friends

Invited players:
- Get their own authenticated session
- Can upload photos independently
- Appear in the session management view
- Cannot revoke other sessions (only primary can do this)

## API Endpoints

### POST /upload/invite
Generates a new invitation link.

**Authentication:** Requires valid token and active session

**Rate Limit:** 10 invitations per hour per session

**Request:**
```http
POST /upload/invite HTTP/1.1
```

**Optional Query Parameters:**
- `hours` - Expiration time in hours (default: 24, max: 168)

**Response:**
```json
{
  "status": "success",
  "invitation_id": 123,
  "invitation_code": "abc123...",
  "invitation_url": "https://id-100.de/upload/accept-invite?code=abc123...",
  "expires_at": "2026-02-06T17:00:00Z"
}
```

### GET /upload/accept-invite
Accepts an invitation and authorizes the current session.

**Query Parameters:**
- `code` (required) - The invitation code

**Flow:**
1. If user has no player name → Show name entry form
2. If invitation is valid → Authorize session and redirect to upload page
3. If invitation is invalid/expired → Show error page

### POST /upload/invite/set-name
Sets player name when accepting invitation.

**Form Data:**
- `player_name` (required) - Player's name
- `invitation_code` (required) - The invitation code
- `agree_privacy` (required) - Privacy consent checkbox

**Response:** Redirects to accept-invite endpoint with name now set

### GET /upload/sessions
Lists all active sessions for the current token.

**Authentication:** Requires valid token

**Response:**
```json
{
  "sessions": [
    {
      "session_uuid": "...",
      "player_name": "Alice",
      "created_at": "2026-02-05T12:00:00Z",
      "last_activity_at": "2026-02-05T17:00:00Z",
      "is_current": true
    }
  ],
  "current_player": "Alice"
}
```

### POST /upload/sessions/:uuid/revoke
Revokes access for a specific session.

**Authentication:** Only primary session can revoke

**Path Parameters:**
- `uuid` - The session UUID to revoke

**Response:**
```json
{
  "status": "success",
  "message": "Session revoked"
}
```

## Security Features

### Rate Limiting
- Maximum 10 invitations per hour per session
- Prevents invitation spam and abuse

### Invitation Expiration
- Default: 24 hours
- Maximum: 7 days (168 hours)
- Expired invitations cannot be accepted

### Session Validation
- Each request validates session authorization
- Inactive sessions are denied access
- Primary session can revoke any invited session

### Cookie Security
- `HttpOnly`: Prevents JavaScript access
- `Secure`: HTTPS-only in production
- `SameSite: Lax`: CSRF protection
- Session UUIDs: Cryptographically secure (32 bytes)

## Testing

Run invitation tests:
```bash
go test ./cmd/id-100 -v -run "TestInvitation"
```

Integration tests require a real database connection.
