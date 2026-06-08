import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { initDrawer } from "../lib/drawer";

const PARTIAL_HTML = '<button class="drawer-close">close</button>';

const flush = (): Promise<void> => new Promise((resolve) => setTimeout(resolve, 0));

function setupDom(extra = ""): HTMLElement {
  document.body.innerHTML = `
    <div id="drawer-backdrop" aria-hidden="true"></div>
    <div id="drawer-panel" aria-hidden="true"></div>
    ${extra}
  `;
  return document.getElementById("drawer-panel") as HTMLElement;
}

describe("initDrawer", () => {
  // initDrawer attaches listeners to document/window; capture them so each test
  // starts from a clean slate instead of accumulating handlers across tests.
  let docListeners: Array<[string, EventListenerOrEventListenerObject, unknown]>;
  let winListeners: Array<[string, EventListenerOrEventListenerObject, unknown]>;

  beforeEach(() => {
    docListeners = [];
    winListeners = [];

    const realDocAdd = document.addEventListener.bind(document);
    const realWinAdd = window.addEventListener.bind(window);
    vi.spyOn(document, "addEventListener").mockImplementation((type, handler, opts) => {
      docListeners.push([type as string, handler as EventListenerOrEventListenerObject, opts]);
      return realDocAdd(type as never, handler as never, opts as never);
    });
    vi.spyOn(window, "addEventListener").mockImplementation((type, handler, opts) => {
      winListeners.push([type as string, handler as EventListenerOrEventListenerObject, opts]);
      return realWinAdd(type as never, handler as never, opts as never);
    });

    vi.spyOn(history, "back").mockImplementation(() => {});
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      text: () => Promise.resolve(PARTIAL_HTML),
    });
  });

  afterEach(() => {
    docListeners.forEach(([t, h, o]) => document.removeEventListener(t, h, o as never));
    winListeners.forEach(([t, h, o]) => window.removeEventListener(t, h, o as never));
    vi.restoreAllMocks();
    document.body.className = "";
    document.body.innerHTML = "";
  });

  it("does nothing when panel or backdrop are missing", () => {
    document.body.innerHTML = `<div id="drawer-panel"></div>`; // backdrop missing

    expect(() => initDrawer()).not.toThrow();
    expect(docListeners).toHaveLength(0);
    expect(fetch).not.toHaveBeenCalled();
  });

  it("does not throw when prefetch fires with a non-Element target", () => {
    // Regression: pointerenter in the capture phase can have e.target === document,
    // which has no closest() — guard must short-circuit instead of throwing.
    setupDom();
    initDrawer();

    expect(() => document.dispatchEvent(new Event("pointerenter"))).not.toThrow();
    expect(fetch).not.toHaveBeenCalled();
  });

  it("does not throw when a click target is not an Element", () => {
    setupDom();
    initDrawer();

    expect(() => document.dispatchEvent(new Event("click"))).not.toThrow();
    expect(fetch).not.toHaveBeenCalled();
  });

  it("opens the drawer when an id card is clicked", async () => {
    const panel = setupDom(
      `<div class="id-grid"><a href="/id/42?page=2" class="id-card">card</a></div>`
    );
    initDrawer();

    const card = document.querySelector(".id-card") as HTMLElement;
    card.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    await flush();

    expect(fetch).toHaveBeenCalledWith("/id/42?partial=1&page=2");
    expect(document.body.classList.contains("drawer-open")).toBe(true);
    expect(panel.innerHTML).toContain("drawer-close");
  });

  it("reuses a hover prefetch for the subsequent click (dedupe)", async () => {
    setupDom(`<div class="id-grid"><a href="/id/7" class="id-card">card</a></div>`);
    initDrawer();

    const card = document.querySelector(".id-card") as HTMLElement;
    card.dispatchEvent(new Event("pointerenter"));
    await flush();
    card.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    await flush();

    expect(document.body.classList.contains("drawer-open")).toBe(true);
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch).toHaveBeenCalledWith("/id/7?partial=1");
  });

  it("opens a drawer-link as a partial without a card target", async () => {
    setupDom(`<a href="/werkzeug" class="drawer-link">werkzeug anfordern</a>`);
    initDrawer();

    const link = document.querySelector(".drawer-link") as HTMLElement;
    link.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    await flush();

    expect(fetch).toHaveBeenCalledWith("/werkzeug?partial=1");
    expect(document.body.classList.contains("drawer-open")).toBe(true);
  });

  it("closes an open drawer on Escape", async () => {
    setupDom(`<div class="id-grid"><a href="/id/1" class="id-card">card</a></div>`);
    initDrawer();

    const card = document.querySelector(".id-card") as HTMLElement;
    card.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    await flush();
    expect(document.body.classList.contains("drawer-open")).toBe(true);

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(document.body.classList.contains("drawer-open")).toBe(false);
  });

  it("closes the drawer when the backdrop is clicked", async () => {
    setupDom(`<div class="id-grid"><a href="/id/1" class="id-card">card</a></div>`);
    initDrawer();

    const card = document.querySelector(".id-card") as HTMLElement;
    card.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    await flush();
    expect(document.body.classList.contains("drawer-open")).toBe(true);

    document.getElementById("drawer-backdrop")!.dispatchEvent(new MouseEvent("click"));
    expect(document.body.classList.contains("drawer-open")).toBe(false);
  });
});
