/**
 * Tests for brand-animation module
 */
import { describe, it, expect, beforeEach, vi } from "vitest";
import { initBrandAnimation } from "../lib/brand-animation";

describe("initBrandAnimation", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    sessionStorage.clear();
  });

  it("should handle missing brand elements gracefully", () => {
    expect(() => initBrandAnimation()).not.toThrow();
  });

  it("should trigger reverse animation on homepage when returning from subpage", () => {
    vi.useFakeTimers();

    // Setup: brand-full element exists and brandAnimated flag is set
    document.body.innerHTML = '<div class="brand-full"></div>';
    sessionStorage.setItem("brandAnimated", "1");

    const brandFull = document.querySelector<HTMLElement>(".brand-full");

    initBrandAnimation();

    // Should have reverse-animated class
    expect(brandFull?.classList.contains("reverse-animated")).toBe(true);

    // Session storage should be reset
    expect(sessionStorage.getItem("brandAnimated")).toBe("0");

    // After 800ms, class should be removed
    vi.advanceTimersByTime(800);
    expect(brandFull?.classList.contains("reverse-animated")).toBe(false);

    vi.useRealTimers();
  });

  it("should trigger forward animation on subpage when not yet animated", () => {
    // Setup: brand-compact element exists and brandAnimated flag is not set
    document.body.innerHTML = '<div class="brand-compact"></div>';
    sessionStorage.removeItem("brandAnimated");

    const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

    initBrandAnimation();

    // Should have animated class
    expect(brandCompact?.classList.contains("animated")).toBe(true);

    // Session storage should be set
    expect(sessionStorage.getItem("brandAnimated")).toBe("1");
  });

  it("should not animate brand-compact if already animated", () => {
    document.body.innerHTML = '<div class="brand-compact"></div>';
    sessionStorage.setItem("brandAnimated", "1");

    const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

    initBrandAnimation();

    // Should not have animated class since already animated
    expect(brandCompact?.classList.contains("animated")).toBe(false);
  });

  it("should not modify brandAnimated flag when no brand elements exist", () => {
    sessionStorage.setItem("brandAnimated", "1");

    initBrandAnimation();

    // Flag should remain unchanged
    expect(sessionStorage.getItem("brandAnimated")).toBe("1");
  });
});
