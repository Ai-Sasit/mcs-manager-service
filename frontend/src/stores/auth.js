import { defineStore } from "pinia";
import { ref } from "vue";
import serverApi from "../api";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("mc_token") || "");
  const username = ref("");

  const isLoggedIn = () => !!token.value;

  async function login(user, pass) {
    const res = await serverApi.login(user, pass);
    token.value = res.data.token;
    username.value = res.data.username;
    localStorage.setItem("mc_token", token.value);
  }

  function logout() {
    token.value = "";
    username.value = "";
    localStorage.removeItem("mc_token");
  }

  return { token, username, isLoggedIn, login, logout };
});
