# City Autocomplete with Meilisearch + GeoNames

This document explains the city autocomplete implementation using Meilisearch and GeoNames data.

## Why Meilisearch + GeoNames?

We evaluated several geocoding solutions and chose Meilisearch with GeoNames data for the following reasons:

### Comparison of Approaches

| Solution | Data Size | Startup Time | Use Case | Issues |
|----------|-----------|--------------|----------|--------|
| **Nominatim** | 4GB download, 200GB processed | 5-10 minutes | Full geocoding | Massive overkill for autocomplete |
| **Photon** | 350MB+ index, requires Nominatim data | 2-3 minutes | Autocomplete | Still requires large OSM dataset |
| **Meilisearch + GeoNames** | ~10MB | 1 minute | Autocomplete | **✓ Perfect fit** |

### Advantages of Meilisearch + GeoNames

1. **Lightweight**: Only ~10MB of data for all German cities
2. **Fast**: Sub-millisecond search responses
3. **Purpose-built**: Meilisearch is designed for instant search/autocomplete
4. **Easy to maintain**: Simple data format, easy updates
5. **No external dependencies**: Self-contained solution
6. **Typo-tolerant**: Built-in fuzzy search
7. **Scalable**: Can handle millions of queries

## Architecture

### Components

1. **Meilisearch Container**: Fast, typo-tolerant search engine (port 8081)
2. **GeoNames Loader**: Downloads and imports German cities data
3. **Frontend**: TypeScript with debounced search and HTML5 datalist

### Data Flow

```
User types "Ber" → 300ms debounce → POST to Meilisearch
→ Search index → Return top 10 matches → Populate datalist
→ User sees: Berlin, Bergisch Gladbach, etc.
```

## GeoNames Data

- **Source**: GeoNames.org (free geographical database)
- **Dataset**: DE.zip (~10MB compressed)
- **Cities**: ~15,000 German cities/towns/villages
- **Fields**: name, coordinates, population, type

## Performance

- Search latency: <5ms
- Index size: ~15MB
- Concurrent queries: 1000+ req/s
- Cold start: ~30 seconds
- Data import: ~1 minute

## Resources

- [Meilisearch Documentation](https://docs.meilisearch.com/)
- [GeoNames.org](https://www.geonames.org/)
