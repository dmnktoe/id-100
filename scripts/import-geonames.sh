#!/bin/sh
# Import German cities from GeoNames to Meilisearch
set -e

MEILI_URL="${MEILI_URL:-http://meilisearch:7700}"
# Den Master Key aus der Umgebungsvariable nehmen
AUTH_HEADER="Authorization: Bearer ${MEILI_MASTER_KEY}"
INDEX_NAME="cities"
GEONAMES_FILE="/tmp/DE.txt"

echo "Processing GeoNames data..."

# Filter for cities and convert to JSON (dein bewährtes awk-Script)
awk -F'\t' '
BEGIN { print "["; first = 1 }
$7 == "P" && ($8 ~ /^(PPL|PPLA|PPLA2|PPLA3|PPLA4|PPLC)$/) {
    if (!first) print ","
    first = 0
    id = $1; name = $2; lat = $5; lon = $6; feature = $8; population = $15
    type = "city"
    if (feature == "PPLC") type = "capital"
    else if (feature == "PPLA") type = "major_city"
    else if (feature == "PPLA2") type = "town"
    printf "  {\"id\":\"%s\",\"name\":\"%s\",\"lat\":%s,\"lon\":%s,\"type\":\"%s\",\"population\":%s}", 
           id, name, lat, lon, type, (population == "" ? "0" : population)
}
END { print "\n]" }
' "$GEONAMES_FILE" > /tmp/cities.json

echo "Created JSON with $(cat /tmp/cities.json | grep -c '"id"') cities"

echo "Waiting for Meilisearch..."
until curl -sf -H "$AUTH_HEADER" "$MEILI_URL/health" > /dev/null 2>&1; do
    sleep 1
done

echo "Creating index and configuring..."

# Create index
curl -X POST "$MEILI_URL/indexes" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  --data-binary "{\"uid\": \"$INDEX_NAME\", \"primaryKey\": \"id\"}" || true

# Searchable attributes
curl -X PATCH "$MEILI_URL/indexes/$INDEX_NAME/settings/searchable-attributes" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  --data-binary '["name"]'

# Ranking rules
curl -X PATCH "$MEILI_URL/indexes/$INDEX_NAME/settings/ranking-rules" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  --data-binary '["words","typo","proximity","attribute","sort","exactness","population:desc"]'

# Typo tolerance
curl -X PATCH "$MEILI_URL/indexes/$INDEX_NAME/settings/typo-tolerance" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  --data-binary '{"enabled": true, "minWordSizeForTypos": {"oneTypo": 4, "twoTypos": 8}}'

echo "Importing cities to Meilisearch..."

# Import documents
curl -X POST "$MEILI_URL/indexes/$INDEX_NAME/documents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  --data-binary @/tmp/cities.json

echo "Import complete!"
# Test query am Ende mit Header
echo "Test query result:"
curl -s -H "$AUTH_HEADER" "$MEILI_URL/indexes/$INDEX_NAME/search?q=Berlin&limit=1"

# Cleanup
rm -f /tmp/cities.json /tmp/DE.txt /tmp/DE.zip