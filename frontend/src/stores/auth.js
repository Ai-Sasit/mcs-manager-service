import { defineStore } from "pinia";
import { ref } from "vue";
import serverApi from "../api";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("mc_token") || "");
  const username = ref(localStorage.getItem("mc_username") || "");
  const role = ref(localStorage.getItem("mc_role") || "");

  const isLoggedIn = () => !!token.value;

  async function login(user, pass) {
    const res = await serverApi.login(user, pass);
    token.value = res.data.token;
    username.value = res.data.username;
    role.value = res.data.role || "admin";
    localStorage.setItem("mc_token", token.value);
    localStorage.setItem("mc_username", username.value);
    localStorage.setItem("mc_role", role.value);
  }

  function logout() {
    token.value = "";
    username.value = "";
    role.value = "";
    localStorage.removeItem("mc_token");
    localStorage.removeItem("mc_username");
    localStorage.removeItem("mc_role");
  }

  return { token, username, role, isLoggedIn, login, logout };
});
