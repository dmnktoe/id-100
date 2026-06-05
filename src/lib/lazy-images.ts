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
      if (!img.getAttribute("src") || img.getAttribute("src")?.[0] === "#") {
        img.src = lqip;
      }
    }
  };

  images.forEach((img) => {
    if (img.dataset.lazyInitialized) return;
    img.dataset.lazyInitialized = "1";

    const onLoad = (): void => {
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
      img.classList.add("loaded");
    };

    img.addEventListener("load", onLoad);
    img.addEventListener("error", onError);

    setPlaceholder(img);

    if (img.complete) {
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

  // Preload off-DOM and swap the visible src only once the full image is ready,
  // otherwise the browser drops the LQIP and shows a white frame mid-load.
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
      if (typeof pre.decode === "function") {
        pre.decode().then(reveal).catch(reveal);
      } else {
        reveal();
      }
    };
    pre.onerror = (): void => {
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
          loadFull(img);
          o.unobserve(img);
        });
      },
      { root: observerRoot ?? null, rootMargin: "200px", threshold: 0.01 }
    );

    images.forEach((img) => {
      if (img.classList.contains("loaded")) return;
      obs.observe(img);
    });
  } else {
    images.forEach((img) => {
      const src = img.getAttribute("data-src");
      if (src && !img.classList.contains("loaded") && img.src !== src) {
        img.src = src;
      }
    });
  }
}
