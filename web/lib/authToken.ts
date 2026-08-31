// Хранилище токенов backend/internal/auth (access+refresh JWT). Отдельно от app-state.tsx —
// там localStorage-блоб демо-данных (savedIds/children_/...), токены реальной аутентификации —
// другой жизненный цикл (не должны сериализоваться вместе с демо-стейтом).
const ACCESS_TOKEN_KEY = "eduhub_access_token";
const REFRESH_TOKEN_KEY = "eduhub_refresh_token";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}
