package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var db *pgxpool.Pool

func initDatabase() {
	godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")

	var err error
	db, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure bag_requests table exists (simple migration)
	_, err = db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS bag_requests (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		handled BOOLEAN DEFAULT FALSE
	)`)
	if err != nil {
		log.Printf("Failed to ensure bag_requests table: %v", err)
	}

	// For existing installations, ensure the column exists
	_, err = db.Exec(context.Background(), `ALTER TABLE bag_requests ADD COLUMN IF NOT EXISTS handled BOOLEAN DEFAULT FALSE`)
	if err != nil {
		log.Printf("Failed to add handled column to bag_requests: %v", err)
	}

	// Create session_invitations table for multi-session support
	_, err = db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS session_invitations (
		id SERIAL PRIMARY KEY,
		token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
		invitation_code TEXT NOT NULL UNIQUE,
		invited_by_session_uuid TEXT NOT NULL,
		invited_session_uuid TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		expires_at TIMESTAMPTZ NOT NULL,
		accepted_at TIMESTAMPTZ,
		revoked_at TIMESTAMPTZ,
		is_active BOOLEAN DEFAULT TRUE,
		max_uses INTEGER DEFAULT 1,
		use_count INTEGER DEFAULT 0
	)`)
	if err != nil {
		log.Printf("Failed to ensure session_invitations table: %v", err)
	}

	// Create indexes for performance
	_, err = db.Exec(context.Background(), `CREATE INDEX IF NOT EXISTS idx_session_invitations_code ON session_invitations(invitation_code)`)
	if err != nil {
		log.Printf("Failed to create index on invitation_code: %v", err)
	}

	_, err = db.Exec(context.Background(), `CREATE INDEX IF NOT EXISTS idx_session_invitations_token_id ON session_invitations(token_id)`)
	if err != nil {
		log.Printf("Failed to create index on token_id: %v", err)
	}

	// Create authorized_sessions table to track all authorized sessions per token
	_, err = db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS authorized_sessions (
		id SERIAL PRIMARY KEY,
		token_id INTEGER NOT NULL REFERENCES upload_tokens(id) ON DELETE CASCADE,
		session_uuid TEXT NOT NULL,
		player_name TEXT,
		invitation_id INTEGER REFERENCES session_invitations(id) ON DELETE SET NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		last_activity_at TIMESTAMPTZ DEFAULT NOW(),
		expires_at TIMESTAMPTZ,
		is_active BOOLEAN DEFAULT TRUE,
		UNIQUE(token_id, session_uuid)
	)`)
	if err != nil {
		log.Printf("Failed to ensure authorized_sessions table: %v", err)
	}

	_, err = db.Exec(context.Background(), `CREATE INDEX IF NOT EXISTS idx_authorized_sessions_token_session ON authorized_sessions(token_id, session_uuid)`)
	if err != nil {
		log.Printf("Failed to create index on token_id, session_uuid: %v", err)
	}
}
