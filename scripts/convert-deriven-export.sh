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

# Create the migration header
cat > "$OUTPUT_FILE" << 'EOF'
-- Migration: 002_insert_initial_deriven.sql
-- Description: Inserts initial derive challenges into the deriven table
-- Date: 2024-02-06
-- Source: Converted from Supabase export (deriven_rows.sql)

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

echo "✓ Conversion complete!"
echo "The migration file has been updated at: $OUTPUT_FILE"
echo "Review the file and restart the Docker containers to apply the migration."
