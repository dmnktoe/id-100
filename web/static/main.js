(function () {
  // Brand animation logic: animate on subpages, reverse on homepage return
  const brandFull = document.querySelector(".brand-full");
  const brandCompact = document.querySelector(".brand-compact");

  // If on homepage and returning from a subpage, trigger reverse animation
  if (brandFull && sessionStorage.getItem("brandAnimated") === "1") {
    brandFull.classList.add("reverse-animated");
    sessionStorage.setItem("brandAnimated", "0");

    // Remove animation class after animation completes to restore normal state
    setTimeout(() => {
      brandFull.classList.remove("reverse-animated");
    }, 800);
  }

  // If on subpage and not yet animated, trigger forward animation
  if (brandCompact && sessionStorage.getItem("brandAnimated") !== "1") {
    brandCompact.classList.add("animated");
    sessionStorage.setItem("brandAnimated", "1");
  }

  // Drawer logic (global)
  const panel = document.getElementById("drawer-panel");
  const backdrop = document.getElementById("drawer-backdrop");

  function closeDrawer(pushBack) {
    document.body.classList.remove("drawer-open");
    panel.innerHTML = "";
    panel.setAttribute("aria-hidden", "true");
    backdrop.setAttribute("aria-hidden", "true");
    if (pushBack) history.back();
  }

  function openDrawer(number, html, pushState, pageParam) {
    panel.innerHTML = html;
    panel.setAttribute("aria-hidden", "false");
    backdrop.setAttribute("aria-hidden", "false");
    document.body.classList.add("drawer-open");

    // Wire close button if present
    const closeBtn = panel.querySelector(".drawer-close");
    if (closeBtn)
      closeBtn.addEventListener("click", () => closeDrawer(true), {
        once: true,
      });

    // Wire back link to close drawer instead of normal navigation
    const backLink = panel.querySelector(".drawer-back-link");
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

    // intercept back/links inside panel
    panel.querySelectorAll("a").forEach((a) => {
      a.addEventListener("click", (e) => {
        // let normal navigation happen if link to full page
      });
    });

    // Initialize lazy images inside the newly-inserted panel and force immediate load
    try {
      initLazyImages(panel);
      Array.from(panel.querySelectorAll("img.lazy")).forEach((img) => {
        const full = img.getAttribute("data-src");
        if (full && img.src !== full) img.src = full;
      });
    } catch (e) {
      console.warn("initLazyImages failed", e);
    }

    if (pushState) {
      const url = pageParam
        ? `/derive/${number}?page=${pageParam}`
        : `/derive/${number}`;
      history.pushState(
        { drawer: true, number: number, page: pageParam },
        "",
        url
      );
    }
  }

  backdrop.addEventListener("click", () => closeDrawer(true));

  // Click delegation for derive cards
  document.addEventListener("click", (e) => {
    const card = e.target.closest(".derive-card");
    if (!card) return;
    e.preventDefault();
    const href =
      card.getAttribute("href") ||
      card.querySelector("a")?.getAttribute("href");
    if (!href) return;

    // extract number and page parameter from href /derive/:number?page=X
    const m = href.match(/\/derive\/(\d+)/);
    if (!m) return;
    const num = m[1];

    // Extract page parameter if present
    const url = new URL(href, window.location.origin);
    const pageParam = url.searchParams.get("page");

    const fetchUrl = pageParam
      ? `/derive/${num}?partial=1&page=${pageParam}`
      : `/derive/${num}?partial=1`;

    fetch(fetchUrl)
      .then((r) => {
        if (!r.ok) throw new Error("fetch failed");
        return r.text();
      })
      .then((html) => openDrawer(num, html, true, pageParam))
      .catch((err) => {
        console.error(err);
        window.location.href = href;
      });
  });

  // popstate: close drawer when state removed
  window.addEventListener("popstate", (ev) => {
    if (document.body.classList.contains("drawer-open")) {
      // if new location is a derive path keep it open, otherwise close
      const path = location.pathname;
      const m = path.match(/\/derive\/(\d+)/);
      if (m) {
        const num = m[1];
        const pageParam = new URLSearchParams(location.search).get("page");
        const fetchUrl = pageParam
          ? `/derive/${num}?partial=1&page=${pageParam}`
          : `/derive/${num}?partial=1`;
        // fetch and open if different
        fetch(fetchUrl)
          .then((r) => r.text())
          .then((html) => openDrawer(num, html, false, pageParam));
      } else {
        closeDrawer(false);
      }
    } else {
      // if not open but user navigated directly to derive path (e.g. via back), and page has .derive-grid, open it
      const m = location.pathname.match(/\/derive\/(\d+)/);
      if (m && document.querySelector(".derive-grid")) {
        const num = m[1];
        const pageParam = new URLSearchParams(location.search).get("page");
        const fetchUrl = pageParam
          ? `/derive/${num}?partial=1&page=${pageParam}`
          : `/derive/${num}?partial=1`;
        fetch(fetchUrl)
          .then((r) => r.text())
          .then((html) => openDrawer(num, html, false, pageParam));
      }
    }
  });

  // Lazy blur-up image initializer
  function initLazyImages(root) {
    root = root || document;
    const images = Array.from(root.querySelectorAll("img.lazy"));
    if (images.length === 0) return;

    const setPlaceholder = (img) => {
      const lqip = img.getAttribute("data-lqip");
      if (lqip) {
        // only set when placeholder differs from current src
        if (!img.getAttribute("src") || img.getAttribute("src")[0] === "#")
          img.src = lqip;
      }
    };

    images.forEach((img) => {
      // Skip if already handled
      if (img.dataset.lazyInitialized) return;
      img.dataset.lazyInitialized = "1";

      // attach handlers BEFORE setting placeholder so we can react to subsequent loads
      const onLoad = () => {
        // Only remove blur when the full-size image has loaded (or there is no data-src)
        const full = img.getAttribute("data-src");
        try {
          const currentSrc = img.currentSrc || img.src || "";
          if (!full || currentSrc === full) {
            img.classList.add("loaded");
          }
        } catch (e) {
          img.classList.add("loaded");
        }
      };
      const onError = () => {
        // show placeholder even on error
        img.classList.add("loaded");
      };

      img.addEventListener("load", onLoad);
      img.addEventListener("error", onError);

      // Prefer to have a tiny placeholder from data-lqip (do not clear blur)
      setPlaceholder(img);

      // If already complete (cached), invoke handler to possibly set 'loaded'
      if (img.complete) {
        // call onLoad or onError depending on whether naturalWidth is present
        if (img.naturalWidth && img.naturalWidth > 0) onLoad();
        else onError();
      }
    });

    if ("IntersectionObserver" in window) {
      const obs = new IntersectionObserver(
        (entries, o) => {
          entries.forEach((en) => {
            if (!en.isIntersecting) return;
            const img = en.target;
            const src = img.getAttribute("data-src");
            if (src && img.src !== src) {
              // trigger full-size load
              img.src = src;
            }
            o.unobserve(img);
          });
        },
        { root: null, rootMargin: "200px", threshold: 0.01 }
      );

      images.forEach((img) => {
        if (img.classList.contains("loaded")) return; // already loaded
        obs.observe(img);
      });
    } else {
      // fallback: eager load immediately
      images.forEach((img) => {
        const src = img.getAttribute("data-src");
        if (src && !img.classList.contains("loaded") && img.src !== src)
          img.src = src;
      });
    }
  }

  // initialize lazy images on first paint
  initLazyImages();
})();
