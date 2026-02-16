/**
 * Main entry point for ID-100 client-side application
 * Initializes all modules when the DOM is ready
 */

import { initBrandAnimation } from "./lib/brand-animation";
import { initDrawer } from "./lib/drawer";
import { initLazyImages } from "./lib/lazy-images";
import { initFormHandlers } from "./lib/form-handler";
import { initCityAutocomplete, initFormValidation } from "./lib/city-autocomplete";
import { initAdminDashboard } from "./lib/admin-dashboard";
import { initUpload } from "./lib/upload";
import "./lib/favicon-emoji";

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

  // Form validation for name entry
  initFormValidation();

  // Initialize lazy images on first paint
  initLazyImages();

  // Admin dashboard functionality
  initAdminDashboard();

  // Upload page functionality
  initUpload();
})();
