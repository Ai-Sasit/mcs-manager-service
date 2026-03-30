import axios from "axios";
import { API_BASE_URL, API_TIMEOUT } from "../constants";

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("mc_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem("mc_token");
      window.location.href = "/login";
    }
    return Promise.reject(err);
  },
);

export default api;
