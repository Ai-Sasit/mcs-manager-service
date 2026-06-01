const STORAGE_PREFIX = "mc_";

export function getToken() {
  return localStorage.getItem(`${STORAGE_PREFIX}token`);
}

export function setToken(token) {
  localStorage.setItem(`${STORAGE_PREFIX}token`, token);
}

export function clearToken() {
  localStorage.removeItem(`${STORAGE_PREFIX}token`);
}

export function getUser() {
  try {
    const raw = localStorage.getItem(`${STORAGE_PREFIX}user`);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function setUser(user) {
  localStorage.setItem(`${STORAGE_PREFIX}user`, JSON.stringify(user));
}

export function clearUser() {
  localStorage.removeItem(`${STORAGE_PREFIX}user`);
}

export function clearAuthSession() {
  clearToken();
  clearUser();
  localStorage.removeItem(`${STORAGE_PREFIX}username`);
  localStorage.removeItem(`${STORAGE_PREFIX}role`);
}