-- Migration: 003_add_session_binding_and_invitations.sql
-- Description: Adds session binding, conflict detection, and invitation system
-- Date: 2026-02-12

-- Add session_uuid to upload_tokens for per-browser session binding
ALTER TABLE upload_tokens ADD COLUMN IF NOT EXISTS session_uuid TEXT DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_upload_tokens_session_uuid ON upload_tokens(session_uuid) WHERE session_uuid IS NOT NULL;

-- Table: invitations
-- Manages invitation codes for allowing multiple users to access same token
CREATE TABLE IF NOT EXISTS invitations (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
    invitation_code TEXT NOT NULL UNIQUE,
    created_by_session_uuid TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_by_session_uuid TEXT DEFAULT NULL,
    accepted_at TIMESTAMPTZ DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_invitations_code ON invitations(invitation_code);
CREATE INDEX IF NOT EXISTS idx_invitations_token_id ON invitations(token_id);
CREATE INDEX IF NOT EXISTS idx_invitations_expires_at ON invitations(expires_at);

-- Table: active_sessions
-- Tracks all active sessions for a token (for multi-session support via invitations)
CREATE TABLE IF NOT EXISTS active_sessions (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
    session_uuid TEXT NOT NULL,
    player_name TEXT NOT NULL,
    player_city TEXT DEFAULT '',
    started_at TIMESTAMPTZ DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(token_id, session_uuid)
);

CREATE INDEX IF NOT EXISTS idx_active_sessions_token_id ON active_sessions(token_id);
CREATE INDEX IF NOT EXISTS idx_active_sessions_session_uuid ON active_sessions(session_uuid);
CREATE INDEX IF NOT EXISTS idx_active_sessions_last_activity ON active_sessions(last_activity_at DESC);

-- Add CSRF token field for form protection
-- Note: CSRF tokens will be stored in session cookies for now, not in DB
