# Recent Fixes - February 2026

This document summarizes the fixes applied in response to user feedback on the PR.

## Issue 1: Remove Supabase References

**Problem**: The app was still referencing Supabase storage paths even though we switched to MinIO.

**Solution**:
- Updated `internal/utils/utils.go::EnsureFullImageURL()` to construct MinIO URLs
- Updated `internal/utils/s3.go::extractFileNameFromURL()` to parse MinIO URL format
- Removed `SUPABASE_URL` from `.env.example`
- Added MinIO defaults (minioadmin/minioadmin) to `.env.example`

**URL Format Change**:
- Before: `https://xxx.supabase.co/storage/v1/object/public/id100-images/file.webp`
- After: `http://minio:9000/id100-images/file.webp`

## Issue 2: Fix Deriven Conversion (100 Records)

**Problem**: Only 1 deriven was being inserted instead of all 100 from the export file.

**Root Cause**: The DO block in the conversion script had a conditional `IF NOT EXISTS` that prevented re-importing data.

**Solution**:
- Removed the DO block and conditional logic
- Added `DELETE FROM deriven` to clear existing data
- Added `ALTER SEQUENCE deriven_id_seq RESTART WITH 1` to reset IDs
- Updated both `scripts/convert-deriven-export.sh` and `scripts/startup.sh`
- Startup script now displays count of INSERT statements found

**Before** (only inserted if table was empty):
```sql
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM deriven LIMIT 1) THEN
        INSERT INTO deriven ...
    END IF;
END $$;
```

**After** (always re-imports):
```sql
DELETE FROM deriven WHERE id IS NOT NULL;
ALTER SEQUENCE deriven_id_seq RESTART WITH 1;
INSERT INTO deriven ...
```

## Issue 3: City Filter Layout

**Problem**: Filter took up too much vertical space on desktop but looked good on mobile.

**Solution**:
- Created `.page-header` flex container
- Desktop: Title on left, filter on right (same height)
- Mobile: Stack vertically (responsive design)
- Reduced filter padding for more compact appearance

**CSS**:
```css
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--gap-lg);
}

@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
```

**Visual Layout**:

Desktop:
```
┌──────────────────────────────────────────────┐
│ [Index: 1-100]        [📍 Stadt filtern ▼] │
└──────────────────────────────────────────────┘
```

Mobile:
```
┌──────────────────────┐
│ [Index: 1-100]       │
│ [📍 Stadt filtern ▼] │
└──────────────────────┘
```

## Issue 4: Datalist Dropdown Styling

**Problem**: City autocomplete dropdown needed better styling to match app's corporate identity.

**Solution**:
- Added explicit datalist styling with app's CSS custom properties
- White background with borders and box shadows
- Hover effects on dropdown options
- Width automatically matches input field

**CSS Added**:
```css
datalist {
  position: absolute;
  background: var(--white);
  border: var(--border-medium);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-md);
  max-height: 250px;
  overflow-y: auto;
}

datalist option {
  padding: var(--pad-sm) var(--pad-md);
  cursor: pointer;
  transition: background var(--transition-fast);
}

datalist option:hover {
  background: var(--gray-50);
}
```

## Issue 5: Makefile Modernization

**Problem**: Makefile lacked Docker Compose commands for the new infrastructure.

**Solution**: Added comprehensive Docker Compose targets:

```makefile
make docker-build    # Build Docker images
make docker-up       # Start all services
make docker-down     # Stop all services
make docker-restart  # Restart services
make docker-logs     # View logs (follow mode)
make docker-clean    # Stop and remove volumes
make rebuild         # Full rebuild (down + build + up)
make dev             # Start services and follow logs
```

Also added `make help` to show all available commands.

## Testing

All changes validated:
- ✅ All 34 TypeScript tests passing
- ✅ Frontend builds successfully (7.2kb)
- ✅ Backend compiles without errors
- ✅ CSS validated
- ✅ Docker Compose configuration valid

## Migration Guide

For existing installations:

1. **Update environment variables**:
   ```bash
   # Remove from .env:
   SUPABASE_URL=...
   
   # Ensure these are set:
   S3_ENDPOINT=http://minio:9000
   S3_BUCKET=id100-images
   S3_ACCESS_KEY=minioadmin
   S3_SECRET_KEY=minioadmin
   ```

2. **Reimport deriven data** (if needed):
   ```bash
   # Replace the placeholder file
   cp /path/to/your-export.sql internal/database/migrations/deriven_rows.sql
   
   # Restart Docker (conversion happens automatically)
   docker-compose down -v
   docker-compose up -d --build
   ```

3. **Verify images load**:
   - Images should now load from MinIO
   - Check browser network tab for `http://minio:9000/id100-images/` URLs

## Files Changed

- `.env.example` - Removed Supabase, added MinIO defaults
- `Makefile` - Added Docker Compose commands
- `internal/utils/utils.go` - MinIO URL construction
- `internal/utils/s3.go` - MinIO URL parsing
- `scripts/convert-deriven-export.sh` - Fixed import logic
- `scripts/startup.sh` - Fixed import logic
- `web/templates/app/deriven.html` - Page header layout
- `web/static/style.css` - Layout and datalist styling

Commit: 25c212d
