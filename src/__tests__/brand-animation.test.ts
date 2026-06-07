/**
 * Tests for brand-animation module
 */
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { initBrandAnimation } from "../lib/brand-animation";

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
    document.body.innerHTML = '<div class="brand-full"></div>';
    sessionStorage.setItem("brandAnimated", "1");

    const brandFull = document.querySelector<HTMLElement>(".brand-full");

    initBrandAnimation();

    expect(brandFull?.classList.contains("reverse-animated")).toBe(true);
    expect(sessionStorage.getItem("brandAnimated")).toBe("0");
  });

  it("should trigger forward animation on subpage when not yet animated", () => {
    document.body.innerHTML = '<div class="brand-compact"></div>';
    sessionStorage.removeItem("brandAnimated");

    const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

    initBrandAnimation();

    expect(brandCompact?.classList.contains("animated")).toBe(true);
    expect(sessionStorage.getItem("brandAnimated")).toBe("1");
  });

  it("should not animate brand-compact if already animated", () => {
    document.body.innerHTML = '<div class="brand-compact"></div>';
    sessionStorage.setItem("brandAnimated", "1");

    const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

    initBrandAnimation();

    expect(brandCompact?.classList.contains("animated")).toBe(false);
  });

  it("should not modify brandAnimated flag when no brand elements exist", () => {
    sessionStorage.setItem("brandAnimated", "1");

    initBrandAnimation();

    expect(sessionStorage.getItem("brandAnimated")).toBe("1");
  });

  describe("prefers-reduced-motion", () => {
    it("should not add the forward class but still update the flag", () => {
      setReducedMotion(true);
      document.body.innerHTML = '<div class="brand-compact"></div>';

      const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

      initBrandAnimation();

      expect(brandCompact?.classList.contains("animated")).toBe(false);
      expect(sessionStorage.getItem("brandAnimated")).toBe("1");
    });

    it("should not add the reverse class but still update the flag", () => {
      setReducedMotion(true);
      document.body.innerHTML = '<div class="brand-full"></div>';
      sessionStorage.setItem("brandAnimated", "1");

      const brandFull = document.querySelector<HTMLElement>(".brand-full");

      initBrandAnimation();

      expect(brandFull?.classList.contains("reverse-animated")).toBe(false);
      expect(sessionStorage.getItem("brandAnimated")).toBe("0");
    });
  });

  it("should not throw when sessionStorage access fails", () => {
    document.body.innerHTML = '<div class="brand-compact"></div>';

    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new DOMException("denied", "SecurityError");
    });
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("denied", "SecurityError");
    });

    expect(() => initBrandAnimation()).not.toThrow();

    const brandCompact = document.querySelector<HTMLElement>(".brand-compact");
    expect(brandCompact?.classList.contains("animated")).toBe(true);
  });
});
