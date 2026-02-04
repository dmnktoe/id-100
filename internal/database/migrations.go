package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migration represents a single database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// runMigrationsFromFiles executes all pending SQL migration files
func runMigrationsFromFiles() error {
	ctx := context.Background()

	// First, ensure the schema_migrations table exists
	_, err := DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read all migration files
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get already applied migrations
	appliedVersions, err := getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			log.Printf("Migration %d (%s) already applied, skipping", migration.Version, migration.Description)
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Description)
		
		// Execute the migration SQL
		_, err := DB.Exec(ctx, migration.SQL)
		if err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Description, err)
		}

		// Record the migration as applied
		_, err = DB.Exec(ctx, `
			INSERT INTO schema_migrations (version, description) 
			VALUES ($1, $2)`,
			migration.Version, migration.Description)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		log.Printf("Successfully applied migration %d: %s", migration.Version, migration.Description)
	}

	log.Printf("All migrations applied successfully")
	return nil
}

// loadMigrations reads and parses all migration files from the embedded filesystem
func loadMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Skip the schema_migrations file as it's handled separately
		if entry.Name() == "000_schema_migrations.sql" {
			continue
		}

		// Parse version from filename (e.g., "001_initial_schema.sql" -> version 1)
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			log.Printf("Skipping invalid migration filename: %s", entry.Name())
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Skipping migration with invalid version: %s", entry.Name())
			continue
		}

		// Read the SQL content
		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Extract description from filename
		description := strings.TrimSuffix(parts[1], ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a map of already applied migration versions
func getAppliedMigrations() (map[int]bool, error) {
	rows, err := DB.Query(context.Background(), "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}
