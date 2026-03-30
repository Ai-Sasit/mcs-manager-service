import api from "../plugins/axios";
import { ENDPOINTS } from "../constants";

export default {
  // Auth
  login: (username, password) =>
    api.post("/auth/login", { username, password }),
  getMe: () => api.get("/auth/me"),

  listServers: () => api.get("/servers"),
  createServer: (data) => api.post("/servers", data),
  getServer: (id) => api.get(`/servers/${encodeURIComponent(id)}`),
  deleteServer: (id) => api.delete(`/servers/${encodeURIComponent(id)}`),
  startServer: (id) => api.post(`/servers/${encodeURIComponent(id)}/start`),
  stopServer: (id) => api.post(`/servers/${encodeURIComponent(id)}/stop`),
  restartServer: (id) => api.post(`/servers/${encodeURIComponent(id)}/restart`),

  listPlugins: (id) => api.get(`/servers/${encodeURIComponent(id)}/plugins`),
  uploadPlugin: (id, formData) =>
    api.post(`/servers/${encodeURIComponent(id)}/plugins`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    }),
  deletePlugin: (id, name) =>
    api.delete(
      `/servers/${encodeURIComponent(id)}/plugins/${encodeURIComponent(name)}`,
    ),

  getJavaVersions: () => api.get("/versions/java"),
  getBedrockVersions: () => api.get("/versions/bedrock"),

  // Players
  getPlayers: (id) => api.get(`/servers/${encodeURIComponent(id)}/players`),
  updatePlayer: (id, data) =>
    api.post(`/servers/${encodeURIComponent(id)}/players`, data),

  // Files
  listFiles: (id, path) =>
    api.get(`/servers/${encodeURIComponent(id)}/files?path=${encodeURIComponent(path || "")}`),
  readFile: (id, path) =>
    api.get(`/servers/${encodeURIComponent(id)}/files/read?path=${encodeURIComponent(path)}`),
  writeFile: (id, path, content) =>
    api.post(`/servers/${encodeURIComponent(id)}/files/write?path=${encodeURIComponent(path)}`, { content }),

  // Backups
  listBackups: (id) => api.get(`/servers/${encodeURIComponent(id)}/backups`),
  createBackup: (id) => api.post(`/servers/${encodeURIComponent(id)}/backups`),
  deleteBackup: (id, name) =>
    api.delete(`/servers/${encodeURIComponent(id)}/backups/${encodeURIComponent(name)}`),

  // Schedules
  listSchedules: () => api.get("/schedules"),
  createSchedule: (data) => api.post("/schedules", data),
  toggleSchedule: (id, enabled) => api.put(`/schedules/${encodeURIComponent(id)}/toggle`, { enabled }),
  deleteSchedule: (id) => api.delete(`/schedules/${encodeURIComponent(id)}`),

  // System
  getAuditLogs: () => api.get("/audit-logs"),
  getSettings: () => api.get("/settings"),
  updateSettings: (data) => api.put("/settings", data),
};
