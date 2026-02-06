# File Reorganization: deriven_rows.sql Moved

## Summary

The `deriven_rows.sql` placeholder file has been moved from the repository root to `internal/database/migrations/` for better organization.

## What Changed

### File Location
- **Before:** `deriven_rows.sql` (in repository root)
- **After:** `internal/database/migrations/deriven_rows.sql`

### Script Behavior
The conversion script now uses the new location as the default:

**Before:**
```bash
./scripts/convert-deriven-export.sh deriven_rows.sql
```

**After:**
```bash
./scripts/convert-deriven-export.sh  # No argument needed!
```

The script automatically finds the file at `internal/database/migrations/deriven_rows.sql`.

## Benefits

### 1. Better Organization
All migration-related files are now in one place:
```
internal/database/migrations/
├── 001_create_initial_tables.sql
├── 002_insert_initial_deriven.sql
├── deriven_rows.sql (placeholder)
└── README.md
```

### 2. Clearer Structure
Users can now see all database migration files in the migrations directory, making it obvious where database-related files belong.

### 3. Simpler Workflow
The conversion script always uses the fixed location:
```bash
# Copy your export to the migrations directory
cp /path/to/export.sql internal/database/migrations/deriven_rows.sql

# Run the script (it automatically finds the file)
./scripts/convert-deriven-export.sh

# Restart Docker
docker-compose down -v && docker-compose up -d --build
```

## Updated Documentation

All documentation has been updated to reflect the new location:

- ✅ `QUICK_FIX_SUMMARY.md` - Quick reference with new path
- ✅ `docs/ADDING_DERIVEN_DATA.md` - Detailed guide updated
- ✅ `docs/PLACEHOLDER_FILE_ADDED.md` - Location updated
- ✅ `internal/database/migrations/README.md` - References "this directory"
- ✅ Placeholder file itself has updated instructions

## Testing

The conversion script has been tested and works correctly with the new default path:

```bash
$ ./scripts/convert-deriven-export.sh
Converting internal/database/migrations/deriven_rows.sql to internal/database/migrations/002_insert_initial_deriven.sql...
✓ Conversion complete!
```

## Migration Path for Existing Users

If you already have `deriven_rows.sql` in the root directory:

```bash
# Move it to the new location
mv deriven_rows.sql internal/database/migrations/deriven_rows.sql

# Or just copy your export to the new location
cp /path/to/export.sql internal/database/migrations/deriven_rows.sql
```

Then run the script as usual (no path argument needed).

## Why This Change?

1. **Consistency**: Database files should be with database files
2. **Discoverability**: Users looking at the migrations directory will see the placeholder
3. **Cleaner root**: Reduces clutter in the repository root
4. **Convention**: Follows common practice of keeping related files together

This change improves the overall organization and makes the repository structure more intuitive!
