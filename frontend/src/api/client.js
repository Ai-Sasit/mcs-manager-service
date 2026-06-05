import axios from "axios";
import { clearAuthSession, getToken } from "@/utils/authStorage";
import { getApiErrorMessage } from "@/utils/apiError";

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1",
});

apiClient.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers["Authorization"] = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    error.apiMessage = getApiErrorMessage(error);
    if (error.response && error.response.status === 401) {
      clearAuthSession();
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
      return Promise.reject(error);
    }

    return Promise.reject(error);
  },
);

export default apiClient;
