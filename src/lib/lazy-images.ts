/**
 * Lazy image loading module
 * Implements blur-up image loading with IntersectionObserver
 */

export function initLazyImages(root?: Document | HTMLElement): void {
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
        onError();
      }
    }
  });

  if ("IntersectionObserver" in window) {
    const obs = new IntersectionObserver(
      (entries, o) => {
        entries.forEach((en) => {
          if (!en.isIntersecting) return;
          const img = en.target as HTMLImageElement;
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
      if (src && !img.classList.contains("loaded") && img.src !== src) {
        img.src = src;
      }
    });
  }
}
