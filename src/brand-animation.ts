/**
 * Brand animation module
 * Handles brand animation logic on page transitions
 */

export function initBrandAnimation(): void {
  const brandFull = document.querySelector<HTMLElement>(".brand-full");
  const brandCompact = document.querySelector<HTMLElement>(".brand-compact");

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
}
