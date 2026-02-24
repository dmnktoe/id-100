/**
 * Interactive map module
 * Fetches city contribution counts from /api/map-data,
 * geocodes each city via Nominatim and renders a Leaflet map.
 */

import L from "leaflet";
import { captureException } from "./sentry";

interface CityContrib {
  name: string;
  count: number;
}

interface GeocodedCity {
  name: string;
  count: number;
  lat: number;
  lon: number;
}

// ─── Colours ────────────────────────────────────────────────────────────────

function markerColor(count: number): string {
  if (count >= 75) return "#fbe780";
  if (count >= 31) return "#f5a870";
  if (count >= 11) return "#ed7470";
  return "#73cac4";
}

// ─── Nominatim geocoding ─────────────────────────────────────────────────────

async function geocodeCity(name: string): Promise<{ lat: number; lon: number } | null> {
  try {
    const params = new URLSearchParams({ q: `${name}, Deutschland`, format: "json", limit: "1" });
    const res = await fetch(`https://nominatim.openstreetmap.org/search?${params}`, {
      headers: { "Accept-Language": "de", "User-Agent": "id-100-map/1.0" },
    });
    if (!res.ok) return null;
    const data = await res.json();
    if (!Array.isArray(data) || data.length === 0) return null;
    return { lat: parseFloat(data[0].lat), lon: parseFloat(data[0].lon) };
  } catch {
    return null;
  }
}

/** Nominatim rate-limit: max 1 req/s – fire requests with small delay */
async function geocodeAll(cities: CityContrib[]): Promise<GeocodedCity[]> {
  const results: GeocodedCity[] = [];
  for (let i = 0; i < cities.length; i++) {
    if (i > 0) await new Promise((r) => setTimeout(r, 1100));
    const coords = await geocodeCity(cities[i].name);
    if (coords) {
      results.push({ ...cities[i], ...coords });
    }
  }
  return results;
}

// ─── Leaflet helpers ─────────────────────────────────────────────────────────

function buildMarker(city: GeocodedCity): L.CircleMarker {
  const color = markerColor(city.count);
  const radius = Math.min(10 + Math.sqrt(city.count) * 2.5, 26);

  const marker = L.circleMarker([city.lat, city.lon], {
    radius,
    fillColor: color,
    color: "#fff",
    weight: 2,
    opacity: 1,
    fillOpacity: 0.92,
  });

  marker.bindPopup(`
    <div class="map-popup">
      <strong>${city.name}</strong>
      <span>${city.count} Beitrag${city.count !== 1 ? "e" : ""}</span>
    </div>
  `);

  marker.bindTooltip(String(city.count), {
    permanent: true,
    direction: "center",
    className: "map-marker-label",
  });

  return marker;
}

function buildLegend(): L.Control {
  const legend = new L.Control({ position: "topright" });
  legend.onAdd = () => {
    const div = L.DomUtil.create("div", "map-legend");
    div.innerHTML = `
      <strong class="map-legend-title">Legende</strong>
      <div class="map-legend-item">
        <span class="map-legend-dot" style="background:#73cac4"></span> 1–10 Beiträge
      </div>
      <div class="map-legend-item">
        <span class="map-legend-dot" style="background:#ed7470"></span> 11–30 Beiträge
      </div>
      <div class="map-legend-item">
        <span class="map-legend-dot" style="background:#f5a870"></span> 31–75 Beiträge
      </div>
      <div class="map-legend-item">
        <span class="map-legend-dot" style="background:#fbe780"></span> 75+ Beiträge
      </div>`;
    return div;
  };
  return legend;
}

// ─── Main export ─────────────────────────────────────────────────────────────

export async function initMap(): Promise<void> {
  const el = document.getElementById("map");
  if (!el) return; // not on map page

  // Show loading state
  el.innerHTML = `<div class="map-loading"><span>Karte wird geladen…</span></div>`;

  // 1) Fetch city data from backend
  let cities: CityContrib[] = [];
  try {
    const res = await fetch("/api/map-data");
    if (res.ok) cities = await res.json();
  } catch (err) {
    captureException(err, { module: "map", action: "fetch-map-data" });
    el.innerHTML = `<div class="map-loading map-error">Karte konnte nicht geladen werden.</div>`;
    return;
  }

  if (!cities || cities.length === 0) {
    el.innerHTML = `<div class="map-loading">Noch keine Städte eingetragen.</div>`;
    return;
  }

  // 2) Geocode cities via Nominatim
  const geocoded = await geocodeAll(cities);

  if (geocoded.length === 0) {
    captureException(new Error("All Nominatim geocoding requests failed"), {
      module: "map",
      cities: cities.map((c) => c.name),
    });
    el.innerHTML = `<div class="map-loading map-error">Koordinaten konnten nicht geladen werden.</div>`;
    return;
  }

  // Clear loading state and init Leaflet
  el.innerHTML = "";

  const avgLat = geocoded.reduce((s, c) => s + c.lat, 0) / geocoded.length;
  const avgLon = geocoded.reduce((s, c) => s + c.lon, 0) / geocoded.length;

  const map = L.map(el, {
    center: [avgLat, avgLon],
    zoom: 7,
    scrollWheelZoom: false,
  });

  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution:
      '© <a href="https://leaflet.js.com">Leaflet</a> | © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 18,
  }).addTo(map);

  geocoded.forEach((city) => buildMarker(city).addTo(map));

  buildLegend().addTo(map);

  if (geocoded.length > 1) {
    const bounds = L.latLngBounds(geocoded.map((c) => [c.lat, c.lon] as [number, number]));
    map.fitBounds(bounds, { padding: [48, 48] });
  }
}
