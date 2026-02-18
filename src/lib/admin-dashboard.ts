/**
 * Admin dashboard functionality
 * Handles token management, bag requests, and contribution management
 */

import { getErrorMessage } from "./utils";

/**
 * Initialize admin dashboard functionality
 */
export function initAdminDashboard(): void {
  // Create new token/bag form submission
  const createTokenForm = document.getElementById("createTokenForm") as HTMLFormElement | null;
  if (createTokenForm) {
    createTokenForm.onsubmit = async function (e: Event) {
      e.preventDefault();

      const bagName = (document.getElementById("bagName") as HTMLInputElement).value;
      const maxUploads = parseInt(
        (document.getElementById("maxUploads") as HTMLInputElement).value
      );

      try {
        const response = await fetch("/admin/tokens", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ bag_name: bagName, max_uploads: maxUploads }),
        });

        const data = await response.json();

        if (response.ok) {
          const resultDiv = document.getElementById("createResult") as HTMLElement;
          resultDiv.style.display = "block";
          resultDiv.style.background = "#e8f5e9";
          resultDiv.style.padding = "1rem";
          resultDiv.style.borderRadius = "4px";
          resultDiv.innerHTML = `
        <strong>✅ Token erstellt!</strong><br>
        <small>Token-ID: ${data.token_id}</small><br>
        <a href="${data.qr_url}?format=svg" target="_blank" style="color: #2196F3;">📥 QR-Code (SVG) herunterladen</a> | 
        <a href="${data.qr_url}?format=png" target="_blank" style="color: #2196F3;">📥 QR-Code (PNG) herunterladen</a><br>
        <small style="word-break: break-all;">URL: ${data.upload_url} <button id="copyNewUrlBtn" style="margin-left:0.5rem; padding:0.2rem 0.4rem; border-radius:4px;">🔗 Kopieren</button></small>
      `;

          const copyBtn = document.getElementById("copyNewUrlBtn");
          if (copyBtn) {
            copyBtn.onclick = async function () {
              try {
                await navigator.clipboard.writeText(data.upload_url);
                alert("URL kopiert");
              } catch (err) {
                prompt("URL kopieren:", data.upload_url);
              }
            };
          }

          // Reload nach 3 Sekunden
          setTimeout(() => location.reload(), 3000);
        } else {
          alert("Fehler: " + (data.error || "Unbekannter Fehler"));
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : "Unbekannter Fehler";
        alert("Fehler: " + errorMessage);
      }
    };
  }

  // Initialize AJAX filter for bag requests
  initBagRequestFilter();
}

/**
 * Reset a token
 */
export async function resetToken(id: number, name: string): Promise<void> {
  if (!confirm(`${name} wirklich zurücksetzen? Der Upload-Counter wird auf 0 gesetzt.`)) return;

  try {
    const response = await fetch(`/admin/tokens/${id}/reset`, { method: "POST" });
    const data = await response.json();
    alert(data.message || "Token wurde zurückgesetzt");
    location.reload();
  } catch (err) {
    alert("Fehler: " + getErrorMessage(err));
  }
}

/**
 * Deactivate a token
 */
export async function deactivateToken(id: number, name: string): Promise<void> {
  if (!confirm(`${name} wirklich deaktivieren?`)) return;

  try {
    const response = await fetch(`/admin/tokens/${id}/deactivate`, { method: "POST" });
    const data = await response.json();
    alert(data.status || "Token wurde deaktiviert");
    location.reload();
  } catch (err) {
    alert("Fehler: " + getErrorMessage(err));
  }
}

/**
 * Update token quota
 */
export async function updateQuota(id: number): Promise<void> {
  const input = document.getElementById(`quota-${id}`) as HTMLInputElement;
  const newQuota = parseInt(input.value);

  if (newQuota <= 0) {
    alert("Kontingent muss größer als 0 sein");
    return;
  }

  try {
    const response = await fetch(`/admin/tokens/${id}/quota`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ max_uploads: newQuota }),
    });

    const data = await response.json();
    alert(`✅ Kontingent auf ${data.max_uploads} aktualisiert`);
    location.reload();
  } catch (err) {
    alert("Fehler: " + getErrorMessage(err));
  }
}

/**
 * Download QR code
 */
export function downloadQR(id: number, _name: string, format: "svg" | "png"): void {
  const url = `/admin/tokens/${id}/qr?format=${format}`;
  window.open(url, "_blank");
}

/**
 * Copy upload URL to clipboard
 */
export async function copyUploadURL(token: string, _name: string): Promise<void> {
  const encodedToken = encodeURIComponent(token);
  const url = `${location.origin}/upload?token=${encodedToken}`;
  try {
    await navigator.clipboard.writeText(url);
    alert("URL kopiert: " + url);
  } catch (err) {
    // fallback to prompt if clipboard API fails
    prompt("URL kopieren:", url);
  }
}

/**
 * Mark bag request as done
 */
export async function markBagRequestDone(id: number, btn: HTMLElement): Promise<void> {
  if (!confirm("Als erledigt markieren?")) return;
  try {
    const res = await fetch(`/admin/werkzeug-anfragen/${id}/complete`, { method: "POST" });
    const data = await res.json();
    if (res.ok) {
      const li = btn.closest("li");
      const span = document.createElement("span");
      span.style.color = "#4CAF50";
      span.style.fontWeight = "600";
      span.textContent = "✅ Erledigt";
      btn.replaceWith(span);
      if (li) {
        li.style.opacity = "0.8";
      }
    } else {
      alert("Fehler: " + (data.error || "Unbekannter Fehler"));
    }
  } catch (err) {
    alert("Fehler: " + getErrorMessage(err));
  }
}

/**
 * Initialize AJAX filter for bag requests to avoid full page reload
 */
function initBagRequestFilter(): void {
  document.addEventListener("click", (e: MouseEvent) => {
    const target = e.target as HTMLElement;
    const link = target.closest("a.filter-btn") as HTMLAnchorElement | null;
    if (!link) return;
    // only handle links inside the admin section
    if (!link.closest(".admin-section")) return;
    e.preventDefault();

    const href = link.getAttribute("href");
    if (!href) return;

    const fetchUrl = href.includes("?") ? href + "&partial=1" : href + "?partial=1";

    fetch(fetchUrl)
      .then((r) => {
        if (!r.ok) throw new Error("fetch failed");
        return r.text();
      })
      .then((html) => {
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, "text/html");
        const newContainer = doc.querySelector("#bagRequestsContainer");
        if (!newContainer) {
          // fallback to full navigation
          window.location.href = href;
          return;
        }
        const container = document.getElementById("bagRequestsContainer");
        if (container) {
          container.innerHTML = newContainer.innerHTML;
        }
        // update active states on buttons
        document
          .querySelectorAll(".admin-section a.filter-btn")
          .forEach((a) => a.classList.remove("active"));
        link.classList.add("active");
        // update URL without scrolling
        history.pushState(null, "", href);
      })
      .catch((err) => {
        console.error(err);
        window.location.href = href;
      });
  });
}

/**
 * Delete contribution (admin)
 */
export async function deleteContribution(id: number, _btn: HTMLElement): Promise<void> {
  if (
    !confirm(
      "Contribution wirklich löschen? Dies wird das Bild aus der Datenbank und dem Storage entfernen."
    )
  )
    return;

  try {
    const response = await fetch(`/admin/contributions/${id}/delete`, { method: "POST" });
    const data = await response.json();

    if (response.ok) {
      // Remove the card from the DOM
      const card = document.getElementById(`contrib-${id}`);
      if (card) {
        card.style.opacity = "0";
        card.style.transition = "opacity 0.3s";
        setTimeout(() => card.remove(), 300);
      }
      alert("Contribution erfolgreich gelöscht");
    } else {
      alert("Fehler: " + (data.error || "Unbekannter Fehler"));
    }
  } catch (err) {
    alert("Fehler: " + getErrorMessage(err));
  }
}

// Export functions to global window object for inline onclick handlers
if (typeof window !== "undefined") {
  (window as any).resetToken = resetToken;
  (window as any).deactivateToken = deactivateToken;
  (window as any).updateQuota = updateQuota;
  (window as any).downloadQR = downloadQR;
  (window as any).copyUploadURL = copyUploadURL;
  (window as any).markBagRequestDone = markBagRequestDone;
  (window as any).deleteContribution = deleteContribution;
}
