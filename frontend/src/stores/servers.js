import { defineStore } from "pinia";
import { ref, computed } from "vue";
import apiClient from "@/api/client";
import { getApiErrorMessage, isApiSuccess } from "@/utils/apiError";

export const useServersStore = defineStore("servers", () => {
  const servers = ref([]);
  const loading = ref(false);
  const error = ref(null);

  const runningCount = computed(() =>
    servers.value.filter((s) => s.status === "running").length,
  );

  const stoppedCount = computed(() =>
    servers.value.filter((s) => s.status === "stopped").length,
  );

  async function fetchServers() {
    loading.value = true;
    error.value = null;
    try {
      const { data } = await apiClient.get("/servers");
      servers.value = data.data || [];
    } catch (e) {
      error.value = getApiErrorMessage(e, "Failed to fetch servers");
    } finally {
      loading.value = false;
    }
  }

  async function createServer(payload) {
    const { data } = await apiClient.post("/servers", payload);
    if (isApiSuccess(data)) {
      servers.value.push(data.data);
    }
    return data;
  }

  async function deleteServer(id) {
    await apiClient.delete(`/servers/${encodeURIComponent(id)}`);
    servers.value = servers.value.filter((s) => s.id !== id);
  }

  async function startServer(id) {
    const s = servers.value.find((s) => s.id === id);
    const prevStatus = s?.status;
    if (s) s.status = "starting";
    try {
      const { data } = await apiClient.post(`/servers/${encodeURIComponent(id)}/start`);
      if (!isApiSuccess(data)) throw new Error(data.message || "Failed to start server");
      if (s) s.status = "running";
    } catch (e) {
      if (s) s.status = prevStatus || "stopped";
      throw e;
    }
  }

  async function stopServer(id) {
    const s = servers.value.find((s) => s.id === id);
    const prevStatus = s?.status;
    if (s) s.status = "stopping";
    try {
      const { data } = await apiClient.post(`/servers/${encodeURIComponent(id)}/stop`);
      if (!isApiSuccess(data)) throw new Error(data.message || "Failed to stop server");
      if (s) s.status = "stopped";
    } catch (e) {
      if (s) s.status = prevStatus || "running";
      throw e;
    }
  }

  function updateServerStatus(id, status) {
    const s = servers.value.find((s) => s.id === id);
    if (s) s.status = status;
  }

  async function restartServer(id) {
    const s = servers.value.find((s) => s.id === id);
    const prevStatus = s?.status;
    if (s) s.status = "starting";
    try {
      const { data } = await apiClient.post(`/servers/${encodeURIComponent(id)}/restart`);
      if (!isApiSuccess(data)) throw new Error(data.message || "Failed to restart server");
      if (s) s.status = "running";
    } catch (e) {
      if (s) s.status = prevStatus || "stopped";
      throw e;
    }
  }

  return {
    servers,
    loading,
    error,
    runningCount,
    stoppedCount,
    fetchServers,
    createServer,
    deleteServer,
    startServer,
    stopServer,
    updateServerStatus,
    restartServer,
  };
});
