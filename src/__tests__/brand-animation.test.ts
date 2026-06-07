/**
 * Tests for brand-animation module
 */
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { initBrandAnimation } from "../lib/brand-animation";

/** Stub matchMedia so tests can control prefers-reduced-motion. */
function setReducedMotion(reduce: boolean): void {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: reduce && query.includes("prefers-reduced-motion"),
    media: query,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  })) as unknown as typeof window.matchMedia;
}

describe("initBrandAnimation", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    sessionStorage.clear();
    setReducedMotion(false);
  });

  afterEach(() => {
    vi.restoreAllMocks();
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

    // After the animation duration, the class should be removed
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

  describe("prefers-reduced-motion", () => {
    it("should not add animation classes but still update the flag (forward)", () => {
      setReducedMotion(true);
      document.body.innerHTML = '<div class="brand-compact"></div>';

      const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

      initBrandAnimation();

      // No animation class, but the state machine still advances
      expect(brandCompact?.classList.contains("animated")).toBe(false);
      expect(sessionStorage.getItem("brandAnimated")).toBe("1");
    });

    it("should not add animation classes but still update the flag (reverse)", () => {
      setReducedMotion(true);
      document.body.innerHTML = '<div class="brand-full"></div>';
      sessionStorage.setItem("brandAnimated", "1");

      const brandFull = document.querySelector<HTMLElement>(".brand-full");

      initBrandAnimation();

      expect(brandFull?.classList.contains("reverse-animated")).toBe(false);
      expect(sessionStorage.getItem("brandAnimated")).toBe("0");
    });
  });

  describe("unavailable sessionStorage", () => {
    it("should not throw when sessionStorage access fails", () => {
      document.body.innerHTML = '<div class="brand-compact"></div>';

      const getItemSpy = vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
        throw new DOMException("denied", "SecurityError");
      });
      const setItemSpy = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
        throw new DOMException("denied", "SecurityError");
      });

      // Must not throw, otherwise it would break the whole app init chain.
      expect(() => initBrandAnimation()).not.toThrow();

      // With storage unreadable we treat it as "not yet animated" and still
      // play the forward animation.
      const brandCompact = document.querySelector<HTMLElement>(".brand-compact");
      expect(brandCompact?.classList.contains("animated")).toBe(true);

      getItemSpy.mockRestore();
      setItemSpy.mockRestore();
    });
  });
});
