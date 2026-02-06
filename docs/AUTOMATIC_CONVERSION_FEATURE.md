# Automatic Deriven Conversion Feature

## Summary

This document explains the automatic deriven data conversion feature that was added to simplify the Docker setup process.

## Problem Solved

**Before:** Users had to manually run a conversion script after replacing the placeholder file:
```bash
cp export.sql internal/database/migrations/deriven_rows.sql
./scripts/convert-deriven-export.sh                          # ❌ Manual step
docker-compose down -v && docker-compose up -d --build
```

**After:** Conversion happens automatically on container startup:
```bash
cp export.sql internal/database/migrations/deriven_rows.sql
docker-compose up -d --build                                  # ✅ Automatic!
```

## How It Works

### The Startup Script

`scripts/startup.sh` runs as the container's ENTRYPOINT before the main application:

```bash
Container Startup Sequence:
┌─────────────────────────────────────┐
│ 1. Docker starts container          │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 2. ENTRYPOINT: startup.sh           │
│    ├─ Check for deriven_rows.sql    │
│    ├─ Count INSERT statements       │
│    ├─ If data found: convert        │
│    │  to 002_insert_initial_deriven │
│    └─ If placeholder: skip          │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 3. CMD: /app/id-100                 │
│    ├─ Run migrations                │
│    │  ├─ 001 (tables)              │
│    │  └─ 002 (deriven data)        │
│    └─ Start web server             │
└─────────────────────────────────────┘
```

### Smart Detection

The script intelligently detects whether to run the conversion:

```bash
# Counts actual INSERT statements
ACTUAL_LINES=$(grep -c "^INSERT INTO" "$INPUT_FILE" 2>/dev/null || echo "0")

if [ "$ACTUAL_LINES" -gt 0 ]; then
    # Real data found - convert it!
    echo "==> Converting deriven_rows.sql to migration format..."
    # ... conversion logic ...
else
    # Placeholder file - skip
    echo "==> Skipping conversion: deriven_rows.sql contains no INSERT statements"
fi
```

This means:
- ✅ If you replace the file with real data: **automatic conversion**
- ✅ If you keep the placeholder: **no error, just skipped**
- ✅ Safe to run repeatedly: **idempotent**

### The Conversion

When data is detected, the script:

1. **Creates migration header:**
   ```sql
   -- Migration: 002_insert_initial_deriven.sql
   -- Description: Inserts initial derive challenges
   -- Auto-converted on container startup
   
   DO $$
   BEGIN
       IF NOT EXISTS (SELECT 1 FROM deriven LIMIT 1) THEN
   ```

2. **Converts and appends data:**
   ```bash
   # Change "public"."deriven" to deriven
   sed 's/"public"\."deriven"/deriven/g' "$INPUT_FILE" >> "$OUTPUT_FILE"
   ```

3. **Closes the block:**
   ```sql
       END IF;
   END $$;
   ```

The result is a proper migration file that:
- Only inserts if table is empty (prevents duplicates)
- Uses correct table name (removes "public" prefix)
- Is transaction-safe

## Files Changed

### New Files

**scripts/startup.sh** - The automatic conversion script
- Runs before main application
- Detects deriven_rows.sql
- Converts if needed
- Starts the app via `exec "$@"`

### Modified Files

**Dockerfile** - Uses startup script as ENTRYPOINT
```dockerfile
# Before:
CMD ["/app/id-100"]

# After:
ENTRYPOINT ["/app/scripts/startup.sh"]
CMD ["/app/id-100"]
```

This allows the startup script to run first, then pass control to the main app.

## User Workflow

### For Development

```bash
# 1. Replace placeholder with your 100 deriven
cp ~/Downloads/deriven_export.sql internal/database/migrations/deriven_rows.sql

# 2. Start Docker (conversion happens automatically!)
docker-compose up -d --build

# 3. Verify data was loaded
docker exec -it id100-db psql -U dev -d id100 -c "SELECT COUNT(*) FROM deriven;"
# Should output: 100

# 4. Check logs to see conversion happened
docker logs id100-webapp 2>&1 | head -20
# Should see: "==> Converting deriven_rows.sql to migration format..."
```

### For Production

Same workflow! The conversion is safe and idempotent:
- First run: Converts data and applies migration
- Subsequent runs: Skips conversion (data already in DB)

## Benefits

### 1. Simpler Workflow
No manual script execution needed. Just replace file and start Docker.

### 2. Fewer Steps
Reduced from 4 steps to 2 steps.

### 3. Less Error-Prone
Users can't forget to run the conversion script.

### 4. Better UX
More intuitive: "replace file, start Docker" vs "replace file, run script, start Docker".

### 5. Consistent with Docker Philosophy
Everything happens in containers - no local script execution required.

## Technical Details

### Why ENTRYPOINT instead of CMD?

**ENTRYPOINT** runs before CMD and can wrap the main process:
```dockerfile
ENTRYPOINT ["/app/scripts/startup.sh"]  # Runs first
CMD ["/app/id-100"]                      # Passed to startup.sh as $@
```

Inside startup.sh:
```bash
# Do startup tasks...
echo "==> Starting application..."
exec "$@"  # Execute CMD with current PID (replaces shell process)
```

Using `exec` is important:
- Replaces the shell process with the app process
- App becomes PID 1 in the container
- Signals (SIGTERM, etc.) are properly received by the app

### Why Check for INSERT Statements?

The placeholder file is intentionally left in place for documentation:
```sql
-- deriven_rows.sql
-- Placeholder file for Supabase export
-- 
-- INSTRUCTIONS:
-- 1. Export your deriven table from Supabase
-- 2. Replace this file with your export
-- ...
```

Without the INSERT check, the script would try to convert the placeholder comments, resulting in a malformed migration.

By checking for actual `INSERT INTO` statements, we:
- Skip placeholder files (no data = no conversion)
- Only convert when user has added real data
- Avoid errors and confusion

### Transaction Safety

The migration system wraps everything in transactions:
```go
tx, err := DB.Begin(ctx)
// ... run migration SQL ...
tx.Commit()  // Only if successful
```

Combined with the `IF NOT EXISTS` check in the migration:
- Atomic operation (all or nothing)
- No partial data inserts
- Safe to retry on failure

## Troubleshooting

### Conversion Doesn't Run

**Check logs:**
```bash
docker logs id100-webapp 2>&1 | grep "==> "
```

**Expected output if conversion ran:**
```
==> Running startup tasks...
==> Found deriven_rows.sql, checking if conversion is needed...
==> Converting deriven_rows.sql to migration format...
==> ✓ Conversion complete!
    Created: /app/internal/database/migrations/002_insert_initial_deriven.sql
==> Starting application...
```

**Expected output if skipped (placeholder):**
```
==> Running startup tasks...
==> Found deriven_rows.sql, checking if conversion is needed...
==> Skipping conversion: deriven_rows.sql contains no INSERT statements
    (Placeholder file detected)
==> Starting application...
```

### Data Not Inserted

1. **Check if conversion created the file:**
   ```bash
   docker exec id100-webapp cat /app/internal/database/migrations/002_insert_initial_deriven.sql | head -20
   ```

2. **Check migration status:**
   ```bash
   docker exec -it id100-db psql -U dev -d id100 -c "SELECT * FROM schema_migrations;"
   ```
   
   Should show:
   ```
   version |           name            |          applied_at
   --------+---------------------------+-------------------------------
        1 | create_initial_tables     | 2024-02-06 14:27:40.123456+00
        2 | insert_initial_deriven    | 2024-02-06 14:27:40.234567+00
   ```

3. **Check for errors in migration log:**
   ```bash
   docker logs id100-webapp 2>&1 | grep -i "migration\|error"
   ```

### Manually Re-run Conversion

If you need to re-run after updating the file:

```bash
# Option 1: Rebuild container (recommended)
docker-compose up -d --build webapp

# Option 2: Run conversion manually in container
docker exec id100-webapp /app/scripts/startup.sh echo "Done"

# Option 3: Run local script (if still needed)
./scripts/convert-deriven-export.sh
docker-compose restart webapp
```

## Migration Path

For existing installations that were using the manual script:

**No changes needed!** The old workflow still works:
```bash
# Old method (still supported)
./scripts/convert-deriven-export.sh
docker-compose up -d
```

But the new automatic method is simpler:
```bash
# New method (automatic)
docker-compose up -d --build
```

## Future Enhancements

Potential improvements:
- Add validation of deriven data format
- Support for multiple export formats
- Progress indicator for large datasets
- Backup of existing data before insert

## Conclusion

The automatic conversion feature makes Docker setup significantly easier while maintaining safety and flexibility. Users can now focus on their data rather than remembering to run scripts.

**Key Takeaway:** Replace `deriven_rows.sql` and run `docker-compose up -d --build`. Everything else is automatic!
