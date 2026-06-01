import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { getToken, setToken, clearAuthSession, getUser, setUser } from "@/utils/authStorage";
import apiClient from "@/api/client";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(getToken() || "");
  const username = ref("");
  const role = ref("");

  // Restore persisted user on init
  const storedUser = getUser();
  if (storedUser) {
    username.value = storedUser.username || "";
    role.value = storedUser.role || "";
  }

  const isLoggedIn = computed(() => !!token.value);

  async function login(user, pass) {
    const res = await apiClient.post("/auth/login", { username: user, password: pass });
    if (!res.data || res.data.status !== "ok" || !res.data.data) {
      throw new Error(res.data?.message || "Login failed");
    }
    const payload = res.data.data;
    token.value = payload.token;
    username.value = payload.username;
    role.value = payload.role || "admin";
    setToken(token.value);
    setUser({ username: username.value, role: role.value });
    // Keep legacy keys for backward compatibility
    localStorage.setItem("mc_username", username.value);
    localStorage.setItem("mc_role", role.value);
  }

  function logout() {
    token.value = "";
    username.value = "";
    role.value = "";
    clearAuthSession();
  }

  return { token, username, role, isLoggedIn, login, logout };
});