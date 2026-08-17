import { computed, ref } from "vue";

const STORAGE_KEY = "mc-manage.theme";
const theme = ref("light");
let initialized = false;

function getSystemTheme() {
  return window.matchMedia?.("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

function getStoredTheme() {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    return stored === "dark" || stored === "light" ? stored : null;
  } catch {
    return null;
  }
}

function setDocumentTheme(nextTheme) {
  document.documentElement.dataset.theme = nextTheme;
  theme.value = nextTheme;
}

export function initTheme() {
  if (typeof window === "undefined" || initialized) return;
  setDocumentTheme(getStoredTheme() || getSystemTheme());
  initialized = true;
}

export function useTheme() {
  function setTheme(nextTheme) {
    const validTheme = nextTheme === "dark" ? "dark" : "light";
    setDocumentTheme(validTheme);
    try {
      window.localStorage.setItem(STORAGE_KEY, validTheme);
    } catch {
      // The active session still gets the chosen appearance when storage is unavailable.
    }
  }

  function toggleTheme() {
    setTheme(theme.value === "dark" ? "light" : "dark");
  }

  return {
    theme,
    isDark: computed(() => theme.value === "dark"),
    setTheme,
    toggleTheme,
  };
}
