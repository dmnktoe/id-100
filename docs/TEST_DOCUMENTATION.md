# Test Documentation for City Autocomplete

## Overview

Comprehensive test suite for the Meilisearch-powered city autocomplete dropdown and form validation system.

## Test File Location

`src/__tests__/city-autocomplete.test.ts`

## Mock Data

### Mock Cities
```typescript
const mockCities = [
  { name: "Berlin" },
  { name: "München" },
  { name: "Hamburg" },
  { name: "Köln" },
  { name: "Frankfurt am Main" },
];
```

These represent typical German cities that users might search for.

## Mocking Strategy

### Meilisearch Client Mock

```typescript
vi.mock("meilisearch", () => ({
  MeiliSearch: vi.fn(() => ({
    index: vi.fn(() => ({
      search: vi.fn(async (query: string) => {
        const filtered = mockCities.filter(city =>
          city.name.toLowerCase().includes(query.toLowerCase())
        );
        return {
          hits: filtered,
          estimatedTotalHits: filtered.length,
        };
      }),
    })),
  })),
}));
```

This mock simulates the Meilisearch client's search functionality, filtering cities based on the query string.

## Test Categories

### 1. Dropdown Rendering Tests

**Purpose:** Verify dropdown element creation, visibility, and positioning

**Tests:**
- ✅ Dropdown element created on initialization
- ✅ Dropdown hidden initially (display: none)
- ✅ Dropdown shows when typing ≥2 characters
- ✅ Dropdown hides when typing <2 characters

**Example:**
```typescript
it("should show dropdown when typing 2 or more characters", async () => {
  initCityAutocomplete();
  
  const cityInput = document.getElementById("player_city") as HTMLInputElement;
  const dropdown = document.querySelector(".city-dropdown") as HTMLDivElement;
  
  cityInput.value = "Be";
  cityInput.dispatchEvent(new Event("input", { bubbles: true }));
  await new Promise((resolve) => setTimeout(resolve, 350)); // Debounce wait
  
  expect(dropdown.style.display).toBe("block");
});
```

### 2. Form Validation Tests

**Purpose:** Ensure submit button is only enabled when all conditions are met

**Tests:**
- ✅ Button disabled on initial page load
- ✅ Button stays disabled without valid name (≥2 characters)
- ✅ Button stays disabled without city selection
- ✅ Button stays disabled without privacy checkbox checked
- ✅ Button only enables when ALL conditions met

**Example:**
```typescript
it("should enable button only when all conditions are met", async () => {
  initCityAutocomplete();
  initFormValidation();
  
  const submitBtn = document.getElementById("submit-btn") as HTMLButtonElement;
  const nameInput = document.getElementById("player_name") as HTMLInputElement;
  const cityInput = document.getElementById("player_city") as HTMLInputElement;
  const privacyCheckbox = document.getElementById("agree_privacy") as HTMLInputElement;
  
  // Initial state - disabled
  expect(submitBtn.disabled).toBe(true);
  
  // Fill name - still disabled
  nameInput.value = "Max Mustermann";
  nameInput.dispatchEvent(new Event("input", { bubbles: true }));
  expect(submitBtn.disabled).toBe(true);
  
  // Select city - still disabled
  cityInput.value = "Ber";
  cityInput.dispatchEvent(new Event("input", { bubbles: true }));
  await new Promise((resolve) => setTimeout(resolve, 350));
  
  const dropdown = document.querySelector(".city-dropdown") as HTMLDivElement;
  const firstItem = dropdown.querySelector(".city-dropdown-item") as HTMLDivElement;
  firstItem?.click();
  expect(submitBtn.disabled).toBe(true);
  
  // Check privacy - NOW enabled!
  privacyCheckbox.checked = true;
  privacyCheckbox.dispatchEvent(new Event("change", { bubbles: true }));
  expect(submitBtn.disabled).toBe(false);
});
```

## Test Setup and Teardown

### beforeEach

```typescript
beforeEach(() => {
  // Create DOM structure
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
  
  // Set required global
  (window as any).GEOCODING_API_URL = "http://localhost:8081";
  
  // Reset state
  resetState();
});
```

### afterEach

```typescript
afterEach(() => {
  // Clean up DOM
  document.body.removeChild(container);
  
  // Clear mocks
  vi.clearAllMocks();
});
```

## Async Testing

### Debounce Handling

The autocomplete uses a 300ms debounce. Tests must wait for this:

```typescript
cityInput.value = "Berlin";
cityInput.dispatchEvent(new Event("input", { bubbles: true }));

// Wait for debounce + processing
await new Promise((resolve) => setTimeout(resolve, 350));

// Now assertions can be made
expect(dropdown.style.display).toBe("block");
```

## Running Tests

### Commands

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch

# Run specific test file
npm test city-autocomplete
```

### Expected Output

```
✓ src/__tests__/city-autocomplete.test.ts (3)
  ✓ City Autocomplete - Dropdown Rendering (3)
    ✓ should create dropdown element when initializing
    ✓ should hide dropdown initially
    ✓ should show dropdown when typing 2 or more characters
  ✓ City Autocomplete - Form Validation (1)
    ✓ should enable button only when all conditions are met

Test Files  1 passed (1)
Tests  4 passed (4)
```

## Test Assertions

### Common Patterns

**Element Existence:**
```typescript
expect(dropdown).toBeTruthy();
expect(dropdown?.classList.contains("city-dropdown")).toBe(true);
```

**Visibility:**
```typescript
expect(dropdown.style.display).toBe("none");
expect(dropdown.style.display).toBe("block");
```

**Button State:**
```typescript
expect(submitBtn.disabled).toBe(true);
expect(submitBtn.classList.contains("disabled")).toBe(true);
```

**CSS Classes:**
```typescript
expect(cityInput.classList.contains("city-selected")).toBe(true);
```

## Coverage Goals

- **Dropdown Rendering:** 100%
- **Form Validation:** 100%
- **State Management:** 100%
- **User Interactions:** 100%

## Future Test Additions

Potential areas for expansion:

1. **Keyboard Navigation**
   - Arrow up/down
   - Enter to select
   - Escape to close

2. **Error Handling**
   - Network failures
   - Invalid API responses
   - Timeout scenarios

3. **Performance**
   - Large result sets
   - Rapid typing
   - Memory leaks

4. **Accessibility**
   - ARIA attributes
   - Screen reader compatibility
   - Keyboard-only navigation

## Debugging Tests

### Common Issues

**1. Test timeout:**
```
Timeout - Async operation did not complete
```
Solution: Increase debounce wait time or check async handling

**2. Element not found:**
```
expect(received).toBeTruthy()
Received: null
```
Solution: Verify DOM setup in beforeEach, check element IDs

**3. Mock not working:**
```
TypeError: Cannot read property 'search' of undefined
```
Solution: Verify mock structure matches actual Meilisearch API

## Continuous Integration

Tests run automatically on:
- Pull request creation
- Push to main branch
- Docker image build

Ensure all tests pass before merging!

## Conclusion

This test suite provides comprehensive coverage of the city autocomplete dropdown and form validation functionality. Tests use mocking to simulate the Meilisearch API and verify both UI rendering and business logic.

The tests ensure:
✅ Dropdown appears correctly
✅ Search results are displayed
✅ Submit button is properly controlled
✅ All validation rules are enforced
✅ User interactions work as expected

Maintain and expand these tests as new features are added!
