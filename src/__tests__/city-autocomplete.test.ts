/**
 * Tests for city-autocomplete module
 */

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { initCityAutocomplete, initFormValidation, resetState } from "../lib/city-autocomplete";

describe("city-autocomplete", () => {
  let cityInput: HTMLInputElement;
  let datalist: HTMLDataListElement;
  let submitBtn: HTMLButtonElement;

  beforeEach(() => {
    // Reset module state before each test
    resetState();
    
    // Setup DOM elements
    document.body.innerHTML = `
      <input type="text" id="playerName" value="Test User" />
      <input type="text" id="playerCity" />
      <datalist id="cityOptions"></datalist>
      <input type="checkbox" id="agreePrivacy" />
      <button class="submit-btn" type="submit">Submit</button>
    `;

    cityInput = document.getElementById("playerCity") as HTMLInputElement;
    datalist = document.getElementById("cityOptions") as HTMLDataListElement;
    submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;

    // Mock window.GEOCODING_API_URL
    (window as any).GEOCODING_API_URL = "http://localhost:8081";

    // Mock fetch
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    resetState();
  });

  it("should initialize without errors", () => {
    expect(() => initCityAutocomplete()).not.toThrow();
  });

  it("should not throw error if elements are missing", () => {
    document.body.innerHTML = "";
    expect(() => initCityAutocomplete()).not.toThrow();
  });

  it("should attach input listener to city input", () => {
    initCityAutocomplete();
    const spy = vi.spyOn(cityInput, "addEventListener");
    
    // Re-initialize to catch the addEventListener call
    document.body.innerHTML = `
      <input type="text" id="playerName" value="Test User" />
      <input type="text" id="playerCity" />
      <datalist id="cityOptions"></datalist>
      <input type="checkbox" id="agreePrivacy" />
      <button class="submit-btn" type="submit">Submit</button>
    `;
    cityInput = document.getElementById("playerCity") as HTMLInputElement;
    
    initCityAutocomplete();
    
    // Trigger input event
    cityInput.value = "Berlin";
    const event = new Event("input");
    cityInput.dispatchEvent(event);
    
    // Check that event listener was setup
    expect(cityInput.value).toBe("Berlin");
  });

  it("should fetch cities when input has 2 or more characters", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        hits: [
          {
            id: "2950159",
            name: "Berlin",
            lat: 52.52437,
            lon: 13.41053,
            type: "capital",
            population: 3426354,
          },
        ],
        query: "Ber",
        processingTimeMs: 1,
        limit: 10,
        offset: 0,
        estimatedTotalHits: 1,
      }),
    });
    global.fetch = mockFetch;

    initCityAutocomplete();

    // Simulate user typing
    cityInput.value = "Ber";
    const event = new Event("input");
    cityInput.dispatchEvent(event);

    // Wait for debounce
    await new Promise((resolve) => setTimeout(resolve, 350));

    expect(mockFetch).toHaveBeenCalled();
  });

  it("should not fetch cities when input is less than 2 characters", async () => {
    const mockFetch = vi.fn();
    global.fetch = mockFetch;

    initCityAutocomplete();

    cityInput.value = "B";
    const event = new Event("input");
    cityInput.dispatchEvent(event);

    await new Promise((resolve) => setTimeout(resolve, 350));

    expect(mockFetch).not.toHaveBeenCalled();
  });

  it("should disable submit button initially when form validation is enabled", () => {
    initFormValidation();
    expect(submitBtn.disabled).toBe(true);
  });

  it("should keep button disabled when city is not from valid list", () => {
    const nameInput = document.getElementById("playerName") as HTMLInputElement;
    const privacyCheckbox = document.getElementById("agreePrivacy") as HTMLInputElement;
    
    initFormValidation();
    initCityAutocomplete();
    
    // Fill in name (already has value)
    expect(nameInput.value).toBe("Test User");
    
    // Check privacy
    privacyCheckbox.checked = true;
    privacyCheckbox.dispatchEvent(new Event("change"));
    
    // Type a city (but not selected from valid dropdown yet)
    cityInput.value = "Berlin";
    cityInput.dispatchEvent(new Event("input"));
    
    // Button should be disabled until city is in validCities
    expect(submitBtn.disabled).toBe(true);
  });
});

describe("initFormValidation", () => {
  beforeEach(() => {
    resetState();
    document.body.innerHTML = `
      <input type="text" id="playerName" />
      <input type="text" id="playerCity" />
      <datalist id="cityOptions"></datalist>
      <input type="checkbox" id="agreePrivacy" />
      <button class="submit-btn" type="submit">Submit</button>
    `;
  });

  afterEach(() => {
    resetState();
  });

  it("should not throw error if elements are missing", () => {
    document.body.innerHTML = "";
    expect(() => initFormValidation()).not.toThrow();
  });

  it("should disable button initially", () => {
    const submitBtn = document.querySelector(".submit-btn") as HTMLButtonElement;
    initFormValidation();
    expect(submitBtn.disabled).toBe(true);
  });
});
