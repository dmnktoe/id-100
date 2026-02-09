#!/bin/sh
# Script to convert Supabase deriven_rows.sql to migration format
# Converts: internal/database/migrations/deriven_rows.sql
# To: internal/database/migrations/002_insert_initial_deriven.sql

set -e

INPUT_FILE="internal/database/migrations/deriven_rows.sql"
OUTPUT_FILE="internal/database/migrations/002_insert_initial_deriven.sql"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found"
    echo "Please ensure your Supabase export is at: $INPUT_FILE"
    exit 1
fi

echo "Converting $INPUT_FILE to $OUTPUT_FILE..."

# Create the migration file with proper SQL structure
cat > "$OUTPUT_FILE" << 'EOF'
-- Migration: 002_insert_initial_deriven.sql
-- Description: Inserts initial derive challenges into the deriven table
-- Date: 2024-02-06
-- Source: Converted from Supabase export (deriven_rows.sql)

-- Delete existing data if any (for clean re-import)
DELETE FROM deriven WHERE id IS NOT NULL;

-- Reset sequence to start from 1
ALTER SEQUENCE deriven_id_seq RESTART WITH 1;

EOF

# Append the actual INSERT statements from the input file
# Convert "public"."deriven" to just deriven and remove quotes
sed -e 's/"public"\."deriven"/deriven/g' -e 's/"public"\.deriven/deriven/g' "$INPUT_FILE" >> "$OUTPUT_FILE"

echo ""
echo "✓ Conversion complete!"
echo "The migration file has been created at: $OUTPUT_FILE"
echo "Restart Docker containers to apply the migration:"
echo "  docker-compose down && docker-compose up -d --build"
