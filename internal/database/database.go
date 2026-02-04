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

// runMigrations runs database migrations from SQL files
func runMigrations() {
	if err := runMigrationsFromFiles(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
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
