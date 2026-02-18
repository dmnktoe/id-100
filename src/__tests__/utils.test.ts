/**
 * Tests for utils module
 */
import { describe, it, expect } from "vitest";
import { getErrorMessage } from "../lib/utils";

describe("getErrorMessage", () => {
  it("should extract message from Error instance", () => {
    const error = new Error("Test error message");
    expect(getErrorMessage(error)).toBe("Test error message");
  });

  it("should return default message for non-Error types", () => {
    expect(getErrorMessage("string error")).toBe("Unbekannter Fehler");
    expect(getErrorMessage(123)).toBe("Unbekannter Fehler");
    expect(getErrorMessage(null)).toBe("Unbekannter Fehler");
    expect(getErrorMessage(undefined)).toBe("Unbekannter Fehler");
    expect(getErrorMessage({ message: "not an Error" })).toBe("Unbekannter Fehler");
  });
});
