# Dropdown and Button Fixes

## Bug 1: Dropdown Not Appearing

### Problem
Custom dropdown wasn't showing when typing in city field, even though the element was being created.

### Root Cause
The dropdown was being appended to the wrong parent element using `appendChild()`, which doesn't guarantee proper positioning in the DOM tree.

### Solution
Changed from:
```typescript
cityInput.parentElement?.appendChild(dropdown);
```

To:
```typescript
if (cityInput.parentNode) {
  cityInput.parentNode.insertBefore(dropdown, cityInput.nextSibling);
}
```

This ensures the dropdown is inserted immediately after the input element, allowing proper CSS positioning.

## Bug 2: Submit Button Always Enabled

### Problem
Submit button was enabled even when validation requirements weren't met (name, city, privacy checkbox).

### Root Causes
1. No initial disabled state was set
2. CSS class not being managed for visual feedback
3. State not properly initialized on page load

### Solutions

**1. Set Initial Disabled State:**
```typescript
// In initCityAutocomplete()
if (submitBtn) {
  submitBtn.disabled = true;
}

// In initFormValidation()
submitBtn.disabled = true;
submitBtn.classList.add("disabled");
```

**2. Manage CSS Classes:**
```typescript
const allValid = nameValid && privacyAccepted && citySelected;
submitBtn.disabled = !allValid;

if (allValid) {
  submitBtn.classList.remove("disabled");
} else {
  submitBtn.classList.add("disabled");
}
```

**3. Consistent Updates:**
Both `updateSubmitButton()` and `initFormValidation()` now manage both the disabled attribute and CSS class.

## Testing Checklist

- [ ] Dropdown appears when typing ≥2 characters
- [ ] Dropdown positioned directly below input
- [ ] Dropdown contains city results from Meilisearch
- [ ] Button disabled on initial page load
- [ ] Button stays disabled until name ≥2 characters
- [ ] Button stays disabled until city selected from dropdown
- [ ] Button stays disabled until privacy checkbox checked
- [ ] Button only enables when ALL three conditions met
- [ ] Green border on city field when valid city selected
- [ ] CSS "disabled" class applied/removed correctly

## Files Changed
- `src/lib/city-autocomplete.ts` - Fixed both issues
