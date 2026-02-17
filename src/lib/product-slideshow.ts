/**
 * Product slideshow module
 * Implements photo slideshow using Swiper.js
 */

import Swiper from 'swiper';
import { Navigation, Pagination, Autoplay, EffectFade, Keyboard } from 'swiper/modules';

export function initProductSlideshow(): void {
  const swiperContainer = document.querySelector<HTMLElement>('.product-slideshow');
  
  if (!swiperContainer) return;

  // Initialize Swiper
  new Swiper('.product-slideshow', {
    modules: [Navigation, Pagination, Autoplay, EffectFade, Keyboard],
    
    // Slideshow settings
    loop: true,
    effect: 'fade',
    fadeEffect: {
      crossFade: true
    },
    speed: 800,
    
    // Auto-play
    autoplay: {
      delay: 4000,
      disableOnInteraction: false,
      pauseOnMouseEnter: true
    },
    
    // Navigation arrows
    navigation: {
      nextEl: '.swiper-button-next',
      prevEl: '.swiper-button-prev',
    },
    
    // Pagination dots
    pagination: {
      el: '.swiper-pagination',
      clickable: true,
      dynamicBullets: true
    },
    
    // Keyboard control
    keyboard: {
      enabled: true,
      onlyInViewport: true
    },
    
    // Accessibility
    a11y: {
      prevSlideMessage: 'Vorheriges Bild',
      nextSlideMessage: 'Nächstes Bild',
      paginationBulletMessage: 'Gehe zu Bild {{index}}'
    }
  });
}
