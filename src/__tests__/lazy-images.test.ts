/**
 * Tests for lazy-images module
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { initLazyImages } from '../lib/lazy-images'

describe('initLazyImages', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('should handle no images gracefully', () => {
    expect(() => initLazyImages()).not.toThrow()
  })

  it('should initialize lazy images with LQIP', () => {
    document.body.innerHTML = `
      <img class="lazy" data-lqip="data:image/png;base64,abc" data-src="/full.jpg" />
    `
    
    const img = document.querySelector<HTMLImageElement>('img.lazy')
    
    initLazyImages()
    
    // Should set placeholder
    expect(img?.src).toContain('data:image/png')
    
    // Should be marked as initialized
    expect(img?.dataset.lazyInitialized).toBe('1')
  })

  it('should not reinitialize already initialized images', () => {
    document.body.innerHTML = `
      <img class="lazy" data-lazy-initialized="1" data-lqip="data:image/png;base64,abc" />
    `
    
    const img = document.querySelector<HTMLImageElement>('img.lazy')
    const originalSrc = img?.src
    
    initLazyImages()
    
    // Should not change already initialized image
    expect(img?.src).toBe(originalSrc)
  })

  it('should mark image as loaded when complete and has naturalWidth', () => {
    document.body.innerHTML = `
      <img class="lazy" data-src="/full.jpg" src="/full.jpg" />
    `
    
    const img = document.querySelector<HTMLImageElement>('img.lazy')!
    
    // Mock complete image with matching src
    Object.defineProperty(img, 'complete', { value: true, writable: true })
    Object.defineProperty(img, 'naturalWidth', { value: 100, writable: true })
    Object.defineProperty(img, 'currentSrc', { value: '/full.jpg', writable: true })
    
    initLazyImages()
    
    // Should be marked as loaded
    expect(img.classList.contains('loaded')).toBe(true)
  })

  it('should handle images without data-lqip', () => {
    document.body.innerHTML = `
      <img class="lazy" data-src="/full.jpg" />
    `
    
    expect(() => initLazyImages()).not.toThrow()
    
    const img = document.querySelector<HTMLImageElement>('img.lazy')
    expect(img?.dataset.lazyInitialized).toBe('1')
  })

  it('should work with custom root container', () => {
    const container = document.createElement('div')
    container.innerHTML = `
      <img class="lazy" data-lqip="data:image/png;base64,abc" data-src="/full.jpg" />
    `
    document.body.appendChild(container)
    
    expect(() => initLazyImages(container)).not.toThrow()
    
    const img = container.querySelector<HTMLImageElement>('img.lazy')
    expect(img?.dataset.lazyInitialized).toBe('1')
  })

  it('should add load event listener', () => {
    document.body.innerHTML = `
      <img class="lazy" data-src="/full.jpg" />
    `
    
    const img = document.querySelector<HTMLImageElement>('img.lazy')!
    const addEventListenerSpy = vi.spyOn(img, 'addEventListener')
    
    initLazyImages()
    
    expect(addEventListenerSpy).toHaveBeenCalledWith('load', expect.any(Function))
    expect(addEventListenerSpy).toHaveBeenCalledWith('error', expect.any(Function))
  })

  it('should fallback to eager loading if IntersectionObserver is not available', () => {
    // Mock no IntersectionObserver
    const originalIO = global.IntersectionObserver
    // @ts-ignore
    delete global.IntersectionObserver
    
    document.body.innerHTML = `
      <img class="lazy" data-lqip="data:image/png;base64,abc" data-src="/full.jpg" />
    `
    
    const img = document.querySelector<HTMLImageElement>('img.lazy')!
    
    initLazyImages()
    
    // Should set src directly
    expect(img.src).toContain('/full.jpg')
    
    // Restore IntersectionObserver
    global.IntersectionObserver = originalIO
  })
})
