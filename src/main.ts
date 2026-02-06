/**
 * Main entry point for ID-100 client-side application
 * Initializes all modules when the DOM is ready
 */

import { initBrandAnimation } from "./brand-animation";
import { initDrawer } from "./drawer";
import { initLazyImages } from "./lazy-images";
import { initFormHandlers } from "./form-handler";
import { initCityAutocomplete } from "./city-autocomplete";
import "./favicon-emoji";

// Initialize all modules when DOM is ready
(() => {
  // Brand animation logic
  initBrandAnimation();

  // Drawer functionality
  initDrawer();

  // Form handlers
  initFormHandlers();

  // City autocomplete
  initCityAutocomplete();

  // Initialize lazy images on first paint
  initLazyImages();
})();
