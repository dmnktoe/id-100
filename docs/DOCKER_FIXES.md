# Docker Runtime Fixes

This document explains the fixes applied to resolve Docker runtime issues.

## Issues Fixed

### 1. Template Parsing Failure

**Problem:**
```
id100-webapp | 2026/02/06 12:57:08 failed to parse templates []: html/template: no files named in call to ParseFiles
```

**Root Cause:**
The template loading code in `internal/templates/templates.go` was failing silently when no template files were found. The glob pattern `web/templates/*.html` was correct, but there was no validation that files were actually found before attempting to parse them.

**Solution:**
Added validation and better error messages:
- Check if `len(files) == 0` and fail with a clear error message
- Log the number of template files found
- Log successful template loading
- Improved error messages to help debug path issues

**Code Changes:**
```go
// Check if we found any template files
if len(files) == 0 {
    log.Fatalf("no template files found. Current directory might be wrong. Looking for: web/templates/*.html")
}

log.Printf("Found %d template files to load", len(files))
// ... parse templates ...
log.Printf("Successfully loaded templates")
```

**Why it works:**
The Dockerfile correctly copies templates to `/app/web`, and the working directory is `/app`, so the relative path `web/templates/*.html` resolves correctly. The issue was just lack of validation and error reporting.

### 2. GeoNames Loader Script Not Found

**Problem:**
```
id100-geonames-loader  | /bin/sh: /scripts/import-geonames.sh: not found
```

**Root Cause:**
The `curlimages/curl` image is a minimal image that:
1. Doesn't have `awk` installed (required by the script)
2. Doesn't have `unzip` installed (required to extract DE.zip)
3. Has a non-standard filesystem layout
4. Missing other standard shell utilities

**Solution:**
Switched from `curlimages/curl:latest` to `alpine:latest` and:
1. Install required tools: `curl`, `unzip`, `awk`
2. Use explicit `sh /scripts/import-geonames.sh` instead of direct execution
3. Keep the script volume mount: `./scripts:/scripts:ro`

**Code Changes:**
```yaml
geonames-loader:
  image: alpine:latest  # Changed from curlimages/curl:latest
  # ...
  command:
    - -c
    - |
      echo "Installing required tools..."
      apk add --no-cache curl unzip awk  # Install dependencies
      echo "Downloading GeoNames data for Germany..."
      cd /tmp
      curl -O https://download.geonames.org/export/dump/DE.zip
      unzip -o DE.zip
      echo "Converting to JSON and importing to Meilisearch..."
      sh /scripts/import-geonames.sh  # Explicit sh execution
      echo "GeoNames import complete"
```

**Why it works:**
- Alpine has a standard Linux filesystem with all standard tools
- Installing packages with `apk` ensures all dependencies are met
- Explicit `sh` execution works around any permission or path issues
- The script volume is correctly mounted to `/scripts`

## Verification

### Template Loading
After the fix, you should see:
```
id100-webapp | Found 25 template files to load
id100-webapp | Successfully loaded templates
id100-webapp | Starting server on port 8080
```

### GeoNames Loader
After the fix, you should see:
```
id100-geonames-loader | Installing required tools...
id100-geonames-loader | Downloading GeoNames data for Germany...
id100-geonames-loader | Converting to JSON and importing to Meilisearch...
id100-geonames-loader | Processing GeoNames data...
id100-geonames-loader | Created JSON with 15247 cities
id100-geonames-loader | Import complete!
```

## Testing

To test the fixes:

```bash
# Clean up any existing containers/volumes
docker-compose down -v

# Build and start the stack
docker-compose up --build

# Check logs for successful startup
docker-compose logs webapp
docker-compose logs geonames-loader

# Verify webapp is responding
curl http://localhost:8080

# Verify Meilisearch has cities
curl http://localhost:8081/indexes/cities/stats
```

## Architecture Notes

### Docker Working Directory
- Dockerfile sets `WORKDIR /app`
- Templates copied to `/app/web`
- Binary at `/app/id-100`
- Relative paths in Go code work from `/app`

### Volume Mounts
Development mode mounts:
- `./web:/app/web` - Template hot-reload
- `./scripts:/scripts:ro` - Read-only script access

### Service Dependencies
```
webapp depends_on:
  - db (healthy)
  - minio (healthy)
  - meilisearch (healthy)

geonames-loader depends_on:
  - meilisearch (healthy)
```

## Future Improvements

1. **Cache GeoNames Data**: Store processed JSON in a volume to avoid re-downloading
2. **Healthcheck for GeoNames**: Add a completion marker file that webapp can check
3. **Template Reload**: Use file watching in development for automatic template reload
4. **Better Logging**: Add structured logging with log levels
