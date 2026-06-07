/**
 * Brand animation module
 *
 * The site header renders the brand title in two variants (see
 * `web/templates/components/header.html`):
 *  - `.brand-full`    on the homepage: full text, emojis de-emphasised (blurred)
 *  - `.brand-compact` on subpages:     text folded away, only the emojis remain
 *
 * On navigation we play a one-shot entrance animation that bridges the two
 * variants, driven entirely by CSS classes:
 *  - forward (home -> subpage): `.brand-compact.animated`     folds the text away
 *  - reverse (subpage -> home): `.brand-full.reverse-animated` unfolds the text
 *
 * A `sessionStorage` flag remembers whether the compact state has already been
 * shown, so the animation only plays on the relevant transition and not on
 * every page load.
 */

const STORAGE_KEY = "brandAnimated";
const COMPACT_SHOWN = "1";
const COMPACT_HIDDEN = "0";

/**
 * Duration after which the reverse-animation class is removed so the element
 * settles back into its plain resting state. Must be at least as long as the
 * longest reverse keyframe (`blurEmoji`, 0.8s in `web/static/style.css`).
 */
const REVERSE_ANIMATION_MS = 800;

/** Read the flag without throwing when storage is unavailable (private mode, etc.). */
function readFlag(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

/** Write the flag, silently ignoring storage failures. */
function writeFlag(value: string): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, value);
  } catch {
    // Storage unavailable: the animation still plays, it just isn't
    // remembered across navigations. Not worth breaking init over.
  }
}

/** Whether the user has asked the OS/browser to minimise motion. */
function prefersReducedMotion(): boolean {
  return (
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

export function initBrandAnimation(): void {
  const brandFull = document.querySelector<HTMLElement>(".brand-full");
  const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

  // No brand header on this page (e.g. error pages) - nothing to do.
  if (!brandFull && !brandCompact) {
    return;
  }

  const flag = readFlag();
  const animate = !prefersReducedMotion();

  // Homepage, returning from a subpage -> unfold the text back in.
  if (brandFull && flag === COMPACT_SHOWN) {
    writeFlag(COMPACT_HIDDEN);
    if (animate) {
      playReverseAnimation(brandFull);
    }
  }

  // Subpage, compact state not shown yet -> fold the text away.
  if (brandCompact && flag !== COMPACT_SHOWN) {
    writeFlag(COMPACT_SHOWN);
    if (animate) {
      brandCompact.classList.add("animated");
    }
  }
}

/**
 * Plays the reverse (text unfold) animation on the full brand, then removes the
 * class once the animation has finished so the element returns to its plain
 * resting state (which is visually identical to the animation's end frame).
 */
function playReverseAnimation(brandFull: HTMLElement): void {
  brandFull.classList.add("reverse-animated");
  window.setTimeout(() => {
    brandFull.classList.remove("reverse-animated");
  }, REVERSE_ANIMATION_MS);
}
