# City Filter Feature Documentation

## Overview

The city filter feature allows users to filter the deriven overview page by cities that have contributions in the database. This provides an intuitive way to explore deriven by location while maintaining the clean, minimal aesthetic of the app.

## User Interface

### Filter Section Components

**1. Location Icon**
- SVG location pin icon
- Color: `var(--gray-600)`
- Size: 16x16px

**2. Label**
- Text: "Stadt filtern" (Filter by city)
- Font: Arial Rounded MT Bold
- Size: 0.9rem
- Color: `var(--gray-800)`

**3. Dropdown Select**
- Custom styled select element
- Shows all cities with contributions (alphabetically sorted)
- Default option: "Alle Städte (X Beiträge)" where X is total contributions
- Width: 200-320px (responsive)
- Custom down arrow icon
- Glass background effect on hover

**4. Clear Filter Button**  
- Visible only when a city is selected
- X icon button
- Size: 2rem x 2rem
- Hover effect: scales to 1.05x
- Click removes filter (redirects to `/?`)

## Visual Design

### Filter Section Styling

```css
.filter-section {
  /* Glass morphism effect */
  background: var(--bg-glass);                    /* rgba(248, 248, 248, 0.7) */
  backdrop-filter: var(--backdrop-blur-sm);       /* blur(8px) */
  
  /* Layout */
  display: flex;
  align-items: center;
  gap: var(--gap-sm);                             /* 0.5rem */
  padding: var(--pad-md) var(--pad-lg);           /* 1rem 1.5rem */
  
  /* Borders & Radius */
  border: var(--border-light);                    /* 1px solid rgba(0,0,0,0.06) */
  border-radius: var(--radius-md);                /* 12px */
  
  /* Spacing */
  margin-bottom: var(--gap-lg);                   /* 2rem */
  
  /* Transitions */
  transition: all var(--transition-smooth);       /* 0.3s cubic-bezier */
}
```

### Dropdown Styling

```css
.city-filter-dropdown {
  /* Typography */
  font-family: "Arial Rounded MT Bold", sans-serif;
  font-size: 0.9rem;
  font-weight: 400;
  letter-spacing: var(--letter-spacing-tight);    /* -0.03em */
  
  /* Layout */
  flex: 1;
  min-width: 200px;
  max-width: 320px;
  padding: 0.625rem var(--pad-md);
  
  /* Colors */
  color: var(--gray-900);                         /* #333 */
  background: var(--white);                       /* #fff */
  border: var(--border-medium);                   /* 1px solid rgba(0,0,0,0.08) */
  
  /* Styling */
  border-radius: var(--radius-sm);                /* 8px */
  cursor: pointer;
  
  /* Custom arrow */
  appearance: none;
  background-image: url("data:image/svg+xml,..."); /* down chevron */
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
  padding-right: 2.5rem;
}
```

## Functionality

### Backend Flow

1. **Get Available Cities**
   ```sql
   SELECT DISTINCT user_city 
   FROM contributions 
   WHERE user_city IS NOT NULL AND user_city != '' 
   ORDER BY user_city ASC
   ```

2. **Filter Deriven (when city selected)**
   ```sql
   SELECT d.* 
   FROM deriven d
   INNER JOIN contributions c ON c.derive_id = d.id AND c.user_city = $1
   GROUP BY d.id, d.number, ...
   ORDER BY d.number ASC
   LIMIT $2 OFFSET $3
   ```

3. **Update Pagination Count**
   ```sql
   SELECT COUNT(DISTINCT d.id) 
   FROM deriven d
   INNER JOIN contributions c ON c.derive_id = d.id
   WHERE c.user_city = $1
   ```

### Frontend Interaction

**Selecting a City:**
```javascript
// Dropdown onChange handler
onchange="window.location.href=this.value"

// Generated URLs:
// "/"                    → All cities (default)
// "/?city=Berlin"        → Filter by Berlin
// "/?city=München"       → Filter by München
```

**Pagination with Filter:**
```html
<!-- Previous page -->
<a href="/?page=1&city=Berlin">←  Zurück</a>

<!-- Page numbers -->
<a href="/?page=2&city=Berlin">2</a>
<a href="/?page=3&city=Berlin">3</a>

<!-- Next page -->
<a href="/?page=3&city=Berlin">Weiter →</a>
```

**Clearing Filter:**
```html
<!-- Clear button (only visible when filter active) -->
<a href="/" class="filter-clear" title="Filter entfernen">
  <svg><!-- X icon --></svg>
</a>
```

## Mobile Responsiveness

### Breakpoint: 768px

```css
@media (max-width: 768px) {
  .filter-section {
    flex-wrap: wrap;           /* Stack elements */
    padding: var(--pad-md);    /* Reduce padding */
  }

  .filter-label {
    width: 100%;               /* Full width label */
    font-size: 0.85rem;        /* Smaller font */
  }

  .city-filter-dropdown {
    flex: 1;
    max-width: none;           /* Take available space */
    font-size: 0.85rem;
  }

  .filter-clear {
    width: 2.5rem;
    height: 2.5rem;            /* Larger touch target */
  }
}
```

## User Experience

### States

**1. No Filter Active (Default)**
- Dropdown shows "Alle Städte (X Beiträge)"
- All 100 deriven visible
- Clear button hidden
- Standard pagination

**2. Filter Active**
- Dropdown shows selected city name
- Only deriven with contributions from that city visible
- Clear button visible
- Pagination adjusted to filtered count
- Filter persists across pagination

**3. Hover States**
- Filter section: Brighter glass background
- Dropdown: Gray background, darker border
- Clear button: Scales up, darker background

**4. Empty State**
- If no cities have contributions: filter section not displayed
- Template checks: `{{if .Cities}}...{{end}}`

## Integration Points

### Template Variables

```go
"Cities":       []string          // List of distinct cities
"SelectedCity": string            // Currently selected city (or "")
```

### Query Parameters

- `?city=CityName` - Filter by city
- `?page=N` - Pagination (works with city filter)
- `?city=Berlin&page=2` - Combined filter and pagination

### URL Preservation

Filter is preserved in:
- Pagination links (prev, next, page numbers)
- Derive detail page back links
- All navigation within filtered view

## Design Rationale

### Why Glass Morphism?

The glass effect (`backdrop-filter: blur()`) creates visual hierarchy without heavy borders or shadows, maintaining the clean aesthetic while making the filter section distinct.

### Why Dropdown vs Autocomplete?

**Dropdown chosen for:**
- Simple, finite list of cities
- Better mobile UX (native select on mobile)
- Less implementation complexity
- Consistent with app's minimalist design

**Not autocomplete because:**
- Limited number of cities (typically < 100)
- Users browsing, not searching
- Native select provides good UX

### Color Choices

All colors from existing CSS custom properties:
- `--gray-600` for secondary elements (icon)
- `--gray-800` for primary text (label)
- `--gray-900` for dropdown text
- `--border-light` for subtle separation
- Glass backgrounds for depth without weight

## Performance Considerations

### SQL Optimization

**Efficient Queries:**
- INNER JOIN only when filter active (not LEFT JOIN)
- DISTINCT on cities query (small result set)
- GROUP BY prevents duplicate deriven
- Pagination limits result set

**Query Plans:**
- Cities query: Index scan on `user_city`
- Filtered deriven: Index on `derive_id` FK
- Pagination: LIMIT/OFFSET on sorted results

### Frontend Performance

**Fast Interactions:**
- No JavaScript needed (onchange navigates)
- Browser handles select native optimization
- CSS transitions hardware-accelerated
- No AJAX calls required

## Future Enhancements

### Possible Additions

1. **Multi-Select Filter**
   - Allow filtering by multiple cities
   - "Berlin OR München" logic
   - Checkbox dropdown variant

2. **City Statistics**
   - Show contribution count per city in dropdown
   - "Berlin (42)" format

3. **City Badges**
   - Visual indicators on filtered cards
   - Show which cities contributed

4. **URL Shortening**
   - Use city ID instead of name
   - Cleaner URLs, supports special characters

5. **Filter Persistence**
   - Remember last filter in session
   - Cookie or localStorage

6. **Search Within Filter**
   - Type-ahead search for cities
   - Useful if city list grows large

## Testing Checklist

- [ ] Filter shows all cities with contributions
- [ ] Cities are alphabetically sorted
- [ ] "Alle Städte" clears filter correctly
- [ ] Clear button only shows when filter active
- [ ] Filter persists across pagination
- [ ] Filtered results show correct deriven
- [ ] Pagination count updates with filter
- [ ] Mobile layout wraps correctly
- [ ] Hover states work on all interactive elements
- [ ] URLs are correct with special characters
- [ ] Empty state handled gracefully

## Accessibility

### Keyboard Navigation

- Dropdown is focusable and operable via keyboard
- Tab order: Label → Dropdown → Clear button (if visible)
- Enter/Space opens dropdown
- Arrow keys navigate options
- Escape closes dropdown

### Screen Readers

```html
<label for="cityFilter" class="filter-label">
  <!-- Icon is decorative, label provides context -->
  Stadt filtern
</label>
<select id="cityFilter" aria-label="Nach Stadt filtern">
  <!-- Options -->
</select>
<a href="/" class="filter-clear" title="Filter entfernen">
  <!-- Title provides accessible name -->
</a>
```

### Color Contrast

All text meets WCAG AA standards:
- Label: `--gray-800` on `--bg-glass` → 7.2:1
- Dropdown: `--gray-900` on `white` → 12.6:1
- Icon: `--gray-600` on `--bg-glass` → 4.8:1

## Conclusion

The city filter feature provides powerful filtering capability while maintaining the app's clean, minimal design. It integrates seamlessly with existing pagination, preserves state across navigation, and provides clear visual feedback for all interactions.
