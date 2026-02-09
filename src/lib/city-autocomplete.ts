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
let validCities: Set<string> = new Set();
let citySelected = false;

/**
 * Reset module state (useful for testing)
 */
export function resetState(): void {
  debounceTimer = undefined;
  validCities = new Set();
  citySelected = false;
}

/**
 * Initialize city autocomplete functionality
 */
export function initCityAutocomplete(): void {
  const cityInput = document.getElementById("playerCity") as HTMLInputElement;
  const datalist = document.getElementById("cityOptions") as HTMLDataListElement;
  const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;

  if (!cityInput || !datalist) {
    return;
  }

  let lastQuery = "";

  // Listen to input changes
  cityInput.addEventListener("input", () => {
    const query = cityInput.value.trim();

    // Mark as not selected when user types
    citySelected = false;
    cityInput.classList.remove("city-selected");
    updateSubmitButton(submitBtn);

    // Check if the current value matches a valid city
    if (validCities.has(query)) {
      citySelected = true;
      cityInput.classList.add("city-selected");
      updateSubmitButton(submitBtn);
      return;
    }

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

  // Listen for selection from datalist
  cityInput.addEventListener("change", () => {
    const value = cityInput.value.trim();
    if (validCities.has(value)) {
      citySelected = true;
      cityInput.classList.add("city-selected");
      updateSubmitButton(submitBtn);
    }
  });

  // Also check on blur
  cityInput.addEventListener("blur", () => {
    const value = cityInput.value.trim();
    if (validCities.has(value)) {
      citySelected = true;
      cityInput.classList.add("city-selected");
      updateSubmitButton(submitBtn);
    }
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
    const meilisearchUrl = window.GEOCODING_API_URL || "http://localhost:8081";

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

    // Clear existing options and valid cities
    datalist.innerHTML = "";
    validCities.clear();

    // Track unique city names to avoid duplicates
    const uniqueCities = new Set<string>();

    // Add new options from hits
    data.hits.forEach((hit) => {
      const cityName = hit.name;

      // Only add if not already in the list
      if (cityName && !uniqueCities.has(cityName)) {
        uniqueCities.add(cityName);
        validCities.add(cityName);
        const option = document.createElement("option");
        option.value = cityName;
        datalist.appendChild(option);
      }
    });
  } catch (error) {
    console.error("Error fetching cities:", error);
  }
}

/**
 * Update submit button state based on form validity
 */
function updateSubmitButton(submitBtn: HTMLButtonElement | null): void {
  if (!submitBtn) return;

  const nameInput = document.getElementById("playerName") as HTMLInputElement;
  const privacyCheckbox = document.getElementById("agreePrivacy") as HTMLInputElement;

  const nameValid = nameInput && nameInput.value.trim().length >= 2;
  const privacyChecked = privacyCheckbox && privacyCheckbox.checked;

  // Enable button only if all conditions are met
  const allValid = nameValid && privacyChecked && citySelected;
  submitBtn.disabled = !allValid;
}

/**
 * Initialize form validation for the name entry form
 */
export function initFormValidation(): void {
  const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;
  const nameInput = document.getElementById("playerName") as HTMLInputElement;
  const privacyCheckbox = document.getElementById("agreePrivacy") as HTMLInputElement;

  if (!submitBtn || !nameInput || !privacyCheckbox) {
    return;
  }

  // Initially disable the button
  submitBtn.disabled = true;

  // Add event listeners for validation
  nameInput.addEventListener("input", () => updateSubmitButton(submitBtn));
  privacyCheckbox.addEventListener("change", () => updateSubmitButton(submitBtn));
}
