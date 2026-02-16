/**
 * City autocomplete module using Meilisearch SDK with custom dropdown
 * Provides autocomplete functionality for city selection with styled dropdown
 * Includes Zod schema validation for form data
 */

import { MeiliSearch } from 'meilisearch';
import { z } from 'zod';

interface CityHit {
  id: string;
  name: string;
  lat: number;
  lon: number;
  type: string;
  population: number;
}

export interface ValidationResult {
  success: boolean;
  errors: string[];
}

// Zod schema for name form validation
const nameFormSchema = z.object({
  player_name: z.string()
    .min(2, "Name muss mindestens 2 Zeichen lang sein")
    .max(100, "Name darf maximal 100 Zeichen lang sein"),
  player_city: z.string()
    .min(2, "Stadt muss mindestens 2 Zeichen lang sein")
    .max(100, "Stadt darf maximal 100 Zeichen lang sein"),
  agree_privacy: z.literal(true, {
    errorMap: () => ({ 
      message: "Datenschutzerklärung muss akzeptiert werden" 
    })
  })
});

/**
 * Validate form data using Zod schema
 */
export function validateForm(data: unknown): ValidationResult {
  const result = nameFormSchema.safeParse(data);
  
  if (result.success) {
    return {
      success: true,
      errors: []
    };
  }
  
  return {
    success: false,
    errors: result.error.errors.map(err => err.message)
  };
}

let debounceTimer: number | undefined;
let validCities: Set<string> = new Set();
let citySelected = false;
let client: MeiliSearch | null = null;
let selectedIndex = -1;
let currentResults: string[] = [];

/**
 * Reset module state (useful for testing)
 */
export function resetState(): void {
  debounceTimer = undefined;
  validCities = new Set();
  citySelected = false;
  selectedIndex = -1;
  currentResults = [];
}

/**
 * Initialize Meilisearch client
 */
function initMeilisearchClient(): MeiliSearch {
  if (!client) {
    const meilisearchUrl = window.GEOCODING_API_URL || "http://localhost:8081";
    const apiKey = window.MEILI_SEARCH_KEY || "";
    client = new MeiliSearch({
      host: meilisearchUrl,
      apiKey: apiKey,
    });
  }
  return client;
}

/**
 * Create custom dropdown element
 */
function createDropdown(): HTMLDivElement {
  const dropdown = document.createElement("div");
  dropdown.id = "city-dropdown";
  dropdown.className = "city-dropdown";
  dropdown.style.display = "none";
  return dropdown;
}

/**
 * Position dropdown below input
 */
function positionDropdown(input: HTMLInputElement, dropdown: HTMLDivElement): void {
  const rect = input.getBoundingClientRect();
  dropdown.style.position = "absolute";
  dropdown.style.top = `${rect.bottom + window.scrollY}px`;
  dropdown.style.left = `${rect.left + window.scrollX}px`;
  dropdown.style.width = `${rect.width}px`;
}

/**
 * Show dropdown with results
 */
function showDropdown(
  input: HTMLInputElement,
  dropdown: HTMLDivElement,
  cities: string[]
): void {
  if (cities.length === 0) {
    dropdown.style.display = "none";
    return;
  }

  currentResults = cities;
  selectedIndex = -1;

  dropdown.innerHTML = cities
    .map(
      (city, index) =>
        `<div class="city-dropdown-item" data-index="${index}">${city}</div>`
    )
    .join("");

  positionDropdown(input, dropdown);
  dropdown.style.display = "block";

  // Add click handlers to dropdown items
  dropdown.querySelectorAll(".city-dropdown-item").forEach((item) => {
    item.addEventListener("click", () => {
      const cityName = item.textContent || "";
      selectCity(input, dropdown, cityName);
    });
  });
}

/**
 * Hide dropdown
 */
function hideDropdown(dropdown: HTMLDivElement): void {
  dropdown.style.display = "none";
  selectedIndex = -1;
}

/**
 * Select a city from dropdown
 */
function selectCity(
  input: HTMLInputElement,
  dropdown: HTMLDivElement,
  cityName: string
): void {
  input.value = cityName;
  citySelected = true;
  input.classList.add("city-selected");
  hideDropdown(dropdown);

  const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;
  updateSubmitButton(submitBtn);
}

/**
 * Handle keyboard navigation
 */
function handleKeyboard(
  e: KeyboardEvent,
  input: HTMLInputElement,
  dropdown: HTMLDivElement
): void {
  const items = dropdown.querySelectorAll(".city-dropdown-item");

  if (e.key === "ArrowDown") {
    e.preventDefault();
    selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
    updateHighlight(items);
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    selectedIndex = Math.max(selectedIndex - 1, -1);
    updateHighlight(items);
  } else if (e.key === "Enter") {
    e.preventDefault();
    if (selectedIndex >= 0 && selectedIndex < currentResults.length) {
      selectCity(input, dropdown, currentResults[selectedIndex]);
    }
  } else if (e.key === "Escape") {
    hideDropdown(dropdown);
  }
}

/**
 * Update highlighted item in dropdown
 */
function updateHighlight(items: NodeListOf<Element>): void {
  items.forEach((item, index) => {
    if (index === selectedIndex) {
      item.classList.add("highlighted");
      item.scrollIntoView({ block: "nearest" });
    } else {
      item.classList.remove("highlighted");
    }
  });
}

/**
 * Initialize city autocomplete functionality
 */
export function initCityAutocomplete(): void {
  const cityInput = document.getElementById("playerCity") as HTMLInputElement;
  const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;

  if (!cityInput) {
    return;
  }

  // Set initial button state to disabled
  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.classList.add("disabled");
  }

  // Initialize Meilisearch client
  const meiliClient = initMeilisearchClient();

  // Create custom dropdown and insert after the input
  const dropdown = createDropdown();
  if (cityInput.parentNode) {
    // Insert dropdown right after the input element
    cityInput.parentNode.insertBefore(dropdown, cityInput.nextSibling);
  }

  let lastQuery = "";

  // Listen to input changes
  cityInput.addEventListener("input", () => {
    const query = cityInput.value.trim();

    // Mark as not selected when user types
    citySelected = false;
    cityInput.classList.remove("city-selected");
    updateSubmitButton(submitBtn);
    updateStatusIndicators();

    // Check if the current value matches a valid city
    if (validCities.has(query)) {
      citySelected = true;
      cityInput.classList.add("city-selected");
      updateSubmitButton(submitBtn);
      updateStatusIndicators();
      hideDropdown(dropdown);
      return;
    }

    // Don't search if query is too short or same as last query
    if (query.length < 2) {
      hideDropdown(dropdown);
      return;
    }

    if (query === lastQuery) {
      return;
    }

    lastQuery = query;

    // Clear existing timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    // Debounce the API call
    debounceTimer = window.setTimeout(() => {
      searchCities(query, cityInput, dropdown, meiliClient);
    }, 300);
  });

  // Handle keyboard navigation
  cityInput.addEventListener("keydown", (e) => {
    if (dropdown.style.display === "block") {
      handleKeyboard(e, cityInput, dropdown);
    }
  });

  // Hide dropdown when clicking outside
  document.addEventListener("click", (e) => {
    if (!cityInput.contains(e.target as Node) && !dropdown.contains(e.target as Node)) {
      hideDropdown(dropdown);
    }
  });

  // Listen for selection
  cityInput.addEventListener("change", () => {
    const value = cityInput.value.trim();
    if (validCities.has(value)) {
      citySelected = true;
      cityInput.classList.add("city-selected");
      updateSubmitButton(submitBtn);
      updateStatusIndicators();
    }
  });

  // Also check on blur
  cityInput.addEventListener("blur", () => {
    // Delay to allow click on dropdown item
    setTimeout(() => {
      const value = cityInput.value.trim();
      if (validCities.has(value)) {
        citySelected = true;
        cityInput.classList.add("city-selected");
        updateSubmitButton(submitBtn);
        updateStatusIndicators();
      } else {
        hideDropdown(dropdown);
      }
    }, 200);
  });
}

/**
 * Search cities using Meilisearch SDK
 */
async function searchCities(
  query: string,
  input: HTMLInputElement,
  dropdown: HTMLDivElement,
  meiliClient: MeiliSearch
): Promise<void> {
  try {
    // Search using Meilisearch SDK
    const searchResults = await meiliClient.index("cities").search<CityHit>(query, {
      limit: 10,
      attributesToRetrieve: ["name"],
    });

    // Clear valid cities
    validCities.clear();

    // Extract unique city names
    const uniqueCities = new Set<string>();
    const cityNames: string[] = [];

    searchResults.hits.forEach((hit) => {
      const cityName = hit.name;
      if (cityName && !uniqueCities.has(cityName)) {
        uniqueCities.add(cityName);
        validCities.add(cityName);
        cityNames.push(cityName);
      }
    });

    // Show dropdown with results
    showDropdown(input, dropdown, cityNames);
  } catch (error) {
    console.error("[CityAutocomplete] Error fetching cities:", error);
    hideDropdown(dropdown);
  }
}

/**
 * Initialize form validation for name entry page
 */
export function initFormValidation(): void {
  const nameInput = document.getElementById("playerName") as HTMLInputElement;
  const privacyCheckbox = document.getElementById("privacyCheckbox") as HTMLInputElement;
  const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;
  const form = document.getElementById("nameForm") as HTMLFormElement;

  if (!nameInput || !privacyCheckbox || !submitBtn) {
    return;
  }

  const updateButton = () => {
    const nameValid = nameInput.value.trim().length >= 2;
    const privacyAccepted = privacyCheckbox.checked;
    const cityValid = citySelected;

    const allValid = nameValid && privacyAccepted && cityValid;
    submitBtn.disabled = !allValid;
    
    // Update button appearance
    if (allValid) {
      submitBtn.classList.remove("disabled");
    } else {
      submitBtn.classList.add("disabled");
    }
    
    updateStatusIndicators();
  };

  nameInput.addEventListener("input", updateButton);
  privacyCheckbox.addEventListener("change", updateButton);

  // Set initial disabled state
  submitBtn.disabled = true;
  submitBtn.classList.add("disabled");
  
  // Prevent form submission if not valid
  if (form) {
    form.addEventListener("submit", (e) => {
      const nameValid = nameInput.value.trim().length >= 2;
      const privacyAccepted = privacyCheckbox.checked;
      const cityValid = citySelected;
      const allValid = nameValid && privacyAccepted && cityValid;
      
      if (allValid) {
        return true;
      }
      
      e.preventDefault();
      alert("Bitte fülle alle Felder korrekt aus und wähle eine Stadt aus der Liste!");
      return false;
    });
  }
  
  // Initialize status indicators
  updateStatusIndicators();
}

/**
 * Update submit button state
 */
function updateSubmitButton(submitBtn: HTMLButtonElement | null): void {
  if (!submitBtn) return;

  const nameInput = document.getElementById("playerName") as HTMLInputElement;
  const privacyCheckbox = document.getElementById("privacyCheckbox") as HTMLInputElement;

  if (!nameInput || !privacyCheckbox) return;

  const nameValid = nameInput.value.trim().length >= 2;
  const privacyAccepted = privacyCheckbox.checked;
  
  const allValid = nameValid && privacyAccepted && citySelected;
  submitBtn.disabled = !allValid;
  
  // Update button appearance
  if (allValid) {
    submitBtn.classList.remove("disabled");
  } else {
    submitBtn.classList.add("disabled");
  }
}

/**
 * Update status indicators in the form
 */
function updateStatusIndicators(): void {
  const nameInput = document.getElementById("playerName") as HTMLInputElement;
  const privacyCheckbox = document.getElementById("privacyCheckbox") as HTMLInputElement;
  const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;
  
  const statusName = document.getElementById("statusName");
  const statusCity = document.getElementById("statusCity");
  const statusPrivacy = document.getElementById("statusPrivacy");
  const statusButton = document.getElementById("statusButton");
  
  if (!nameInput || !privacyCheckbox || !submitBtn) return;
  
  const nameValid = nameInput.value.trim().length >= 2;
  const privacyAccepted = privacyCheckbox.checked;
  
  // Update name status
  if (statusName) {
    if (nameValid) {
      statusName.innerHTML = '✅ Name: Gültig';
      statusName.style.color = 'green';
    } else {
      statusName.innerHTML = '❌ Name: Nicht ausgefüllt';
      statusName.style.color = 'red';
    }
  }
  
  // Update city status
  if (statusCity) {
    if (citySelected) {
      statusCity.innerHTML = '✅ Stadt: Ausgewählt';
      statusCity.style.color = 'green';
    } else {
      statusCity.innerHTML = '❌ Stadt: Nicht ausgewählt';
      statusCity.style.color = 'red';
    }
  }
  
  // Update privacy status
  if (statusPrivacy) {
    if (privacyAccepted) {
      statusPrivacy.innerHTML = '✅ Datenschutz: Akzeptiert';
      statusPrivacy.style.color = 'green';
    } else {
      statusPrivacy.innerHTML = '❌ Datenschutz: Nicht akzeptiert';
      statusPrivacy.style.color = 'red';
    }
  }
  
  // Update button status
  if (statusButton) {
    if (!submitBtn.disabled) {
      statusButton.innerHTML = '✅ Submit-Button: Aktiviert';
      statusButton.style.color = 'green';
    } else {
      statusButton.innerHTML = '❌ Submit-Button: Deaktiviert';
      statusButton.style.color = 'red';
    }
  }
}
