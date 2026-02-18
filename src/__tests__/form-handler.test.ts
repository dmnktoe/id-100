/**
 * Tests for form-handler module
 */
import { describe, it, expect, beforeEach, vi } from "vitest";
import { initFormHandlers } from "../lib/form-handler";

describe("initFormHandlers", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    global.fetch = vi.fn();
  });

  it("should not process non-requestBagForm forms", () => {
    document.body.innerHTML = `
      <form id="otherForm">
        <input name="email" value="test@example.com" />
        <button type="submit">Submit</button>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("otherForm") as HTMLFormElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    expect(() => form.dispatchEvent(event)).not.toThrow();
  });

  it("should validate email before submission", () => {
    document.body.innerHTML = `
      <form id="requestBagForm">
        <input name="email" value="invalid-email" />
        <button type="submit">Submit</button>
        <div id="requestResult"></div>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("requestBagForm") as HTMLFormElement;
    const resultDiv = document.getElementById("requestResult") as HTMLDivElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    form.dispatchEvent(event);

    // Should show error message
    expect(resultDiv.style.display).toBe("block");
    expect(resultDiv.style.color).toBe("#d32f2f");
    expect(resultDiv.innerText).toContain("gültige E‑Mail");

    // Should not have called fetch
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it("should validate empty email", () => {
    document.body.innerHTML = `
      <form id="requestBagForm">
        <input name="email" value="  " />
        <button type="submit">Submit</button>
        <div id="requestResult"></div>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("requestBagForm") as HTMLFormElement;
    const resultDiv = document.getElementById("requestResult") as HTMLDivElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    form.dispatchEvent(event);

    expect(resultDiv.innerText).toContain("gültige E‑Mail");
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it("should submit valid email", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ status: "ok" }),
    });
    global.fetch = mockFetch;

    document.body.innerHTML = `
      <form id="requestBagForm">
        <input name="email" value="test@example.com" />
        <button type="submit">anfragen</button>
        <div id="requestResult"></div>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("requestBagForm") as HTMLFormElement;
    const btn = form.querySelector("button[type=submit]") as HTMLButtonElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    form.dispatchEvent(event);

    // Should disable button and show loading state
    expect(btn.disabled).toBe(true);
    expect(btn.innerText).toBe("sende...");

    // Should call fetch
    expect(mockFetch).toHaveBeenCalledWith("/werkzeug-anfordern", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: "test@example.com" }),
    });

    // Wait for async operations
    await vi.waitFor(() => {
      expect(form.innerHTML).toContain("Danke!");
    });
  });

  it("should handle server error response", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ status: "error", error: "Server error" }),
    });
    global.fetch = mockFetch;

    document.body.innerHTML = `
      <form id="requestBagForm">
        <input name="email" value="test@example.com" />
        <button type="submit">anfragen</button>
        <div id="requestResult"></div>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("requestBagForm") as HTMLFormElement;
    const btn = form.querySelector("button[type=submit]") as HTMLButtonElement;
    const resultDiv = document.getElementById("requestResult") as HTMLDivElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    form.dispatchEvent(event);

    await vi.waitFor(() => {
      expect(resultDiv.style.display).toBe("block");
      expect(resultDiv.innerText).toBe("Server error");
      expect(btn.disabled).toBe(false);
      expect(btn.innerText).toBe("anfragen");
    });
  });

  it("should handle network error", async () => {
    const mockFetch = vi.fn().mockRejectedValue(new Error("Network error"));
    global.fetch = mockFetch;

    document.body.innerHTML = `
      <form id="requestBagForm">
        <input name="email" value="test@example.com" />
        <button type="submit">anfragen</button>
        <div id="requestResult"></div>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("requestBagForm") as HTMLFormElement;
    const btn = form.querySelector("button[type=submit]") as HTMLButtonElement;
    const resultDiv = document.getElementById("requestResult") as HTMLDivElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    form.dispatchEvent(event);

    await vi.waitFor(() => {
      expect(resultDiv.style.display).toBe("block");
      expect(resultDiv.innerText).toContain("Netzwerkfehler");
      expect(btn.disabled).toBe(false);
      expect(btn.innerText).toBe("anfragen");
    });
  });

  it("should handle missing form elements gracefully", () => {
    document.body.innerHTML = `
      <form id="requestBagForm">
        <button type="submit">Submit</button>
      </form>
    `;

    initFormHandlers();

    const form = document.getElementById("requestBagForm") as HTMLFormElement;
    const event = new Event("submit", { bubbles: true, cancelable: true });

    // Should not throw when elements are missing
    expect(() => form.dispatchEvent(event)).not.toThrow();
  });
});
