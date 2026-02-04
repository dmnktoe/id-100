/**
 * Dynamic Emoji Favicon
 * Sets a random emoji as favicon on each page load
 */

// Curated emoji selection matching the 🏠🆔💯 theme
const EMOJI_POOL = [
  '🏠', // house
  '🆔', // ID
  '💯', // 100
  '🚨' // Alert
];

/**
 * Generates an SVG data URL with the given emoji
 */
function createEmojiSvg(emoji: string): string {
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <text y="80" font-size="80">${emoji}</text>
    </svg>
  `;
  return `data:image/svg+xml,${encodeURIComponent(svg)}`;
}

/**
 * Picks a random emoji from the pool
 */
function getRandomEmoji(): string {
  return EMOJI_POOL[Math.floor(Math.random() * EMOJI_POOL.length)];
}

/**
 * Sets the favicon to a random emoji
 */
export function setRandomEmojiFavicon(): void {
  const emoji = getRandomEmoji();
  const faviconUrl = createEmojiSvg(emoji);

  // Find existing favicon or create new one
  let link = document.querySelector<HTMLLinkElement>("link[rel*='icon']");
  
  if (!link) {
    link = document.createElement('link');
    link.rel = 'icon';
    document.head.appendChild(link);
  }

  link.href = faviconUrl;
}

// Auto-run on module load
setRandomEmojiFavicon();
