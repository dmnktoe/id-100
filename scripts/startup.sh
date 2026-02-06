#!/bin/sh
# Startup script for the webapp container
# This runs before the main application starts

set -e

echo "==> Running startup tasks..."

# Path to the deriven conversion script and files
SCRIPT_DIR="/app/scripts"
MIGRATION_DIR="/app/internal/database/migrations"
INPUT_FILE="$MIGRATION_DIR/deriven_rows.sql"
OUTPUT_FILE="$MIGRATION_DIR/002_insert_initial_deriven.sql"

# Check if the conversion script should run
if [ -f "$INPUT_FILE" ]; then
    echo "==> Found deriven_rows.sql, checking if conversion is needed..."
    
    # Check if the input file has actual data (not just placeholder/comments)
    # Count lines that are actual INSERT statements
    ACTUAL_LINES=$(grep -c "^INSERT INTO" "$INPUT_FILE" 2>/dev/null || echo "0")
    
    if [ "$ACTUAL_LINES" -gt 0 ]; then
        echo "==> Converting deriven_rows.sql to migration format..."
        
        # Run the conversion inline (since we're in the container)
        cat > "$OUTPUT_FILE" << 'EOF'
-- Migration: 002_insert_initial_deriven.sql
-- Description: Inserts initial derive challenges into the deriven table
-- Date: 2024-02-06
-- Source: Converted from Supabase export (deriven_rows.sql)
-- Auto-converted on container startup

-- Insert initial deriven data
-- Only insert if the table is empty to avoid duplicates
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM deriven LIMIT 1) THEN
EOF

        # Append the actual INSERT statements from the input file
        # Convert "public"."deriven" to just deriven
        sed 's/"public"\."deriven"/deriven/g' "$INPUT_FILE" >> "$OUTPUT_FILE"

        # Add the closing of the DO block
        cat >> "$OUTPUT_FILE" << 'EOF'
    END IF;
END $$;
EOF

        echo "==> ✓ Conversion complete!"
        echo "    Created: $OUTPUT_FILE"
    else
        echo "==> Skipping conversion: deriven_rows.sql contains no INSERT statements"
        echo "    (Placeholder file detected)"
    fi
else
    echo "==> No deriven_rows.sql found, skipping conversion"
fi

echo "==> Starting application..."
exec "$@"
