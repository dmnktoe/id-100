#!/bin/sh
# Import German cities from GeoNames to Meilisearch
# This script processes the GeoNames data and imports only cities/towns

set -e

MEILI_URL="${MEILI_URL:-http://meilisearch:7700}"
INDEX_NAME="cities"
GEONAMES_FILE="/tmp/DE.txt"

echo "Processing GeoNames data..."

# Create JSON from GeoNames data
# GeoNames format: geonameid, name, asciiname, alternatenames, latitude, longitude, feature_class, feature_code, ...
# We only want places with feature_class=P (populated place) and specific feature_codes

# Filter for cities and convert to JSON
awk -F'\t' '
BEGIN { 
    print "["
    first = 1
}
$7 == "P" && ($8 ~ /^(PPL|PPLA|PPLA2|PPLA3|PPLA4|PPLC)$/) {
    if (!first) print ","
    first = 0
    
    # Extract fields
    id = $1
    name = $2
    lat = $5
    lon = $6
    feature = $8
    population = $15
    
    # Determine type
    type = "city"
    if (feature == "PPLC") type = "capital"
    else if (feature == "PPLA") type = "major_city"
    else if (feature == "PPLA2") type = "town"
    
    # Print JSON object
    printf "  {\"id\":\"%s\",\"name\":\"%s\",\"lat\":%s,\"lon\":%s,\"type\":\"%s\",\"population\":%s}", 
           id, name, lat, lon, type, (population == "" ? "0" : population)
}
END { 
    print ""
    print "]"
}
' "$GEONAMES_FILE" > /tmp/cities.json

echo "Created JSON with $(cat /tmp/cities.json | grep -c '"id"') cities"

# Wait for Meilisearch to be ready
echo "Waiting for Meilisearch..."
until curl -sf "$MEILI_URL/health" > /dev/null 2>&1; do
    sleep 1
done

echo "Creating index and configuring..."

# Create index with settings
curl -X POST "$MEILI_URL/indexes" \
  -H "Content-Type: application/json" \
  --data-binary "{
    \"uid\": \"$INDEX_NAME\",
    \"primaryKey\": \"id\"
  }" || true

# Configure searchable attributes
curl -X PATCH "$MEILI_URL/indexes/$INDEX_NAME/settings/searchable-attributes" \
  -H "Content-Type: application/json" \
  --data-binary '["name"]'

# Configure ranking rules for better autocomplete
curl -X PATCH "$MEILI_URL/indexes/$INDEX_NAME/settings/ranking-rules" \
  -H "Content-Type: application/json" \
  --data-binary '["words","typo","proximity","attribute","sort","exactness","population:desc"]'

# Configure typo tolerance
curl -X PATCH "$MEILI_URL/indexes/$INDEX_NAME/settings/typo-tolerance" \
  -H "Content-Type: application/json" \
  --data-binary '{
    "enabled": true,
    "minWordSizeForTypos": {
      "oneTypo": 4,
      "twoTypos": 8
    }
  }'

echo "Importing cities to Meilisearch..."

# Import documents
curl -X POST "$MEILI_URL/indexes/$INDEX_NAME/documents" \
  -H "Content-Type: application/json" \
  --data-binary @/tmp/cities.json

echo "Import complete! Index '$INDEX_NAME' ready for queries."
echo "Test query: curl '$MEILI_URL/indexes/$INDEX_NAME/search?q=Berlin&limit=5'"

# Cleanup
rm -f /tmp/cities.json /tmp/DE.txt /tmp/DE.zip

exit 0
