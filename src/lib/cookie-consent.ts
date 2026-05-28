const CONSENT_KEY = "id100_analytics_consent";

function hasConsent(): boolean {
  return document.cookie.split("; ").some((c) => c.startsWith(`${CONSENT_KEY}=accepted`));
}

function setConsent(): void {
  const maxAge = 60 * 60 * 24 * 365;
  document.cookie = `${CONSENT_KEY}=accepted; path=/; max-age=${maxAge}; SameSite=Lax`;
}

function loadUmami(): void {
  const scriptUrl = window.UMAMI_SCRIPT_URL;
  const websiteId = window.UMAMI_WEBSITE_ID;

  if (!scriptUrl || !websiteId) return;

  const script = document.createElement("script");
  script.defer = true;
  script.src = scriptUrl;
  script.setAttribute("data-website-id", websiteId);
  document.head.appendChild(script);
}

function createBanner(): HTMLElement {
  const banner = document.createElement("div");
  banner.className = "cookie-banner";
  banner.setAttribute("role", "dialog");
  banner.setAttribute("aria-label", "Cookie-Hinweis");

  const text = document.createElement("p");
  text.className = "cookie-banner-text";
  text.innerHTML =
    'Diese Website verwendet Cookies zur Analyse. <a href="/datenschutz">Mehr erfahren</a>';

  const btn = document.createElement("button");
  btn.className = "cookie-banner-accept";
  btn.textContent = "OK";
  btn.addEventListener("click", () => {
    setConsent();
    loadUmami();
    banner.classList.add("cookie-banner-hidden");
    setTimeout(() => banner.remove(), 300);
  });

  banner.appendChild(text);
  banner.appendChild(btn);
  return banner;
}

export function initCookieConsent(): void {
  if (hasConsent()) {
    loadUmami();
    return;
  }

  const banner = createBanner();
  document.body.appendChild(banner);
}
