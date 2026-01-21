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
}
