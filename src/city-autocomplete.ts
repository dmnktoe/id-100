/**
 * City autocomplete module using Nominatim API
 * Provides autocomplete functionality for city selection
 */

interface NominatimResult {
  place_id: number;
  display_name: string;
  name: string;
  address: {
    city?: string;
    town?: string;
    village?: string;
    municipality?: string;
  };
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
 * Fetch cities from Nominatim API
 */
async function fetchCities(
  query: string,
  datalist: HTMLDataListElement
): Promise<void> {
  try {
    // Get Nominatim URL from environment or use default
    const nominatimUrl = (window as any).NOMINATIM_URL || "http://localhost:8081";

    // Search for cities in Germany
    const params = new URLSearchParams({
      q: query,
      format: "json",
      limit: "10",
      countrycodes: "de",
      addressdetails: "1",
      "accept-language": "de",
    });

    // Filter to only cities/towns/villages
    params.append("featuretype", "city");

    const response = await fetch(
      `${nominatimUrl}/search?${params.toString()}`
    );

    if (!response.ok) {
      console.error("Nominatim API error:", response.statusText);
      return;
    }

    const results: NominatimResult[] = await response.json();

    // Clear existing options
    datalist.innerHTML = "";

    // Add new options
    results.forEach((result) => {
      const option = document.createElement("option");
      
      // Extract city name from address
      const cityName =
        result.address.city ||
        result.address.town ||
        result.address.village ||
        result.address.municipality ||
        result.name;

      option.value = cityName;
      datalist.appendChild(option);
    });
  } catch (error) {
    console.error("Error fetching cities:", error);
  }
}
