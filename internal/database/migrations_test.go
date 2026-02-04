package database

import (
	"testing"
)

// TestLoadMigrations verifies that migration files can be loaded correctly
func TestLoadMigrations(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("Failed to load migrations: %v", err)
	}

	if len(migrations) == 0 {
		t.Fatal("Expected at least one migration, got none")
	}

	// Verify migrations are sorted by version
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			t.Errorf("Migrations not sorted: version %d comes after %d", 
				migrations[i].Version, migrations[i-1].Version)
		}
	}

	// Verify each migration has required fields
	for _, migration := range migrations {
		if migration.Version <= 0 {
			t.Errorf("Invalid migration version: %d", migration.Version)
		}
		if migration.Description == "" {
			t.Errorf("Migration %d has no description", migration.Version)
		}
		if migration.SQL == "" {
			t.Errorf("Migration %d has no SQL content", migration.Version)
		}
	}
}

// TestMigrationFileEmbedding verifies that migration files are embedded correctly
func TestMigrationFileEmbedding(t *testing.T) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	sqlFileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".sql" {
			sqlFileCount++
		}
	}

	if sqlFileCount < 2 {
		t.Errorf("Expected at least 2 SQL migration files, found %d", sqlFileCount)
	}
}

// TestMigrationContentStructure verifies that migrations have proper SQL structure
func TestMigrationContentStructure(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("Failed to load migrations: %v", err)
	}

	for _, migration := range migrations {
		// Check for SQL comments in migrations
		if len(migration.SQL) < 10 {
			t.Errorf("Migration %d has suspiciously short SQL content: %d bytes", 
				migration.Version, len(migration.SQL))
		}
	}
}
