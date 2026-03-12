/**
 * Top-solved badge strip
 * Horizontally draggable FreeMode slider showing the most solved IDs
 */

import Swiper from "swiper";
import { FreeMode } from "swiper/modules";

export function initTopSolvedSlider(): void {
  const el = document.querySelector<HTMLElement>(".top-solved-swiper");
  if (!el) return;

  new Swiper(".top-solved-swiper", {
    modules: [FreeMode],
    slidesPerView: "auto",
    spaceBetween: 10,
    freeMode: {
      enabled: true,
      momentum: true,
      momentumRatio: 0.5,
    },
    grabCursor: true,
    a11y: {
      enabled: false,
    },
  });
}
