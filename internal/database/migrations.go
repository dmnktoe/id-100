package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// runMigrations executes all pending database migrations
func runMigrations() {
	ctx := context.Background()

	// Create migrations tracking table
	createMigrationsTable(ctx)

	// Get all migrations
	migrations, err := loadMigrations()
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// Get applied migrations
	applied := getAppliedMigrations(ctx)

	// Run pending migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			log.Printf("Migration %03d_%s already applied, skipping", migration.Version, migration.Name)
			continue
		}

		log.Printf("Running migration %03d_%s...", migration.Version, migration.Name)

		tx, err := DB.Begin(ctx)
		if err != nil {
			log.Fatalf("Failed to begin transaction for migration %d: %v", migration.Version, err)
		}

		// Execute migration SQL
		_, err = tx.Exec(ctx, migration.SQL)
		if err != nil {
			tx.Rollback(ctx)
			log.Fatalf("Failed to execute migration %d: %v", migration.Version, err)
		}

		// Record migration as applied
		_, err = tx.Exec(ctx,
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, NOW())",
			migration.Version, migration.Name)
		if err != nil {
			tx.Rollback(ctx)
			log.Fatalf("Failed to record migration %d: %v", migration.Version, err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Failed to commit migration %d: %v", migration.Version, err)
		}

		log.Printf("Migration %03d_%s applied successfully", migration.Version, migration.Name)
	}

	log.Println("All migrations completed successfully")
}

// createMigrationsTable creates the schema_migrations tracking table
func createMigrationsTable(ctx context.Context) {
	_, err := DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}
}

// loadMigrations reads all migration files from the embedded filesystem
func loadMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Parse migration filename: 001_create_tables.sql
		name := entry.Name()
		var version int
		_, err := fmt.Sscanf(name, "%d_", &version)
		if err != nil {
			log.Printf("Warning: skipping invalid migration filename: %s", name)
			continue
		}

		// Read migration content
		content, err := migrationsFS.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		// Extract name without version and extension
		migrationName := strings.TrimSuffix(name, ".sql")
		migrationName = strings.TrimPrefix(migrationName, fmt.Sprintf("%03d_", version))

		migrations = append(migrations, Migration{
			Version: version,
			Name:    migrationName,
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(ctx context.Context) map[int]bool {
	rows, err := DB.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		log.Printf("Warning: failed to query applied migrations: %v", err)
		return make(map[int]bool)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			log.Printf("Warning: failed to scan migration version: %v", err)
			continue
		}
		applied[version] = true
	}

	return applied
}
