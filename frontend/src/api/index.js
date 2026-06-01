import apiClient from "./client";

class ApiService {
  constructor() {
    this.https = apiClient;
  }

  // ─── Auth ───
  async login(username, password) {
    return this.https.post("/auth/login", { username, password });
  }

  async logout() {
    return this.https.post("/auth/logout");
  }

  async getMe() {
    return this.https.get("/auth/me");
  }

  async changePassword(currentPassword, newPassword) {
    return this.https.put("/auth/me/password", {
      current_password: currentPassword,
      new_password: newPassword,
    });
  }

  // ─── Servers ───
  async listServers() {
    return this.https.get("/servers");
  }

  async createServer(payload) {
    return this.https.post("/servers", payload);
  }

  async getServer(id) {
    return this.https.get(`/servers/${encodeURIComponent(id)}`);
  }

  async deleteServer(id) {
    return this.https.delete(`/servers/${encodeURIComponent(id)}`);
  }

  async startServer(id) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/start`);
  }

  async stopServer(id) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/stop`);
  }

  async restartServer(id) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/restart`);
  }

  async killServer(id) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/kill`);
  }

  // ─── Config ───
  async getConfig(id) {
    return this.https.get(`/servers/${encodeURIComponent(id)}/config`);
  }

  async updateConfig(id, config) {
    return this.https.put(`/servers/${encodeURIComponent(id)}/config`, config);
  }

  // ─── Plugins ───
  async listPlugins(id) {
    return this.https.get(`/servers/${encodeURIComponent(id)}/plugins`);
  }

  async uploadPlugin(id, formData) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/plugins`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  }

  async deletePlugin(id, name) {
    return this.https.delete(
      `/servers/${encodeURIComponent(id)}/plugins/${encodeURIComponent(name)}`,
    );
  }

  // ─── Versions ───
  async getJavaVersions() {
    return this.https.get("/versions/java");
  }

  async getBedrockVersions() {
    return this.https.get("/versions/bedrock");
  }

  // ─── Players ───
  async getPlayers(id) {
    return this.https.get(`/servers/${encodeURIComponent(id)}/players`);
  }

  async updatePlayer(id, data) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/players`, data);
  }

  // ─── Files ───
  async listFiles(id, path) {
    return this.https.get(
      `/servers/${encodeURIComponent(id)}/files?path=${encodeURIComponent(path || "")}`,
    );
  }

  async readFile(id, path) {
    return this.https.get(
      `/servers/${encodeURIComponent(id)}/files/read?path=${encodeURIComponent(path)}`,
    );
  }

  async writeFile(id, path, content) {
    return this.https.post(
      `/servers/${encodeURIComponent(id)}/files/write?path=${encodeURIComponent(path)}`,
      { content },
    );
  }

  // ─── Backups ───
  async listBackups(id) {
    return this.https.get(`/servers/${encodeURIComponent(id)}/backups`);
  }

  async createBackup(id) {
    return this.https.post(`/servers/${encodeURIComponent(id)}/backups`);
  }

  async deleteBackup(id, name) {
    return this.https.delete(
      `/servers/${encodeURIComponent(id)}/backups/${encodeURIComponent(name)}`,
    );
  }

  // ─── Schedules ───
  async listSchedules() {
    return this.https.get("/schedules");
  }

  async createSchedule(payload) {
    return this.https.post("/schedules", payload);
  }

  async toggleSchedule(id, enabled) {
    return this.https.put(`/schedules/${encodeURIComponent(id)}/toggle`, { enabled });
  }

  async deleteSchedule(id) {
    return this.https.delete(`/schedules/${encodeURIComponent(id)}`);
  }

  // ─── System ───
  async getSystemInfo() {
    return this.https.get("/system/info");
  }

  async lookupPort(port) {
    return this.https.get(`/system/port-lookup?port=${port}`);
  }

  async killPid(pid) {
    return this.https.post("/system/kill-pid", { pid });
  }

  async getAuditLogs() {
    return this.https.get("/audit-logs");
  }

  async listBackendLogFiles() {
    return this.https.get("/backend-logs");
  }

  async getBackendLogFile(file, tail = 500) {
    return this.https.get(`/backend-logs/file?file=${encodeURIComponent(file)}&tail=${tail}`);
  }

  // ─── Users ───
  async listUsers() {
    return this.https.get("/users");
  }

  async createUser(data) {
    return this.https.post("/users", data);
  }

  async updateUser(id, data) {
    return this.https.put(`/users/${encodeURIComponent(id)}`, data);
  }

  async deleteUser(id) {
    return this.https.delete(`/users/${encodeURIComponent(id)}`);
  }

  // ─── Settings ───
  async getSettings() {
    return this.https.get("/settings");
  }

  async updateSettings(data) {
    return this.https.put("/settings", data);
  }
}

export default {
  install: (app) => {
    app.config.globalProperties.$api = new ApiService();
  },
};