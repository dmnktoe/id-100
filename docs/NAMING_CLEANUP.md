# Naming Cleanup: NOMINATIM → GEOCODING_API_URL

This document explains the renaming of variables and references from NOMINATIM to GEOCODING_API_URL throughout the codebase.

## Background

The application originally planned to use Nominatim for geocoding, but the implementation was changed to use **Meilisearch + GeoNames** instead. However, many variable names and references still used "NOMINATIM" which was confusing and misleading.

## What Changed

### Environment Variable

**Old:** `NOMINATIM_URL`  
**New:** `GEOCODING_API_URL`

This better describes what the variable does (provides geocoding/city search) without implying a specific technology.

### File Rename

**Old:** `internal/config/nominatim.go`  
**New:** `internal/config/geocoding.go`

### Function Rename

**Old:** `config.GetNominatimURL()`  
**New:** `config.GetGeocodingURL()`

### Template Variable

**Old:** `{{.NominatimURL}}`  
**New:** `{{.GeocodingURL}}`

### JavaScript Global

**Old:** `window.NOMINATIM_URL`  
**New:** `window.GEOCODING_API_URL`

## Files Updated

1. **`.env.example`** - Environment variable renamed with clarifying comment
2. **`internal/config/geocoding.go`** - File renamed, function renamed, constants updated
3. **`internal/templates/helpers.go`** - Template data key renamed
4. **`src/globals.d.ts`** - TypeScript Window interface property renamed
5. **`src/city-autocomplete.ts`** - Updated to use `window.GEOCODING_API_URL`
6. **`src/city-autocomplete.test.ts`** - Test mocks updated
7. **`web/templates/layout.html`** - JavaScript variable renamed
8. **`docker-compose.yml`** - Environment variable renamed
9. **`README.md`** - Configuration examples updated

## Migration Guide

### For Existing Deployments

If you have an existing `.env` file or environment variables:

**Update your configuration:**
```bash
# In your .env file or environment
GEOCODING_API_URL=http://meilisearch:7700
```

### For Docker Compose

Update your docker-compose.yml:
```yaml
environment:
  GEOCODING_API_URL: http://meilisearch:7700  # Previously NOMINATIM_URL
```

### For Documentation

When referring to the geocoding service in documentation:
- ✅ "Geocoding API" or "Meilisearch geocoding API"
- ✅ "City search API"
- ❌ "Nominatim API" (we don't use Nominatim)

## Why This Matters

1. **Accuracy**: The variable name now reflects the actual implementation (Meilisearch + GeoNames)
2. **Clarity**: New developers won't be confused about which service is being used
3. **Maintainability**: Easier to understand the codebase without misleading names
4. **Documentation**: README and code comments now align with reality

## Technical Details

### What is Meilisearch?

Meilisearch is a fast, typo-tolerant search engine. In this application, it indexes German city data from GeoNames.org and provides instant autocomplete suggestions.

### What is GeoNames?

GeoNames is a free geographical database containing over 25 million place names. We use it as the data source for German cities.

### API Endpoint

The `GEOCODING_API_URL` points to the Meilisearch instance:
- **Development:** `http://localhost:8081` (Meilisearch running in Docker)
- **Production:** Your deployed Meilisearch instance URL

### API Usage

The JavaScript code makes POST requests to:
```
${GEOCODING_API_URL}/indexes/cities/search
```

With body:
```json
{
  "q": "Berlin",
  "limit": 10,
  "attributesToRetrieve": ["name"]
}
```

## Testing

All tests have been updated and pass:
- ✅ 30 TypeScript tests (including 5 city-autocomplete tests)
- ✅ All Go tests
- ✅ Frontend builds successfully
- ✅ Backend builds successfully
- ✅ Docker Compose validates

## Future Considerations

If we ever need to support multiple geocoding providers, we could extend this to:
```env
GEOCODING_PROVIDER=meilisearch  # or 'nominatim', 'google', etc.
GEOCODING_API_URL=http://meilisearch:7700
```

But for now, the single `GEOCODING_API_URL` variable is sufficient.
