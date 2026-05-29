/**
 * Drawer module
 * Handles drawer/modal functionality for viewing content
 */

import { initLazyImages } from "./lazy-images";

interface DrawerState {
  drawer: boolean;
  number: number;
  page: string | null;
  city: string | null;
}

// partial fetch URL, keeping the active page/city filters
function buildPartialUrl(num: string | number, page: string | null, city: string | null): string {
  const qs = new URLSearchParams();
  qs.set("partial", "1");
  if (page) qs.set("page", page);
  if (city) qs.set("city", city);
  return `/id/${num}?${qs.toString()}`;
}

// history URL pushed when a drawer opens (params encoded, no raw spaces)
function buildHistoryUrl(num: string | number, page: string | null, city: string | null): string {
  const qs = new URLSearchParams();
  if (page) qs.set("page", page);
  if (city) qs.set("city", city);
  const search = qs.toString();
  return search ? `/id/${num}?${search}` : `/id/${num}`;
}

export function initDrawer(): void {
  const panel = document.getElementById("drawer-panel") as HTMLElement;
  const backdrop = document.getElementById("drawer-backdrop") as HTMLElement;

  if (!panel || !backdrop) return;

  // focus is restored here on close
  let lastFocused: HTMLElement | null = null;

  // dedupe fetches so a hover-prefetch is reused on click
  const partialCache = new Map<string, Promise<string>>();

  function fetchPartial(url: string): Promise<string> {
    let pending = partialCache.get(url);
    if (!pending) {
      pending = fetch(url).then((r) => {
        if (!r.ok) throw new Error("fetch failed");
        return r.text();
      });
      pending.catch(() => partialCache.delete(url)); // allow retry after failure
      partialCache.set(url, pending);
    }
    return pending;
  }

  function isOpen(): boolean {
    return document.body.classList.contains("drawer-open");
  }

  function closeDrawer(pushBack: boolean): void {
    if (!isOpen()) return;
    const pushed = panel.dataset.drawerPushed === "true";
    document.body.classList.remove("drawer-open");
    panel.innerHTML = "";
    panel.setAttribute("aria-hidden", "true");
    backdrop.setAttribute("aria-hidden", "true");
    // only navigate back if caller requested AND this drawer actually pushed a history entry
    if (pushBack && pushed) history.back();
    // clear stored flag
    delete panel.dataset.drawerPushed;
    // restore focus to the opener
    if (lastFocused && document.contains(lastFocused)) {
      lastFocused.focus();
    }
    lastFocused = null;
  }

  function openDrawer(
    number: number,
    html: string,
    pushState: boolean,
    pageParam: string | null,
    cityParam: string | null,
    trigger?: HTMLElement | null
  ): void {
    if (trigger) lastFocused = trigger;
    panel.innerHTML = html;
    panel.setAttribute("aria-hidden", "false");
    backdrop.setAttribute("aria-hidden", "false");
    document.body.classList.add("drawer-open");

    // Wire close button if present
    const closeBtn = panel.querySelector<HTMLElement>(".drawer-close");
    if (closeBtn) {
      closeBtn.addEventListener("click", () => closeDrawer(true), {
        once: true,
      });
    }

    // Wire back link to close drawer instead of normal navigation
    const backLink = panel.querySelector<HTMLAnchorElement>(".drawer-back-link");
    if (backLink) {
      backLink.addEventListener(
        "click",
        (e) => {
          e.preventDefault();
          closeDrawer(true);
        },
        {
          once: true,
        }
      );
    }

    // lazy-load panel images with the panel as scroll root, so only visible
    // contributions are fetched instead of all at once
    try {
      initLazyImages(panel, panel);
    } catch (e) {
      console.warn("initLazyImages failed", e);
    }

    if (pushState) {
      history.pushState(
        { drawer: true, number, page: pageParam, city: cityParam } as DrawerState,
        "",
        buildHistoryUrl(number, pageParam, cityParam)
      );
      panel.dataset.drawerPushed = "true";
    } else {
      // record that we did not push history for this drawer instance
      panel.dataset.drawerPushed = "false";
    }

    (closeBtn || panel).focus();
  }

  backdrop.addEventListener("click", () => closeDrawer(true));

  // Close on Escape
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && isOpen()) {
      e.preventDefault();
      closeDrawer(true);
    }
  });

  // parse num/page/city from a card's href, or null if not a card link
  function parseCard(
    el: HTMLElement
  ): { num: string; page: string | null; city: string | null } | null {
    const href =
      el.getAttribute("href") || el.querySelector<HTMLAnchorElement>("a")?.getAttribute("href");
    if (!href) return null;
    const m = href.match(/\/id\/(\d+)/);
    if (!m) return null;
    const url = new URL(href, window.location.origin);
    return {
      num: m[1],
      page: url.searchParams.get("page"),
      city: url.searchParams.get("city"),
    };
  }

  // prefetch the partial on hover/touch for instant open
  const prefetchCard = (e: Event): void => {
    const card = (e.target as HTMLElement).closest<HTMLElement>(".id-card");
    if (!card) return;
    const info = parseCard(card);
    if (!info) return;
    fetchPartial(buildPartialUrl(info.num, info.page, info.city)).catch(() => {});
  };
  document.addEventListener("pointerenter", prefetchCard, true);
  document.addEventListener("touchstart", prefetchCard, { passive: true });

  // Click delegation for id cards
  document.addEventListener("click", (e) => {
    const card = (e.target as HTMLElement).closest<HTMLElement>(".id-card");
    if (!card) return;
    const info = parseCard(card);
    if (!info) return;
    e.preventDefault();

    const href =
      card.getAttribute("href") ||
      card.querySelector<HTMLAnchorElement>("a")?.getAttribute("href") ||
      buildHistoryUrl(info.num, info.page, info.city);

    fetchPartial(buildPartialUrl(info.num, info.page, info.city))
      .then((html) => openDrawer(parseInt(info.num), html, true, info.page, info.city, card))
      .catch((err) => {
        console.error(err);
        window.location.href = href;
      });
  });

  // Click delegation for simple drawer links (e.g., "werkzeug anfordern")
  document.addEventListener("click", (e) => {
    const link = (e.target as HTMLElement).closest<HTMLAnchorElement>(".drawer-link");
    if (!link) return;
    e.preventDefault();
    const href = link.getAttribute("href");
    if (!href) return;
    const fetchUrl = href + (href.includes("?") ? "&partial=1" : "?partial=1");
    fetchPartial(fetchUrl)
      .then((html) => openDrawer(0, html, false, null, null, link))
      .catch((err) => {
        console.error(err);
        window.location.href = href;
      });
  });

  // popstate: open/close/update drawer to match the new location
  window.addEventListener("popstate", () => {
    const m = location.pathname.match(/\/id\/(\d+)/);
    if (m) {
      // only hijack navigation when there is a grid to drawer over
      if (!isOpen() && !document.querySelector(".id-grid")) return;
      const num = m[1];
      const params = new URLSearchParams(location.search);
      const page = params.get("page");
      const city = params.get("city");
      fetchPartial(buildPartialUrl(num, page, city))
        .then((html) => openDrawer(parseInt(num), html, false, page, city))
        .catch((err) => console.error(err));
    } else {
      closeDrawer(false);
    }
  });
}
