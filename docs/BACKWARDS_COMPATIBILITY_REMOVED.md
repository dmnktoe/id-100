# Backwards Compatibility Removed

## Summary

All backwards compatibility features have been removed from the codebase for simplicity and clarity.

## Changes Made

### 1. Conversion Script

**File:** `scripts/convert-deriven-export.sh`

**Before:**
- Accepted optional path argument
- Used default path if no argument provided
- Supported custom file locations

```bash
./scripts/convert-deriven-export.sh [optional-path]
./scripts/convert-deriven-export.sh /custom/path/export.sql  # worked
./scripts/convert-deriven-export.sh  # used default
```

**After:**
- No arguments accepted
- Always uses fixed path: `internal/database/migrations/deriven_rows.sql`
- Simple, single-purpose script

```bash
./scripts/convert-deriven-export.sh  # only way to use it
```

### 2. Environment Variable Fallback

**File:** `internal/config/geocoding.go`

**Before:**
- Checked `GEOCODING_API_URL` first
- Fell back to old `NOMINATIM_URL` if not found
- Maintained compatibility with old deployments

```go
func GetGeocodingURL() string {
    geocodingURL := os.Getenv("GEOCODING_API_URL")
    if geocodingURL == "" {
        // Fall back to old variable name
        geocodingURL = os.Getenv("NOMINATIM_URL")
    }
    if geocodingURL == "" {
        geocodingURL = DefaultGeocodingURL
    }
    return geocodingURL
}
```

**After:**
- Only checks `GEOCODING_API_URL`
- No fallback to old variable names
- Cleaner, simpler code

```go
func GetGeocodingURL() string {
    geocodingURL := os.Getenv("GEOCODING_API_URL")
    if geocodingURL == "" {
        geocodingURL = DefaultGeocodingURL
    }
    return geocodingURL
}
```

### 3. Documentation

All documentation updated to reflect the simplified approach:

- ✅ `docs/FILE_REORGANIZATION.md` - Removed "Backwards Compatible" section
- ✅ `docs/NAMING_CLEANUP.md` - Removed backwards compatibility explanation
- ✅ `QUICK_FIX_SUMMARY.md` - Removed compatibility note

## Benefits

### Simplicity
- One way to do things
- No confusing alternatives
- Clear expectations

### Maintainability
- Less code to maintain
- No fallback logic to test
- Easier to understand

### Clarity
- Script has one clear purpose
- Environment variables are explicit
- No "magic" fallbacks

## Migration Required

### For Environment Variables

If you were using the old `NOMINATIM_URL` variable, you must update to `GEOCODING_API_URL`:

```bash
# Old (no longer works)
NOMINATIM_URL=http://meilisearch:7700

# New (required)
GEOCODING_API_URL=http://meilisearch:7700
```

Update in:
- `.env` files
- `docker-compose.yml`
- Production environment configuration
- CI/CD pipelines

### For Conversion Script

The script no longer accepts arguments. Always place your export at the fixed location:

```bash
# Copy export to the expected location
cp /path/to/export.sql internal/database/migrations/deriven_rows.sql

# Run the script (no arguments)
./scripts/convert-deriven-export.sh
```

## Testing

All changes have been tested:

✅ Conversion script works with fixed path  
✅ Go tests pass (config package)  
✅ Code compiles successfully  
✅ Documentation updated  

## Result

The codebase is now simpler and more maintainable with clear, single-purpose components and no backwards compatibility layers.
