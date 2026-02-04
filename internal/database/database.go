package database

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the database connection pool
var DB *pgxpool.Pool

// Init initializes the database connection and runs migrations
func Init() {
	connStr := os.Getenv("DATABASE_URL")

	var err error
	DB, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Run migrations
	runMigrations()
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// runMigrations runs simple database migrations
func runMigrations() {
	// Ensure bag_requests table exists
	_, err := DB.Exec(context.Background(), `
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
	_, err = DB.Exec(context.Background(), `ALTER TABLE bag_requests ADD COLUMN IF NOT EXISTS handled BOOLEAN DEFAULT FALSE`)
	if err != nil {
		log.Printf("Failed to add handled column to bag_requests: %v", err)
	}

	// Ensure contributions have a user_city column and upload_tokens track current_player_city
	_, err = DB.Exec(context.Background(), `ALTER TABLE contributions ADD COLUMN IF NOT EXISTS user_city TEXT DEFAULT ''`)
	if err != nil {
		log.Printf("Failed to add user_city column to contributions: %v", err)
	}
	_, err = DB.Exec(context.Background(), `ALTER TABLE upload_tokens ADD COLUMN IF NOT EXISTS current_player_city TEXT DEFAULT ''`)
	if err != nil {
		log.Printf("Failed to add current_player_city column to upload_tokens: %v", err)
	}
}

// GetFooterStats fetches creative database statistics
func GetFooterStats() (totalDeriven, totalContribs, activeUsers int, lastActivity sql.NullTime) {
	// Count total deriven
	DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM deriven").Scan(&totalDeriven)

	// Count total contributions
	DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM contributions").Scan(&totalContribs)

	// Count active users (users who contributed)
	DB.QueryRow(context.Background(), "SELECT COUNT(DISTINCT user_name) FROM contributions WHERE user_name != ''").Scan(&activeUsers)

	// Get last activity timestamp
	err := DB.QueryRow(context.Background(), "SELECT MAX(created_at) FROM contributions").Scan(&lastActivity)
	if err != nil {
		log.Printf("Error fetching last activity: %v", err)
	}

	return
}
