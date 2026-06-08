/**
 * Brand animation module
 * Plays the one-shot brand header animation on homepage <-> subpage navigation
 */

const STORAGE_KEY = "brandAnimated";
const COMPACT_SHOWN = "1";
const COMPACT_HIDDEN = "0";

function readFlag(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

function writeFlag(value: string): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, value);
  } catch {
    /* storage unavailable */
  }
}

function prefersReducedMotion(): boolean {
  return (
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

export function initBrandAnimation(): void {
  const brandFull = document.querySelector<HTMLElement>(".brand-full");
  const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

  if (!brandFull && !brandCompact) return;

  const flag = readFlag();
  const animate = !prefersReducedMotion();

  // Homepage, returning from a subpage: unfold the text back in.
  if (brandFull && flag === COMPACT_SHOWN) {
    writeFlag(COMPACT_HIDDEN);
    if (animate) brandFull.classList.add("reverse-animated");
  }

  // Subpage, compact state not shown yet: fold the text away.
  if (brandCompact && flag !== COMPACT_SHOWN) {
    writeFlag(COMPACT_SHOWN);
    if (animate) brandCompact.classList.add("animated");
  }
}
