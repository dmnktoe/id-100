/**
 * City autocomplete module using Meilisearch with GeoNames data
 * Provides autocomplete functionality for city selection
 */

interface MeilisearchHit {
  id: string;
  name: string;
  lat: number;
  lon: number;
  type: string;
  population: number;
}

interface MeilisearchResponse {
  hits: MeilisearchHit[];
  query: string;
  processingTimeMs: number;
  limit: number;
  offset: number;
  estimatedTotalHits: number;
}

let debounceTimer: number | undefined;

/**
 * Initialize city autocomplete functionality
 */
export function initCityAutocomplete(): void {
  const cityInput = document.getElementById("playerCity") as HTMLInputElement;
  const datalist = document.getElementById("cityOptions") as HTMLDataListElement;

  if (!cityInput || !datalist) {
    return;
  }

  let lastQuery = "";

  // Listen to input changes
  cityInput.addEventListener("input", () => {
    const query = cityInput.value.trim();

    // Don't search if query is too short or same as last query
    if (query.length < 2 || query === lastQuery) {
      return;
    }

    lastQuery = query;

    // Clear existing timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    // Debounce the API call
    debounceTimer = window.setTimeout(() => {
      fetchCities(query, datalist);
    }, 300);
  });
}

/**
 * Fetch cities from Meilisearch API
 */
async function fetchCities(
  query: string,
  datalist: HTMLDataListElement
): Promise<void> {
  try {
    // Get Meilisearch URL from window (set by template)
    const meilisearchUrl = window.NOMINATIM_URL || "http://localhost:8081";

    // Search for cities using Meilisearch API
    const response = await fetch(
      `${meilisearchUrl}/indexes/cities/search`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          q: query,
          limit: 10,
          attributesToRetrieve: ["name"],
        }),
      }
    );

    if (!response.ok) {
      console.error("Meilisearch API error:", response.statusText);
      return;
    }

    const data: MeilisearchResponse = await response.json();

    // Clear existing options
    datalist.innerHTML = "";

    // Track unique city names to avoid duplicates
    const uniqueCities = new Set<string>();

    // Add new options from hits
    data.hits.forEach((hit) => {
      const cityName = hit.name;

      // Only add if not already in the list
      if (cityName && !uniqueCities.has(cityName)) {
        uniqueCities.add(cityName);
        const option = document.createElement("option");
        option.value = cityName;
        datalist.appendChild(option);
      }
    });
  } catch (error) {
    console.error("Error fetching cities:", error);
  }
}
