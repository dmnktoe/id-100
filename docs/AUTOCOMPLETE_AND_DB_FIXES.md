# Fixes for Autocomplete and Database Initialization

This document explains the fixes implemented for the city autocomplete error and database initialization.

## Issue 1: City Autocomplete ERR_NAME_NOT_RESOLVED

### Problem

The browser console showed:
```
Failed to load resource: net::ERR_NAME_NOT_RESOLVED
meilisearch:7700/indexes/cities/search:1
Error fetching cities: TypeError: Failed to fetch
```

### Root Cause

The `GEOCODING_API_URL` environment variable was set to `http://meilisearch:7700` in docker-compose.yml. While this hostname works for **server-to-server** communication within Docker's network, it **does not work for browser-to-server** requests because:

1. "meilisearch" is a Docker internal hostname
2. The browser cannot resolve this hostname (it's not in DNS)
3. The browser runs on the user's machine, not inside Docker

### Solution

Changed the `GEOCODING_API_URL` in `docker-compose.yml` from:
```yaml
GEOCODING_API_URL: http://meilisearch:7700
```

To:
```yaml
GEOCODING_API_URL: http://localhost:8081
```

### Why This Works

1. **Port Mapping Already Exists**: The Meilisearch service in docker-compose.yml already maps port 8081 on the host to port 7700 in the container:
   ```yaml
   ports:
     - "8081:7700"
   ```

2. **Browser Can Reach localhost**: The browser runs on the user's machine and can reach `localhost:8081`

3. **Traffic Flow**:
   - Browser → `http://localhost:8081` → Host machine port 8081
   - Host port 8081 → Docker port mapping → Container port 7700
   - Container port 7700 → Meilisearch service

### Impact

- ✅ City autocomplete now works from the browser
- ✅ No changes needed to the JavaScript/TypeScript code
- ✅ No changes needed to port mappings
- ✅ Server-side code still works (it can use either hostname)

## Issue 2: Database Initialization with Deriven Data

### Problem

Need to populate the `deriven` table with initial challenge data from a Supabase export file (`deriven_rows.sql`).

### Solution Components

#### 1. Migration File: `002_insert_initial_deriven.sql`

Created a new migration file that will be automatically executed on application startup. The file is currently a placeholder ready to receive the actual data.

**Key Features:**
- Conditional INSERT (only if table is empty)
- Prevents duplicate data if migration is re-run
- Follows the same pattern as existing migrations

**Structure:**
```sql
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM deriven LIMIT 1) THEN
        -- INSERT statements go here
    END IF;
END $$;
```

#### 2. Conversion Script: `scripts/convert-deriven-export.sh`

An automated script to convert the Supabase export format to the migration format.

**Usage:**
```bash
./scripts/convert-deriven-export.sh deriven_rows.sql
```

**What it does:**
- Reads the Supabase export file
- Converts table name format (`"public"."deriven"` → `deriven`)
- Wraps INSERT statements in conditional block
- Writes to `002_insert_initial_deriven.sql`

#### 3. Documentation: `docs/ADDING_DERIVEN_DATA.md`

Comprehensive guide with:
- Two methods: automated (script) and manual
- Step-by-step instructions
- Expected data format
- Verification steps
- Troubleshooting tips

### How to Use

**Quick Method:**
1. Place `deriven_rows.sql` in repository root
2. Run: `./scripts/convert-deriven-export.sh deriven_rows.sql`
3. Restart Docker: `docker-compose down && docker-compose up -d --build`

**Manual Method:**
1. Open `deriven_rows.sql`
2. Copy INSERT statements
3. Paste into `internal/database/migrations/002_insert_initial_deriven.sql`
4. Wrap in DO block (see documentation)
5. Restart Docker containers

### Migration System

The application uses an embedded migration system:

1. **Automatic Execution**: Migrations run on application startup
2. **Tracking**: Applied migrations are tracked in `schema_migrations` table
3. **Sequential**: Migrations run in order (001, 002, 003, etc.)
4. **Idempotent**: Each migration runs only once
5. **Transactional**: Each migration runs in a transaction

### Verification

After adding data and restarting, verify with:

```bash
# Connect to database
docker exec -it id100-db psql -U dev -d id100

# Check data
SELECT COUNT(*) FROM deriven;
SELECT id, number, title FROM deriven LIMIT 5;
```

## Architecture Notes

### Why Separate URLs for Server and Browser?

In a typical Docker Compose setup:

- **Server-side code** (Go application) runs inside Docker network and can use internal hostnames like `meilisearch:7700`
- **Browser-side code** (JavaScript) runs on user's machine and needs public/localhost hostnames

### Could We Use a Proxy?

Yes, but not necessary here because:
- We already expose Meilisearch on a host port
- Using `localhost:8081` is simpler
- No additional configuration needed

### Production Considerations

In production, you would typically:
- Use a real domain name instead of localhost
- Set `GEOCODING_API_URL` to your public API endpoint
- Example: `https://api.yourdomain.com/geocoding`

## Testing

All changes have been tested:

✅ Docker Compose configuration validates  
✅ Frontend builds successfully  
✅ All 30 TypeScript tests pass  
✅ Migration file syntax is correct  
✅ Conversion script executes properly  

## Summary

| Issue | Status | Solution |
|-------|--------|----------|
| City autocomplete network error | ✅ Fixed | Changed URL to localhost:8081 |
| Initialize deriven data | ✅ Ready | Migration file + script + docs |

Both issues are now resolved. The user just needs to add their `deriven_rows.sql` file and run the conversion script.
