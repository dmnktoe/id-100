/**
 * Drawer module — opens id detail pages in an overlay without a full navigation,
 * keeping the URL (page/city filters) in sync for deep-linking and back/forward.
 */

import { initLazyImages } from "./lazy-images";
import {
  buildHistoryUrl,
  buildPartialUrl,
  parseDrawerHref,
  parseDrawerLocation,
  type DrawerTarget,
} from "./drawer-urls";

/** Dedupes partial fetches so a hover-prefetch can be reused by the subsequent click. */
function createPartialFetcher(): (url: string) => Promise<string> {
  const cache = new Map<string, Promise<string>>();
  return (url) => {
    let pending = cache.get(url);
    if (!pending) {
      pending = fetch(url).then((res) => {
        if (!res.ok) throw new Error(`drawer partial failed: ${res.status}`);
        return res.text();
      });
      pending.catch(() => cache.delete(url)); // allow retry after a failed request
      cache.set(url, pending);
    }
    return pending;
  };
}

/** Resolve the drawer target for an id card from its href (or nested anchor). */
function targetFromCard(card: HTMLElement): DrawerTarget | null {
  const href =
    card.getAttribute("href") ?? card.querySelector<HTMLAnchorElement>("a")?.getAttribute("href");
  return href ? parseDrawerHref(href) : null;
}

/** Owns the panel DOM: open/close lifecycle, focus handling, images and history. */
class DrawerController {
  private lastFocused: HTMLElement | null = null;

  constructor(
    private readonly panel: HTMLElement,
    private readonly backdrop: HTMLElement
  ) {}

  get isOpen(): boolean {
    return document.body.classList.contains("drawer-open");
  }

  open(
    html: string,
    target: DrawerTarget | null,
    pushState: boolean,
    trigger?: HTMLElement | null
  ): void {
    if (trigger) this.lastFocused = trigger;

    this.panel.innerHTML = html;
    this.panel.setAttribute("aria-hidden", "false");
    this.backdrop.setAttribute("aria-hidden", "false");
    document.body.classList.add("drawer-open");

    this.wireDismissControls();
    this.loadImages();
    this.syncHistory(target, pushState);

    (this.panel.querySelector<HTMLElement>(".drawer-close") ?? this.panel).focus();
  }

  close(pushBack: boolean): void {
    if (!this.isOpen) return;
    const pushed = this.panel.dataset.drawerPushed === "true";

    document.body.classList.remove("drawer-open");
    this.panel.innerHTML = "";
    this.panel.setAttribute("aria-hidden", "true");
    this.backdrop.setAttribute("aria-hidden", "true");
    delete this.panel.dataset.drawerPushed;

    // only navigate back when this instance actually pushed a history entry
    if (pushBack && pushed) history.back();
    this.restoreFocus();
  }

  private wireDismissControls(): void {
    const dismiss = (e: Event): void => {
      e.preventDefault();
      this.close(true);
    };
    this.panel
      .querySelector<HTMLElement>(".drawer-close")
      ?.addEventListener("click", dismiss, { once: true });
    this.panel
      .querySelector<HTMLAnchorElement>(".drawer-back-link")
      ?.addEventListener("click", dismiss, { once: true });
  }

  private loadImages(): void {
    // panel as scroll root → only contributions in view are fetched, not all at once
    try {
      initLazyImages(this.panel, this.panel);
    } catch (e) {
      console.warn("initLazyImages failed", e);
    }
  }

  private syncHistory(target: DrawerTarget | null, pushState: boolean): void {
    if (pushState && target) {
      history.pushState({ drawer: true, target }, "", buildHistoryUrl(target));
      this.panel.dataset.drawerPushed = "true";
    } else {
      this.panel.dataset.drawerPushed = "false";
    }
  }

  private restoreFocus(): void {
    if (this.lastFocused && document.contains(this.lastFocused)) {
      this.lastFocused.focus();
    }
    this.lastFocused = null;
  }
}

export function initDrawer(): void {
  const panel = document.getElementById("drawer-panel");
  const backdrop = document.getElementById("drawer-backdrop");
  if (!panel || !backdrop) return;

  const drawer = new DrawerController(panel, backdrop);
  const fetchPartial = createPartialFetcher();

  const openTarget = (
    target: DrawerTarget,
    pushState: boolean,
    trigger?: HTMLElement | null
  ): Promise<void> =>
    fetchPartial(buildPartialUrl(target)).then((html) =>
      drawer.open(html, target, pushState, trigger)
    );

  // dismiss: backdrop click + Escape
  backdrop.addEventListener("click", () => drawer.close(true));
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && drawer.isOpen) {
      e.preventDefault();
      drawer.close(true);
    }
  });

  // prefetch on hover/touch for instant open (pointerenter doesn't bubble → capture)
  const prefetch = (e: Event): void => {
    const card = (e.target as HTMLElement).closest<HTMLElement>(".id-card");
    const target = card && targetFromCard(card);
    if (target) fetchPartial(buildPartialUrl(target)).catch(() => {});
  };
  document.addEventListener("pointerenter", prefetch, true);
  document.addEventListener("touchstart", prefetch, { passive: true });

  // open id cards in the drawer
  document.addEventListener("click", (e) => {
    const card = (e.target as HTMLElement).closest<HTMLElement>(".id-card");
    if (!card) return;
    const target = targetFromCard(card);
    if (!target) return;
    e.preventDefault();
    openTarget(target, true, card).catch((err) => {
      console.error(err);
      window.location.href = buildHistoryUrl(target);
    });
  });

  // open simple drawer links (e.g. "werkzeug anfordern") without a history entry
  document.addEventListener("click", (e) => {
    const link = (e.target as HTMLElement).closest<HTMLAnchorElement>(".drawer-link");
    if (!link) return;
    const href = link.getAttribute("href");
    if (!href) return;
    e.preventDefault();
    const url = href + (href.includes("?") ? "&partial=1" : "?partial=1");
    fetchPartial(url)
      .then((html) => drawer.open(html, null, false, link))
      .catch((err) => {
        console.error(err);
        window.location.href = href;
      });
  });

  // sync drawer with browser back/forward
  window.addEventListener("popstate", () => {
    const target = parseDrawerLocation(window.location);
    if (!target) {
      drawer.close(false);
      return;
    }
    // only hijack navigation when there is a grid to drawer over
    if (!drawer.isOpen && !document.querySelector(".id-grid")) return;
    openTarget(target, false).catch((err) => console.error(err));
  });
}
