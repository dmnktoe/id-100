/**
 * Lazy image loading module
 * Implements blur-up image loading with IntersectionObserver
 */

export function initLazyImages(root?: Document | HTMLElement, observerRoot?: Element | null): void {
  const container = root || document;
  const images = Array.from(container.querySelectorAll<HTMLImageElement>("img.lazy"));

  if (images.length === 0) return;

  const setPlaceholder = (img: HTMLImageElement): void => {
    const lqip = img.getAttribute("data-lqip");
    if (lqip) {
      // only set when placeholder differs from current src
      if (!img.getAttribute("src") || img.getAttribute("src")?.[0] === "#") {
        img.src = lqip;
      }
    }
  };

  images.forEach((img) => {
    // Skip if already handled
    if (img.dataset.lazyInitialized) return;
    img.dataset.lazyInitialized = "1";

    // attach handlers BEFORE setting placeholder so we can react to subsequent loads
    const onLoad = (): void => {
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

    const onError = (): void => {
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
      if (img.naturalWidth && img.naturalWidth > 0) {
        onLoad();
      } else {
        const full = img.getAttribute("data-src");
        const currentSrc = img.currentSrc || img.src || "";
        if (!full || currentSrc === full) {
          onError();
        }
      }
    }
  });

  // Preload the full-size image off-DOM and only swap the visible src once it
  // is fully loaded (and decoded). Swapping src directly would discard the
  // currently displayed LQIP and show a blank/white frame until the full image
  // arrives — preloading keeps the blurred placeholder visible until the moment
  // the sharp image is ready to paint.
  const loadFull = (img: HTMLImageElement): void => {
    const src = img.getAttribute("data-src");
    if (!src || img.src === src || img.dataset.fullLoaded) return;
    img.dataset.fullLoaded = "1";

    const reveal = (): void => {
      if (img.src !== src) img.src = src;
      img.classList.add("loaded");
    };

    const pre = new Image();
    pre.onload = (): void => {
      // decode (when available) so the swap paints instantly without a flash
      if (typeof pre.decode === "function") {
        pre.decode().then(reveal).catch(reveal);
      } else {
        reveal();
      }
    };
    pre.onerror = (): void => {
      // keep the placeholder but drop the blur so we don't stay blurred forever
      img.classList.add("loaded");
    };
    pre.src = src;
  };

  if ("IntersectionObserver" in window) {
    const obs = new IntersectionObserver(
      (entries, o) => {
        entries.forEach((en) => {
          if (!en.isIntersecting) return;
          const img = en.target as HTMLImageElement;
          // preload full-size, then swap once ready (no blank frame)
          loadFull(img);
          o.unobserve(img);
        });
      },
      { root: observerRoot ?? null, rootMargin: "200px", threshold: 0.01 }
    );

    images.forEach((img) => {
      if (img.classList.contains("loaded")) return; // already loaded
      obs.observe(img);
    });
  } else {
    // fallback: eager load immediately
    images.forEach((img) => {
      const src = img.getAttribute("data-src");
      if (src && !img.classList.contains("loaded") && img.src !== src) {
        img.src = src;
      }
    });
  }
}
