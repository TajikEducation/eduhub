import type { Region } from "./data";

// Approximate city/region centroids for Tajikistan (WGS84).
export const REGION_CENTROIDS: Record<Region, { lat: number; lng: number }> = {
  dushanbe: { lat: 38.5598, lng: 68.787 },
  sughd: { lat: 40.2833, lng: 69.6333 },
  khatlon: { lat: 37.8606, lng: 68.7803 },
  gbao: { lat: 37.4917, lng: 71.5167 },
  rrp: { lat: 38.75, lng: 69.5 },
};

export function haversine(a: { lat: number; lng: number }, b: { lat: number; lng: number }): number {
  const R = 6371;
  const dLat = ((b.lat - a.lat) * Math.PI) / 180;
  const dLng = ((b.lng - a.lng) * Math.PI) / 180;
  const lat1 = (a.lat * Math.PI) / 180;
  const lat2 = (b.lat * Math.PI) / 180;
  const h =
    Math.sin(dLat / 2) ** 2 + Math.sin(dLng / 2) ** 2 * Math.cos(lat1) * Math.cos(lat2);
  return 2 * R * Math.asin(Math.sqrt(h));
}

export function nearestRegion(lat: number, lng: number): Region {
  let best: Region = "dushanbe";
  let bestDist = Infinity;
  for (const region of Object.keys(REGION_CENTROIDS) as Region[]) {
    const d = haversine({ lat, lng }, REGION_CENTROIDS[region]);
    if (d < bestDist) {
      bestDist = d;
      best = region;
    }
  }
  return best;
}

export function detectRegionByGPS(): Promise<Region> {
  return new Promise((resolve, reject) => {
    if (typeof navigator === "undefined" || !navigator.geolocation) {
      reject(new Error("geolocation unavailable"));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve(nearestRegion(pos.coords.latitude, pos.coords.longitude)),
      (err) => reject(err),
      { timeout: 8000 }
    );
  });
}

// Точные координаты — только для сортировки «рядом со мной» в текущей сессии.
// НЕ персистить (localStorage/app-state) и не отправлять на сервер без необходимости —
// минимизация PII (CLAUDE.md NFR). Регион (nearestRegion) персистить можно, это не PII.
export function detectCoords(): Promise<{ lat: number; lng: number }> {
  return new Promise((resolve, reject) => {
    if (typeof navigator === "undefined" || !navigator.geolocation) {
      reject(new Error("geolocation unavailable"));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve({ lat: pos.coords.latitude, lng: pos.coords.longitude }),
      (err) => reject(err),
      { timeout: 8000 }
    );
  });
}
