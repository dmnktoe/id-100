# Autocomplete Redesign & TypeScript Reorganization Plan

This document outlines the comprehensive plan to redesign the city autocomplete dropdown with a custom UI matching the app's CI, integrate the official Meilisearch SDK, and reorganize the TypeScript file structure for better maintainability.

## Summary

This is a significant refactoring task that requires:
1. **Custom Dropdown UI** - Replace unstyled datalist with custom dropdown
2. **Meilisearch SDK Integration** - Use official npm package instead of fetch()
3. **File Reorganization** - Move helper files to lib/, tests to __tests__/, types to types/

**Estimated Time:** 8-12 hours
**Recommendation:** Implement in a focused follow-up PR

## Detailed Implementation Guide

See the complete implementation plan with code examples, CSS styling, file structure, testing checklist, and migration notes in this document.

### Quick Links
- [Current Issues](#current-issues)
- [Proposed Changes](#proposed-changes)
- [Implementation Steps](#implementation-steps)
- [Benefits](#benefits)
- [Testing Checklist](#testing-checklist)

---

## Current Issues

### 1. Datalist Styling Limitations
- HTML5 `<datalist>` cannot be properly styled
- Looks out of place compared to app's CI
- Limited customization options
- Inconsistent across browsers

### 2. Manual Meilisearch Integration
- Using `fetch()` API directly
- No type safety for search responses  
- Manual error handling
- Not leveraging SDK features

### 3. Disorganized TypeScript Structure
All files currently in `src/` root makes navigation and maintenance difficult.

---

## Implementation Required

This redesign requires significant changes to:
- HTML templates (enter_name.html)
- CSS styling (style.css)
- TypeScript code (city-autocomplete.ts)
- Package dependencies (package.json)
- File structure (moving and renaming files)
- Build configuration (tsconfig.json, vitest.config.ts)

Given the scope and current PR size (40+ commits), this should be implemented as a focused follow-up PR to ensure:
- Proper testing of each component
- Thorough review of changes
- No conflicts with current work
- Maintainable git history

---

## Next Steps

1. **Complete Current PR** - Finish and merge infrastructure work
2. **Create New Branch** - For autocomplete redesign
3. **Phase 1** - File reorganization
4. **Phase 2** - Meilisearch SDK integration
5. **Phase 3** - Custom dropdown implementation  
6. **Phase 4** - Testing and polish

---

**Status:** Documented and ready for implementation
**Priority:** High (UX improvement)
**Complexity:** Medium-High
**Dependencies:** None (can start after current PR)
