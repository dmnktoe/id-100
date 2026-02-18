/**
 * Tests for product-slideshow module
 */
import { describe, it, expect, beforeEach, vi } from "vitest";
import { initProductSlideshow } from "../lib/product-slideshow";
import Swiper from "swiper";

// Mock Swiper
vi.mock("swiper", () => {
  return {
    default: vi.fn(),
  };
});

vi.mock("swiper/modules", () => {
  return {
    Navigation: {},
    Pagination: {},
    Autoplay: {},
    EffectFade: {},
    Keyboard: {},
  };
});

vi.mock("swiper/css", () => ({}));
vi.mock("swiper/css/navigation", () => ({}));
vi.mock("swiper/css/pagination", () => ({}));
vi.mock("swiper/css/effect-fade", () => ({}));

describe("initProductSlideshow", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    vi.clearAllMocks();
  });

  it("should handle missing slideshow container gracefully", () => {
    expect(() => initProductSlideshow()).not.toThrow();
    expect(Swiper).not.toHaveBeenCalled();
  });

  it("should initialize Swiper when container exists", () => {
    document.body.innerHTML = `
      <div class="swiper product-slideshow">
        <div class="swiper-wrapper">
          <div class="swiper-slide">
            <img src="/photo1.jpg" alt="Photo 1">
          </div>
          <div class="swiper-slide">
            <img src="/photo2.jpg" alt="Photo 2">
          </div>
        </div>
        <div class="swiper-button-prev"></div>
        <div class="swiper-button-next"></div>
        <div class="swiper-pagination"></div>
      </div>
    `;

    initProductSlideshow();

    // Should initialize Swiper
    expect(Swiper).toHaveBeenCalledWith(
      ".product-slideshow",
      expect.objectContaining({
        loop: true,
        effect: "fade",
        autoplay: expect.objectContaining({
          delay: 4000,
        }),
      })
    );
  });

  it("should configure navigation and pagination", () => {
    document.body.innerHTML = `
      <div class="swiper product-slideshow">
        <div class="swiper-wrapper">
          <div class="swiper-slide"><img src="/photo.jpg" alt="Photo"></div>
        </div>
        <div class="swiper-button-prev"></div>
        <div class="swiper-button-next"></div>
        <div class="swiper-pagination"></div>
      </div>
    `;

    initProductSlideshow();

    expect(Swiper).toHaveBeenCalledWith(
      ".product-slideshow",
      expect.objectContaining({
        navigation: {
          nextEl: ".swiper-button-next",
          prevEl: ".swiper-button-prev",
        },
        pagination: {
          el: ".swiper-pagination",
          clickable: true,
          dynamicBullets: true,
        },
      })
    );
  });

  it("should configure accessibility options", () => {
    document.body.innerHTML = `
      <div class="swiper product-slideshow">
        <div class="swiper-wrapper">
          <div class="swiper-slide"><img src="/photo.jpg" alt="Photo"></div>
        </div>
      </div>
    `;

    initProductSlideshow();

    expect(Swiper).toHaveBeenCalledWith(
      ".product-slideshow",
      expect.objectContaining({
        a11y: {
          prevSlideMessage: "Vorheriges Bild",
          nextSlideMessage: "Nächstes Bild",
          paginationBulletMessage: "Gehe zu Bild {{index}}",
        },
      })
    );
  });
});
