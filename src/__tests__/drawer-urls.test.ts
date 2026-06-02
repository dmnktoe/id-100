/**
 * Tests for drawer URL helpers
 */
import { describe, it, expect } from "vitest";
import {
  buildPartialUrl,
  buildHistoryUrl,
  parseDrawerHref,
  parseDrawerLocation,
} from "../lib/drawer-urls";

const ORIGIN = "https://example.com";

describe("buildPartialUrl", () => {
  it("always requests the partial fragment", () => {
    expect(buildPartialUrl({ number: "5", page: null, city: null })).toBe("/id/5?partial=1");
  });

  it("preserves page and city filters", () => {
    expect(buildPartialUrl({ number: "5", page: "2", city: "Berlin" })).toBe(
      "/id/5?partial=1&page=2&city=Berlin"
    );
  });

  it("encodes city values so spaces never leak raw", () => {
    const url = buildPartialUrl({ number: "5", page: null, city: "Frankfurt am Main" });
    expect(url).not.toMatch(/ /);
    expect(url).toContain("city=Frankfurt+am+Main");
  });
});

describe("buildHistoryUrl", () => {
  it("returns a bare id path without filters", () => {
    expect(buildHistoryUrl({ number: "7", page: null, city: null })).toBe("/id/7");
  });

  it("includes only the present filters", () => {
    expect(buildHistoryUrl({ number: "7", page: "3", city: null })).toBe("/id/7?page=3");
    expect(buildHistoryUrl({ number: "7", page: null, city: "Köln" })).toBe("/id/7?city=K%C3%B6ln");
  });

  it("encodes city values so spaces never leak raw", () => {
    const url = buildHistoryUrl({ number: "7", page: "1", city: "Frankfurt am Main" });
    expect(url).not.toMatch(/ /);
    expect(url).toBe("/id/7?page=1&city=Frankfurt+am+Main");
  });
});

describe("parseDrawerHref", () => {
  it("returns null for non-id hrefs", () => {
    expect(parseDrawerHref("/about", ORIGIN)).toBeNull();
  });

  it("extracts number, page and city", () => {
    expect(parseDrawerHref("/id/12?page=4&city=Berlin", ORIGIN)).toEqual({
      number: "12",
      page: "4",
      city: "Berlin",
    });
  });

  it("decodes encoded city values", () => {
    expect(parseDrawerHref("/id/12?city=Frankfurt%20am%20Main", ORIGIN)?.city).toBe(
      "Frankfurt am Main"
    );
  });

  it("round-trips with buildHistoryUrl", () => {
    const target = { number: "12", page: "4", city: "Frankfurt am Main" };
    expect(parseDrawerHref(buildHistoryUrl(target), ORIGIN)).toEqual(target);
  });
});

describe("parseDrawerLocation", () => {
  it("returns null off id routes", () => {
    expect(parseDrawerLocation({ pathname: "/", search: "" })).toBeNull();
  });

  it("reads number and filters from a location", () => {
    expect(parseDrawerLocation({ pathname: "/id/3", search: "?page=2&city=Berlin" })).toEqual({
      number: "3",
      page: "2",
      city: "Berlin",
    });
  });
});
