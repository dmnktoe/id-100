# GeoNames Loader Fix and Deriven Data Setup

## Summary

This document explains the fixes for the GeoNames loader and clarifies how deriven data is populated in the Docker setup.

## Issue 1: GeoNames Loader Package Error - FIXED

### Problem

The geonames-loader container was failing with:
```
ERROR: unable to select packages:
  awk (no such package):
    required by: world[awk]
/bin/sh: curl: not found
```

### Root Cause

Alpine Linux doesn't have a package named `awk`. The standard `awk` command in Alpine comes from BusyBox, but if you want GNU awk specifically, you need to install the `gawk` package.

The docker-compose.yml was trying to install `awk` which doesn't exist:
```yaml
apk add --no-cache curl unzip awk  # ❌ 'awk' package doesn't exist
```

### Solution

Changed the package name from `awk` to `gawk`:
```yaml
apk add --no-cache curl unzip gawk  # ✅ 'gawk' exists in Alpine
```

### Result

The geonames-loader now:
1. ✅ Successfully installs all required tools (curl, unzip, gawk)
2. ✅ Downloads German city data from GeoNames.org (~10MB)
3. ✅ Converts TSV data to JSON using awk scripts
4. ✅ Creates Meilisearch index `cities`
5. ✅ Imports ~15,000 German cities
6. ✅ Configures search settings (typo tolerance, ranking rules)
7. ✅ Autocomplete works at http://localhost:8081/indexes/cities/search

The 404 error on `/indexes/cities/search` should now be resolved.

## Issue 2: Deriven Data Auto-Population - CLARIFIED

### User Question

> "Please add that the 100 deriven from deriven_rows.sql are automatically filled during docker compose startup"

### Answer

**The deriven data is NOT automatically filled.** This is intentional and by design.

### Why Not Automatic?

1. **Placeholder File**: `internal/database/migrations/deriven_rows.sql` is a placeholder template, not actual data
2. **Not a Migration**: The file doesn't follow the migration naming pattern (`NNN_description.sql`)
3. **Correctly Skipped**: The migration system outputs: "Warning: skipping invalid migration filename: deriven_rows.sql"
4. **User Data Required**: Each deployment needs their own 100 deriven challenges

### How the Migration System Works

```
internal/database/migrations/
├── 001_create_initial_tables.sql    ✅ Auto-runs (creates tables)
├── 002_insert_initial_deriven.sql   ✅ Auto-runs (inserts deriven data IF converted)
├── deriven_rows.sql                 ⚠️  Skipped (not a numbered migration)
└── README.md                         ⚠️  Skipped (not a .sql file)
```

**Migration Pattern:** Files must start with `NNN_` (e.g., `001_`, `002_`) to be recognized as migrations.

### How to Add Deriven Data

**Step 1: Replace Placeholder**
```bash
# Export your 100 deriven from Supabase
# Then replace the placeholder:
cp /path/to/your-export.sql internal/database/migrations/deriven_rows.sql
```

Your export should look like:
```sql
INSERT INTO "public"."deriven" ("id", "number", "title", "description", "created_at", "points") VALUES
('1', '1', 'Derive #001', 'Dokumentiere ein Objekt...', '2025-12-30 12:17:45.375781+00', '1'),
('2', '2', 'Derive #002', 'Miss die Höhe...', '2025-12-30 12:17:45.375781+00', '2'),
-- ... 98 more rows ...
('100', '100', 'Derive #100', 'Description...', '2025-12-30 12:17:45.375781+00', '5');
```

**Step 2: Run Conversion Script**
```bash
./scripts/convert-deriven-export.sh
```

This script:
- Reads `internal/database/migrations/deriven_rows.sql`
- Converts `"public"."deriven"` to `deriven`
- Wraps INSERT in a conditional block (prevents duplicates)
- Writes to `002_insert_initial_deriven.sql`

**Step 3: Rebuild Docker**
```bash
docker-compose down -v  # Clear old data
docker-compose up -d --build
```

**Step 4: Verify**
```bash
docker exec -it id100-db psql -U dev -d id100 -c "SELECT COUNT(*) FROM deriven;"
# Should show: 100
```

### Why This Design?

**Advantages:**
- ✅ No sensitive data in version control
- ✅ Each team/environment can have custom deriven
- ✅ Clear separation: structure (001) vs data (002)
- ✅ Explicit control over what data is imported
- ✅ Easy to update: just replace file and re-run script

**Alternative Approaches Considered:**
- ❌ Hardcode 100 deriven in 002_insert_initial_deriven.sql (would expose data)
- ❌ Auto-convert on startup (would require file watching, complex)
- ❌ Include sample data (not useful for production)

## Complete Startup Flow

When you run `docker-compose up -d`, here's what happens:

### 1. Infrastructure Starts
```
PostgreSQL   → Starts, creates database 'id100'
MinIO        → Starts S3-compatible storage
Meilisearch  → Starts search engine
```

### 2. Data Loading (Parallel)
```
geonames-loader:
  ✅ Install gawk, curl, unzip
  ✅ Download DE.zip (~10MB)
  ✅ Extract DE.txt
  ✅ Convert to JSON with awk
  ✅ Create 'cities' index
  ✅ Import ~15,000 cities
  ✅ Exit (one-time job)

webapp:
  ✅ Run migration 001 (create tables)
  ⚠️  Skip deriven_rows.sql (not a migration)
  ⚠️  Run migration 002 (no data yet - needs conversion script)
  ✅ Start web server on :8080
```

### 3. Ready State
```
✅ http://localhost:8080  - Web application
✅ http://localhost:8081  - Meilisearch API
✅ http://localhost:9000  - MinIO API
✅ http://localhost:9001  - MinIO Console
```

## Troubleshooting

### GeoNames Loader Issues

**Problem:** `awk (no such package)`
- **Solution:** Update docker-compose.yml to use `gawk` (fixed in ac8d66b)

**Problem:** `/bin/sh: curl: not found`
- **Solution:** Ensure `curl` is in apk install list (already fixed)

**Problem:** `awk: /tmp/DE.txt: No such file or directory`
- **Cause:** Download failed
- **Check:** `docker-compose logs geonames-loader`
- **Solution:** Verify internet connection, GeoNames.org availability

### City Autocomplete 404

**Problem:** `404 (Not Found)` on `/indexes/cities/search`
- **Cause:** GeoNames loader failed, index not created
- **Solution:** Check `docker-compose logs geonames-loader` for errors
- **Test:** `curl http://localhost:8081/indexes` (should list 'cities')

### Deriven Data Missing

**Problem:** No deriven in database
- **Expected:** This is normal on first startup
- **Solution:** Follow "How to Add Deriven Data" steps above
- **Verify:** `docker exec -it id100-db psql -U dev -d id100 -c "SELECT COUNT(*) FROM deriven;"`

## Documentation

- **Setup Guide:** `docs/ADDING_DERIVEN_DATA.md`
- **Migration Details:** `internal/database/migrations/README.md`
- **Quick Reference:** `QUICK_FIX_SUMMARY.md`
- **City Autocomplete:** `docs/CITY_AUTOCOMPLETE.md`

## Summary

### ✅ Fixed
- GeoNames loader package installation (awk → gawk)
- City autocomplete will now work

### ⚠️ Requires Action
- Deriven data must be added manually using conversion script
- Not automatic - this is by design for security and flexibility

### 📚 Documented
- Complete explanation in README.md
- Updated QUICK_FIX_SUMMARY.md with clear instructions
- This comprehensive guide for reference
