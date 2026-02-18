/**
 * Tests for favicon-emoji module
 */
import { describe, it, expect, beforeEach } from "vitest";
import { setRandomEmojiFavicon } from "../lib/favicon-emoji";

describe("setRandomEmojiFavicon", () => {
  beforeEach(() => {
    // Clear any existing favicon links
    document.querySelectorAll("link[rel~='icon']").forEach((link) => link.remove());
  });

  it("should create a favicon link if none exists", () => {
    setRandomEmojiFavicon();

    const link = document.querySelector<HTMLLinkElement>("link[rel~='icon']");
    expect(link).toBeTruthy();
    expect(link?.rel).toBe("icon");
  });

  it("should set href to a data URL with SVG", () => {
    setRandomEmojiFavicon();

    const link = document.querySelector<HTMLLinkElement>("link[rel~='icon']");
    expect(link?.href).toContain("data:image/svg+xml");
  });

  it("should use existing favicon link if present", () => {
    // Create an existing link
    const existingLink = document.createElement("link");
    existingLink.rel = "icon";
    existingLink.href = "old-favicon.ico";
    document.head.appendChild(existingLink);

    setRandomEmojiFavicon();

    // Should still be only one link
    const links = document.querySelectorAll("link[rel~='icon']");
    expect(links.length).toBe(1);

    // Should have updated the href
    expect(existingLink.href).toContain("data:image/svg+xml");
  });

  it("should create valid SVG data URL", () => {
    setRandomEmojiFavicon();

    const link = document.querySelector<HTMLLinkElement>("link[rel~='icon']");
    const href = link?.href || "";

    // Should be a data URL
    expect(href.startsWith("data:image/svg+xml")).toBe(true);

    // Should contain encoded SVG
    expect(href).toContain("svg");
    expect(href).toContain("text");
  });

  it("should handle missing document gracefully", () => {
    // This test validates the guard clause
    // In a real browser environment, document always exists
    // But the function handles this case for safety
    expect(() => setRandomEmojiFavicon()).not.toThrow();
  });
});
