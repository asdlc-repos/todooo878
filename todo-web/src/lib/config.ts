declare global {
  interface Window {
    __TODOOO_CONFIG__?: { apiBaseUrl?: string };
  }
}

export function getApiBaseUrl(): string {
  if (typeof window !== "undefined" && window.__TODOOO_CONFIG__?.apiBaseUrl) {
    return window.__TODOOO_CONFIG__.apiBaseUrl.replace(/\/$/, "");
  }
  return "/api";
}
