# Adding Initial Deriven Data

This guide explains how to populate the database with initial deriven (challenge) data.

## Overview

The application uses a migration system that automatically runs SQL files in the `internal/database/migrations/` directory on startup. **As of the latest update, the conversion happens automatically** - you just need to replace the placeholder file before starting Docker.

## Quick Method (Automatic - Recommended)

1. **Replace the placeholder with your Supabase export:**
   ```bash
   # Copy your export over the placeholder
   cp /path/to/your/supabase-export.sql internal/database/migrations/deriven_rows.sql
   ```

2. **Start Docker (conversion happens automatically):**
   ```bash
   docker-compose up -d --build
   ```

**That's it!** The startup script (`scripts/startup.sh`) will automatically:
- ✅ Detect your `deriven_rows.sql` file
- ✅ Check if it contains actual data (not just placeholder)
- ✅ Convert it to `002_insert_initial_deriven.sql` format
- ✅ The migration system will then apply it

3. **Verify:**
   ```bash
   # Check that all 100 deriven were inserted
   docker exec -it id100-db psql -U dev -d id100 -c "SELECT COUNT(*) FROM deriven;"
   ```

## Manual Method (Advanced)

If you need to run the conversion manually:

1. Replace the placeholder with your Supabase export:
   ```bash
   cp /path/to/your/supabase-export.sql internal/database/migrations/deriven_rows.sql
   ```

2. Run the conversion script manually:
   ```bash
   ./scripts/convert-deriven-export.sh
   ```

3. Restart Docker containers:
   ```bash
   docker-compose down -v  # -v removes old data
   docker-compose up -d --build
   ```

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
