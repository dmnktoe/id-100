# Adding Initial Deriven Data

This guide explains how to populate the database with initial deriven (challenge) data.

## Overview

The application uses a migration system that automatically runs SQL files in the `internal/database/migrations/` directory on startup. To add initial deriven data, you need to populate the migration file `002_insert_initial_deriven.sql`.

## Method 1: Using the Conversion Script (Recommended)

A placeholder file `deriven_rows.sql` already exists at `internal/database/migrations/`. Simply replace it with your Supabase export:

1. Replace the placeholder with your Supabase export:
   ```bash
   # Copy your export over the placeholder
   cp /path/to/your/supabase-export.sql internal/database/migrations/deriven_rows.sql
   ```

2. Run the conversion script:
   ```bash
   # The script will automatically use the default location
   ./scripts/convert-deriven-export.sh
   ```

3. The script will automatically update `internal/database/migrations/002_insert_initial_deriven.sql`

4. Restart Docker containers to apply the migration:
   ```bash
   docker-compose down -v  # -v removes old data
   docker-compose up -d --build
   ```

**Note:** The placeholder file includes helpful comments showing the expected format and instructions.

## Method 2: Manual Copy-Paste

If you prefer to manually add the data:

1. Open your `deriven_rows.sql` export file
2. Copy all the INSERT statements
3. Open `internal/database/migrations/002_insert_initial_deriven.sql`
4. Replace the placeholder comment with your INSERT statements
5. Change `"public"."deriven"` to just `deriven` (if needed)
6. Wrap the INSERTs in a conditional block to avoid duplicates:

```sql
-- Migration: 002_insert_initial_deriven.sql
-- Description: Inserts initial derive challenges into the deriven table

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM deriven LIMIT 1) THEN
        -- Your INSERT statements here
        INSERT INTO deriven ("id", "number", "title", "description", "created_at", "points") VALUES
        ('1', '1', 'Derive #001', 'Description...', '2025-12-30 12:17:45.375781+00', '1'),
        ('2', '2', 'Derive #002', 'Description...', '2025-12-30 12:17:45.375781+00', '2'),
        -- ... more rows
        ('100', '100', 'Derive #100', 'Description...', '2025-12-30 12:17:45.375781+00', '100');
    END IF;
END $$;
```

## Expected Format

The deriven table has the following structure:

```sql
CREATE TABLE deriven (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    points INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

Your INSERT statements should match this structure:

```sql
INSERT INTO deriven ("id", "number", "title", "description", "created_at", "points") VALUES
('1', '1', 'Derive #001', 'Dokumentiere ein Objekt...', '2025-12-30 12:17:45.375781+00', '1'),
('2', '2', 'Derive #002', 'Miss die Höhe von...', '2025-12-30 12:17:45.375781+00', '2');
```

## Verification

After restarting the containers, verify the data was inserted:

```bash
# Connect to the database
docker exec -it id100-db psql -U dev -d id100

# Check the deriven table
SELECT COUNT(*) FROM deriven;
SELECT id, number, title FROM deriven LIMIT 5;

# Exit
\q
```

## Troubleshooting

### Migration already applied

If the migration was already applied but with empty data:

1. Connect to the database:
   ```bash
   docker exec -it id100-db psql -U dev -d id100
   ```

2. Delete the migration record:
   ```sql
   DELETE FROM schema_migrations WHERE version = 2;
   ```

3. Restart the containers:
   ```bash
   docker-compose restart webapp
   ```

### Duplicate key errors

If you get duplicate key errors, the data might already be in the database. You can either:

- Clear the database: `docker-compose down -v` (WARNING: deletes all data)
- Or update the migration to use `INSERT ... ON CONFLICT DO NOTHING`

## Notes

- The migration runs automatically on application startup
- Migrations are tracked in the `schema_migrations` table
- Each migration runs only once
- The conditional block (`IF NOT EXISTS`) prevents duplicate insertions if the migration is re-run
