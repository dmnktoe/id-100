-- Migration 001: Initial Schema
-- Creates all core tables for the id-100 application

-- Table: deriven
-- Stores the main derives/tasks that users can contribute to
CREATE TABLE IF NOT EXISTS deriven (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    points INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: contributions
-- Stores user contributions (images) for derives
CREATE TABLE IF NOT EXISTS contributions (
    id SERIAL PRIMARY KEY,
    derive_id INTEGER NOT NULL REFERENCES deriven(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    image_lqip TEXT,
    user_name TEXT DEFAULT '',
    user_city TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: upload_tokens
-- Manages authorization tokens for uploads with session tracking
CREATE TABLE IF NOT EXISTS upload_tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    bag_name TEXT,
    current_player TEXT,
    current_player_city TEXT DEFAULT '',
    is_active BOOLEAN DEFAULT TRUE,
    max_uploads INTEGER DEFAULT 10,
    total_uploads INTEGER DEFAULT 0,
    total_sessions INTEGER DEFAULT 1,
    session_started_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: upload_logs
-- Audit trail of all uploads with session information
CREATE TABLE IF NOT EXISTS upload_logs (
    id SERIAL PRIMARY KEY,
    token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
    contribution_id INTEGER NOT NULL REFERENCES contributions(id) ON DELETE CASCADE,
    derive_number INTEGER NOT NULL,
    player_name TEXT,
    session_number INTEGER NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: bag_requests
-- Stores user requests for new bags/tokens
CREATE TABLE IF NOT EXISTS bag_requests (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    handled BOOLEAN DEFAULT FALSE
);
