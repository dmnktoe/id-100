/**
 * Drawer module
 * Handles drawer/modal functionality for viewing content
 */

import { initLazyImages } from "./lazy-images";

interface DrawerState {
  drawer: boolean;
  number: number;
  page: string | null;
}

export function initDrawer(): void {
  const panel = document.getElementById("drawer-panel") as HTMLElement;
  const backdrop = document.getElementById("drawer-backdrop") as HTMLElement;

  if (!panel || !backdrop) return;

  function closeDrawer(pushBack: boolean): void {
    const pushed = panel.dataset.drawerPushed === "true";
    document.body.classList.remove("drawer-open");
    panel.innerHTML = "";
    panel.setAttribute("aria-hidden", "true");
    backdrop.setAttribute("aria-hidden", "true");
    // only navigate back if caller requested AND this drawer actually pushed a history entry
    if (pushBack && pushed) history.back();
    // clear stored flag
    delete panel.dataset.drawerPushed;
  }

  function openDrawer(
    number: number,
    html: string,
    pushState: boolean,
    pageParam: string | null
  ): void {
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

    // Initialize lazy images inside the newly-inserted panel and force immediate load
    try {
      initLazyImages(panel);
      Array.from(panel.querySelectorAll<HTMLImageElement>("img.lazy")).forEach((img) => {
        const full = img.getAttribute("data-src");
        if (full && img.src !== full) img.src = full;
      });
    } catch (e) {
      console.warn("initLazyImages failed", e);
    }

    if (pushState) {
      const url = pageParam
        ? `/id/${number}?page=${pageParam}`
        : `/id/${number}`;
      history.pushState(
        { drawer: true, number: number, page: pageParam } as DrawerState,
        "",
        url
      );
      panel.dataset.drawerPushed = "true";
    } else {
      // record that we did not push history for this drawer instance
      panel.dataset.drawerPushed = "false";
    }
  }

  backdrop.addEventListener("click", () => closeDrawer(true));

  // Click delegation for id cards
  document.addEventListener("click", (e) => {
    const card = (e.target as HTMLElement).closest<HTMLElement>(".id-card");
    if (!card) return;
    e.preventDefault();
    const href =
      card.getAttribute("href") ||
      card.querySelector<HTMLAnchorElement>("a")?.getAttribute("href");
    if (!href) return;

    // extract number and page parameter from href /id/:number?page=X
    const m = href.match(/\/id\/(\d+)/);
    if (!m) return;
    const num = m[1];

    // Extract page parameter if present
    const url = new URL(href, window.location.origin);
    const pageParam = url.searchParams.get("page");

    const fetchUrl = pageParam
      ? `/id/${num}?partial=1&page=${pageParam}`
      : `/id/${num}?partial=1`;

    fetch(fetchUrl)
      .then((r) => {
        if (!r.ok) throw new Error("fetch failed");
        return r.text();
      })
      .then((html) => openDrawer(parseInt(num), html, true, pageParam))
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
    fetch(fetchUrl)
      .then((r) => {
        if (!r.ok) throw new Error("fetch failed");
        return r.text();
      })
      .then((html) => openDrawer(0, html, false, null))
      .catch((err) => {
        console.error(err);
        window.location.href = href;
      });
  });

  // popstate: close drawer when state removed
  window.addEventListener("popstate", (_ev) => {
    if (document.body.classList.contains("drawer-open")) {
      // if new location is an id path keep it open, otherwise close
      const path = location.pathname;
      const m = path.match(/\/id\/(\d+)/);
      if (m) {
        const num = m[1];
        const pageParam = new URLSearchParams(location.search).get("page");
        const fetchUrl = pageParam
          ? `/id/${num}?partial=1&page=${pageParam}`
          : `/id/${num}?partial=1`;
        // fetch and open if different
        fetch(fetchUrl)
          .then((r) => r.text())
          .then((html) => openDrawer(parseInt(num), html, false, pageParam));
      } else {
        closeDrawer(false);
      }
    } else {
      // if not open but user navigated directly to id path (e.g. via back), and page has .id-grid, open it
      const m = location.pathname.match(/\/id\/(\d+)/);
      if (m && document.querySelector(".id-grid")) {
        const num = m[1];
        const pageParam = new URLSearchParams(location.search).get("page");
        const fetchUrl = pageParam
          ? `/id/${num}?partial=1&page=${pageParam}`
          : `/id/${num}?partial=1`;
        fetch(fetchUrl)
          .then((r) => r.text())
          .then((html) => openDrawer(parseInt(num), html, false, pageParam));
      }
    }
  });
}
