/**
 * Form handler module
 * Handles form submissions (e.g., request bag form)
 */

import { captureException, addBreadcrumb } from "./sentry";

export function initFormHandlers(): void {
  // Formular "werkzeug anfordern" submission handler
  document.addEventListener("submit", (e) => {
    const form = e.target as HTMLFormElement;
    if (!form || form.id !== "requestBagForm") return;
    e.preventDefault();

    const emailInput = form.querySelector<HTMLInputElement>('input[name="email"]');
    const btn = form.querySelector<HTMLButtonElement>("button[type=submit]");
    const resultDiv = form.querySelector<HTMLDivElement>("#requestResult");

    if (!emailInput || !btn || !resultDiv) return;

    const email = emailInput.value.trim();

    if (!email || !email.includes("@")) {
      resultDiv.style.display = "block";
      resultDiv.style.color = "#d32f2f";
      resultDiv.innerText = "Bitte gib eine gültige E‑Mail an.";
      return;
    }

    btn.disabled = true;
    btn.innerText = "sende...";

    addBreadcrumb("Bag request form submitted", "form", { email });

    fetch("/werkzeug-anfordern", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email }),
    })
      .then((r) => r.json())
      .then((data: { status: string; error?: string }) => {
        if (data.status === "ok") {
          form.innerHTML = `<p style="font-weight:500;">Danke! Wir benachrichtigen dich per E‑Mail, sobald ein Werkzeug verfügbar ist.</p>`;
        } else {
          resultDiv.style.display = "block";
          resultDiv.style.color = "#d32f2f";
          resultDiv.innerText = data.error || "Etwas ging schief.";
          btn.disabled = false;
          btn.innerText = "anfragen";
        }
      })
      .catch((err) => {
        captureException(err, { context: "bag-request-form", email });
        resultDiv.style.display = "block";
        resultDiv.style.color = "#d32f2f";
        resultDiv.innerText = "Netzwerkfehler. Bitte versuche es erneut.";
        btn.disabled = false;
        btn.innerText = "anfragen";
      });
  });
}
