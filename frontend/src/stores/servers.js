import { defineStore } from "pinia";
import api from "@/api";

export const useServersStore = defineStore("servers", {
  state: () => ({
    servers: [],
    loading: false,
    error: null,
  }),

  getters: {
    runningCount: (state) =>
      state.servers.filter((s) => s.status === "running").length,
    stoppedCount: (state) =>
      state.servers.filter((s) => s.status === "stopped").length,
  },

  actions: {
    async fetchServers() {
      this.loading = true;
      this.error = null;
      try {
        const { data } = await api.listServers();
        this.servers = data.data || [];
      } catch (e) {
        this.error = e.message;
      } finally {
        this.loading = false;
      }
    },

    async createServer(payload) {
      const { data } = await api.createServer(payload);
      if (data.success) {
        this.servers.push(data.data);
      }
      return data;
    },

    async deleteServer(id) {
      await api.deleteServer(id);
      this.servers = this.servers.filter((s) => s.id !== id);
    },

    async startServer(id) {
      await api.startServer(id);
      const s = this.servers.find((s) => s.id === id);
      if (s) s.status = "running";
    },

    async stopServer(id) {
      await api.stopServer(id);
      const s = this.servers.find((s) => s.id === id);
      if (s) s.status = "stopped";
    },

    async restartServer(id) {
      const s = this.servers.find((s) => s.id === id);
      if (s) s.status = "starting";
      await api.restartServer(id);
      if (s) s.status = "running";
    },
  },
});
