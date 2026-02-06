-- Migration: 001_create_initial_tables.sql
-- Description: Creates the core tables for the ID-100 application
-- Date: 2024-02-06

-- Table: deriven
-- Stores the main derive/challenge definitions
CREATE TABLE IF NOT EXISTS deriven (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    points INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deriven_number ON deriven(number);

-- Table: contributions
-- Stores user contributions/uploads for derives
CREATE TABLE IF NOT EXISTS contributions (
    id SERIAL PRIMARY KEY,
    derive_id INTEGER NOT NULL REFERENCES deriven(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    image_lqip TEXT DEFAULT '',
    user_name TEXT NOT NULL,
    user_city TEXT DEFAULT '',
    user_comment TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contributions_derive_id ON contributions(derive_id);
CREATE INDEX IF NOT EXISTS idx_contributions_created_at ON contributions(created_at DESC);

-- Table: upload_tokens
-- Manages upload tokens and user sessions
CREATE TABLE IF NOT EXISTS upload_tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    bag_name TEXT NOT NULL,
    current_player TEXT DEFAULT '',
    current_player_city TEXT DEFAULT '',
    is_active BOOLEAN DEFAULT TRUE,
    max_uploads INTEGER DEFAULT 100,
    total_uploads INTEGER DEFAULT 0,
    total_sessions INTEGER DEFAULT 0,
    session_started_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_upload_tokens_token ON upload_tokens(token);
CREATE INDEX IF NOT EXISTS idx_upload_tokens_is_active ON upload_tokens(is_active);

-- Table: upload_logs
-- Tracks all uploads for auditing and session management
CREATE TABLE IF NOT EXISTS upload_logs (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
    contribution_id INTEGER REFERENCES contributions(id) ON DELETE SET NULL,
    derive_number INTEGER NOT NULL,
    player_name TEXT NOT NULL,
    session_number INTEGER NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_upload_logs_token_session ON upload_logs(token_id, session_number);
CREATE INDEX IF NOT EXISTS idx_upload_logs_uploaded_at ON upload_logs(uploaded_at DESC);

-- Table: bag_requests
-- Stores bag/tool requests from users
CREATE TABLE IF NOT EXISTS bag_requests (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    handled BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_bag_requests_handled ON bag_requests(handled);
CREATE INDEX IF NOT EXISTS idx_bag_requests_created_at ON bag_requests(created_at DESC);
