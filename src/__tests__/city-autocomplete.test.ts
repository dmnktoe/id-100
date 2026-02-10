import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  initCityAutocomplete,
  initFormValidation,
  resetState,
} from "../lib/city-autocomplete";

// Mock city data for testing
const mockCities = [
  { name: "Berlin" },
  { name: "München" },
  { name: "Hamburg" },
  { name: "Köln" },
  { name: "Frankfurt am Main" },
];

// Mock Meilisearch
vi.mock("meilisearch", () => {
  return {
    MeiliSearch: vi.fn(() => ({
      index: vi.fn(() => ({
        search: vi.fn(async (query: string) => {
          // Simulate search based on query
          const filtered = mockCities.filter((city) =>
            city.name.toLowerCase().includes(query.toLowerCase())
          );
          return {
            hits: filtered,
            estimatedTotalHits: filtered.length,
          };
        }),
      })),
    })),
  };
});

describe("City Autocomplete - Dropdown Rendering", () => {
  let container: HTMLDivElement;

  beforeEach(() => {
    // Setup DOM
    container = document.createElement("div");
    container.innerHTML = `
      <form>
        <input type="text" id="player_name" name="player_name" />
        <input type="text" id="player_city" name="player_city" />
        <input type="checkbox" id="agree_privacy" name="agree_privacy" />
        <button type="submit" id="submit-btn">Weiter zum Upload</button>
      </form>
    `;
    document.body.appendChild(container);

    // Set window.GEOCODING_API_URL for tests
    (window as any).GEOCODING_API_URL = "http://localhost:8081";

    resetState();
  });

  afterEach(() => {
    document.body.removeChild(container);
    vi.clearAllMocks();
  });

  it("should create dropdown element when initializing", () => {
    initCityAutocomplete();

    const dropdown = document.querySelector(".city-dropdown");
    expect(dropdown).toBeTruthy();
    expect(dropdown?.classList.contains("city-dropdown")).toBe(true);
  });

  it("should hide dropdown initially", () => {
    initCityAutocomplete();

    const dropdown = document.querySelector(
      ".city-dropdown"
    ) as HTMLDivElement;
    expect(dropdown).toBeTruthy();
    expect(dropdown?.style.display).toBe("none");
  });

  it("should show dropdown when typing 2 or more characters", async () => {
    initCityAutocomplete();

    const cityInput = document.getElementById(
      "player_city"
    ) as HTMLInputElement;
    const dropdown = document.querySelector(
      ".city-dropdown"
    ) as HTMLDivElement;

    // Type less than 2 characters - dropdown should stay hidden
    cityInput.value = "B";
    cityInput.dispatchEvent(new Event("input", { bubbles: true }));
    await new Promise((resolve) => setTimeout(resolve, 350)); // Wait for debounce
    expect(dropdown.style.display).toBe("none");

    // Type 2 or more characters - dropdown should appear
    cityInput.value = "Be";
    cityInput.dispatchEvent(new Event("input", { bubbles: true }));
    await new Promise((resolve) => setTimeout(resolve, 350)); // Wait for debounce

    expect(dropdown.style.display).toBe("block");
  });

  it("should enable button only when all conditions are met", async () => {
    initCityAutocomplete();
    initFormValidation();

    const submitBtn = document.getElementById("submit-btn") as HTMLButtonElement;
    const nameInput = document.getElementById("player_name") as HTMLInputElement;
    const cityInput = document.getElementById("player_city") as HTMLInputElement;
    const privacyCheckbox = document.getElementById(
      "agree_privacy"
    ) as HTMLInputElement;

    // Initial state
    expect(submitBtn.disabled).toBe(true);

    // Fill in name
    nameInput.value = "Max Mustermann";
    nameInput.dispatchEvent(new Event("input", { bubbles: true }));
    expect(submitBtn.disabled).toBe(true); // Still disabled

    // Select city
    cityInput.value = "Ber";
    cityInput.dispatchEvent(new Event("input", { bubbles: true }));
    await new Promise((resolve) => setTimeout(resolve, 350));

    const dropdown = document.querySelector(".city-dropdown") as HTMLDivElement;
    const firstItem = dropdown.querySelector(
      ".city-dropdown-item"
    ) as HTMLDivElement;
    firstItem?.click();
    expect(submitBtn.disabled).toBe(true); // Still disabled

    // Check privacy
    privacyCheckbox.checked = true;
    privacyCheckbox.dispatchEvent(new Event("change", { bubbles: true }));

    // Now button should be enabled
    expect(submitBtn.disabled).toBe(false);
    expect(submitBtn.classList.contains("disabled")).toBe(false);
  });
});
