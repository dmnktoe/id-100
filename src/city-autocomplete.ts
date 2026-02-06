/**
 * City autocomplete module using Photon geocoding API
 * Provides autocomplete functionality for city selection
 */

interface PhotonProperties {
  name: string;
  city?: string;
  state?: string;
  country?: string;
  osm_key?: string;
  osm_value?: string;
}

interface PhotonFeature {
  properties: PhotonProperties;
  geometry: {
    coordinates: [number, number];
  };
}

interface PhotonResponse {
  features: PhotonFeature[];
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
 * Fetch cities from Photon API
 */
async function fetchCities(
  query: string,
  datalist: HTMLDataListElement
): Promise<void> {
  try {
    // Get Photon URL from window (set by template)
    const photonUrl = window.NOMINATIM_URL || "http://localhost:8081";

    // Search for cities using Photon API
    const params = new URLSearchParams({
      q: query,
      limit: "10",
      lang: "de",
      osm_tag: "place:city",
    });

    // Also search for towns
    const response = await fetch(`${photonUrl}/api?${params.toString()}`);

    if (!response.ok) {
      console.error("Photon API error:", response.statusText);
      return;
    }

    const data: PhotonResponse = await response.json();

    // Clear existing options
    datalist.innerHTML = "";

    // Track unique city names to avoid duplicates
    const uniqueCities = new Set<string>();

    // Add new options from features
    data.features.forEach((feature) => {
      const cityName = feature.properties.name;

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
