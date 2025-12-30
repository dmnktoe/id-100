package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var db *pgxpool.Pool // Global erreichbar für main.go

func initDatabase() {
	godotenv.Load() // Lädt .env aus dem Root
	connStr := os.Getenv("DATABASE_URL")

	var err error
	db, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
}
