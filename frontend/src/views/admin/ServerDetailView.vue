<template>
  <div v-if="server" class="detail-page">
    <!-- Header -->
    <div class="detail-header card">
      <div class="header-left">
        <el-button
          text
          size="small"
          :icon="ArrowLeft"
          @click="$router.push('/')">
          Back
        </el-button>
        <div class="server-info">
          <h2 class="server-name">{{ server.name }}</h2>
          <div class="meta-row">
            <el-tag
              size="small"
              :type="server.status === 'running' ? 'success' : 'info'"
              effect="light">
              {{ server.status }}
            </el-tag>
            <span class="meta-text">v{{ server.version }}</span>
            <span class="meta-sep">·</span>
            <span class="meta-text">Port {{ server.port }}</span>
            <template v-if="server.edition === 'java'">
              <span class="meta-sep">·</span>
              <span class="meta-text">{{ server.memory_mb }}MB RAM</span>
            </template>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <el-button
          v-if="server.status === 'stopped'"
          type="primary"
          :loading="loading"
          :icon="VideoPlay"
          @click="start">
          Start
        </el-button>
        <el-button v-else :loading="loading" :icon="VideoPause" @click="stop">
          Stop
        </el-button>
        <el-button
          :disabled="loading || server.status === 'stopped'"
          :icon="RefreshRight"
          @click="restart">
          Restart
        </el-button>
        <el-button
          v-if="server.status !== 'stopped'"
          type="danger"
          plain
          :loading="loading"
          :icon="CircleClose"
          @click="kill">
          Kill
        </el-button>
        <el-button
          type="danger"
          :loading="loading"
          :icon="Delete"
          @click="confirmDelete">
          Delete
        </el-button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="detail-tabs card">
      <el-tabs v-model="activeTab">
        <!-- Overview -->
        <el-tab-pane label="Overview" name="overview">
          <div class="tab-content">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Instance ID">
                <code>{{ server.id }}</code>
              </el-descriptions-item>
              <el-descriptions-item label="Edition">
                <span class="edition-badge" :class="server.edition">
                  {{
                    server.edition === "java"
                      ? "Java Edition"
                      : "Bedrock Edition"
                  }}
                </span>
              </el-descriptions-item>
              <el-descriptions-item
                label="Server Type"
                v-if="server.edition === 'java'">
                {{
                  { vanilla: "Vanilla", paper: "Paper", spigot: "Spigot" }[
                    server.server_type
                  ] || "Vanilla"
                }}
              </el-descriptions-item>
              <el-descriptions-item label="Version">{{
                server.version
              }}</el-descriptions-item>
              <el-descriptions-item label="Port">{{
                server.port
              }}</el-descriptions-item>
              <el-descriptions-item label="Max Players">{{
                server.max_players
              }}</el-descriptions-item>
              <el-descriptions-item
                label="Memory"
                v-if="server.edition === 'java'">
                {{ server.memory_mb }} MB
              </el-descriptions-item>
              <el-descriptions-item label="Created">{{
                formatDate(server.created_at)
              }}</el-descriptions-item>
              <el-descriptions-item label="Path">
                <code class="path-code">{{ server.server_dir }}</code>
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </el-tab-pane>

        <!-- Logs -->
        <el-tab-pane label="Live Logs" name="logs">
          <div class="tab-content">
            <div class="tab-toolbar">
              <el-button size="small" :icon="Remove" @click="clearLogs"
                >Clear Logs</el-button
              >
              <el-checkbox v-model="autoScroll">Auto-scroll</el-checkbox>
            </div>
            <div class="log-box" ref="logBox">
              <div
                v-for="(line, i) in logLines"
                :key="i"
                :class="['log-line', logLineClass(line)]">
                {{ line }}
              </div>
              <div v-if="logLines.length === 0" class="log-empty">
                {{
                  server.status === "running"
                    ? "Waiting for log stream..."
                    : "Start the instance to view logs."
                }}
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- Terminal -->
        <el-tab-pane label="Terminal" name="terminal">
          <div class="tab-content">
            <div class="log-box" ref="termBox">
              <div
                v-for="(line, i) in termLines"
                :key="i"
                :class="['log-line', logLineClass(line)]">
                {{ line }}
              </div>
              <div v-if="termLines.length === 0" class="log-empty">
                {{
                  server.status === "running"
                    ? "Connection established. Waiting for input..."
                    : "Start the server to use the terminal."
                }}
              </div>
            </div>
            <div class="cmd-row">
              <el-input
                v-model="cmdInput"
                placeholder="Enter server command..."
                :disabled="server.status !== 'running'"
                @keyup.enter="sendCommand">
                <template #prepend><span class="cmd-prompt">/</span></template>
              </el-input>
              <el-button
                type="primary"
                :disabled="server.status !== 'running'"
                @click="sendCommand">
                Execute
              </el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- Config -->
        <el-tab-pane label="Config" name="config">
          <div class="tab-content">
            <div v-if="configLoading" class="center-info">
              <el-icon class="spin" size="24"><Loading /></el-icon>
              <p>Reading configuration...</p>
            </div>
            <div v-else-if="configError" class="center-info">
              <el-alert :title="configError" type="error" :closable="false" />
            </div>
            <div v-else>
              <div class="config-toolbar">
                <el-input
                  v-model="configSearch"
                  placeholder="Search keys..."
                  clearable
                  style="width: 300px"
                  size="small" />
                <div style="display: flex; gap: 8px">
                  <el-button size="small" @click="loadConfig">Reload</el-button>
                  <el-button
                    type="primary"
                    size="small"
                    :loading="configSaving"
                    :icon="CircleCheck"
                    @click="saveConfig">
                    Save Changes
                  </el-button>
                </div>
              </div>
              <el-table :data="filteredConfig" style="width: 100%" size="small">
                <el-table-column prop="key" label="Property" width="300" />
                <el-table-column label="Value">
                  <template #default="{ row }">
                    <el-input v-model="row.value" size="small" />
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </el-tab-pane>

        <!-- Plugins -->
        <el-tab-pane label="Addons & Plugins" name="plugins">
          <div class="tab-content">
            <PluginUpload
              :server-id="server.id"
              :edition="server.edition"
              @uploaded="loadPlugins" />

            <el-table
              :data="plugins"
              style="width: 100%; margin-top: 20px"
              size="small"
              v-if="plugins.length > 0">
              <el-table-column label="File Name" prop="name" />
              <el-table-column label="File Size" width="120">
                <template #default="{ row }">{{
                  formatSize(row.size)
                }}</template>
              </el-table-column>
              <el-table-column label="Action" width="80" align="center">
                <template #default="{ row }">
                  <el-button
                    type="danger"
                    circle
                    size="small"
                    :icon="Delete"
                    @click="removePlugin(row.name)" />
                </template>
              </el-table-column>
            </el-table>
            <el-empty
              v-else
              description="No plugins installed."
              :image-size="80" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>

  <div v-else-if="loadingServer" class="center-info full-page">
    <el-icon class="spin" size="32"><Loading /></el-icon>
    <p>Loading instance data...</p>
  </div>

  <div v-else class="center-info full-page">
    <el-result
      icon="error"
      title="Instance Not Found"
      sub-title="The requested server instance does not exist or has been deleted.">
      <template #extra>
        <el-button type="primary" :icon="ArrowLeft" @click="$router.push('/')"
          >Return Home</el-button
        >
      </template>
    </el-result>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  ArrowLeft,
  VideoPlay,
  VideoPause,
  RefreshRight,
  Delete,
  Loading,
  CircleCheck,
  CircleClose,
  Remove,
} from "@element-plus/icons-vue";
import api from "@/api";
import { useServerLogs } from "@/composables/useServerLogs";
import { useServerTerminal } from "@/composables/useServerTerminal";
import PluginUpload from "@/components/servers/PluginUpload.vue";
import { ElMessage, ElMessageBox } from "element-plus";

const route = useRoute();
const router = useRouter();

const server = ref(null);
const plugins = ref([]);
const loading = ref(false);
const loadingServer = ref(true);
const activeTab = ref("overview");

const logLines = ref([]);
const autoScroll = ref(true);
const logBox = ref(null);
let logsComp = null;

const termLines = ref([]);
const cmdInput = ref("");
const termBox = ref(null);
let termComp = null;

const configRows = ref([]);
const configSearch = ref("");
const configLoading = ref(false);
const configError = ref("");
const configSaving = ref(false);

const filteredConfig = computed(() => {
  if (!configSearch.value) return configRows.value;
  const q = configSearch.value.toLowerCase();
  return configRows.value.filter((r) => r.key.toLowerCase().includes(q));
});

async function loadServer() {
  loadingServer.value = true;
  try {
    const { data } = await api.getServer(route.params.id);
    server.value = data.data;
  } catch {
    server.value = null;
  } finally {
    loadingServer.value = false;
  }
}

async function loadPlugins() {
  if (!server.value) return;
  try {
    const { data } = await api.listPlugins(server.value.id);
    plugins.value = data.data || [];
  } catch {
    plugins.value = [];
  }
}

async function loadConfig() {
  configLoading.value = true;
  configError.value = "";
  try {
    const { data } = await api.getConfig(server.value.id);
    const map = data.data || {};
    configRows.value = Object.entries(map).map(([key, value]) => ({
      key,
      value,
    }));
  } catch (e) {
    configError.value = e.response?.data?.message || "Failed to load config";
  } finally {
    configLoading.value = false;
  }
}

async function saveConfig() {
  configSaving.value = true;
  try {
    const map = Object.fromEntries(
      configRows.value.map((r) => [r.key, r.value]),
    );
    await api.updateConfig(server.value.id, map);
    ElMessage.success({ message: "Configuration updated.", type: "success" });
  } catch (e) {
    ElMessage.error(e.response?.data?.message || "Failed to save config");
  } finally {
    configSaving.value = false;
  }
}

function scrollToBottom(el) {
  nextTick(() => {
    if (el) el.scrollTop = el.scrollHeight;
  });
}

function setupLogs() {
  if (logsComp) logsComp.disconnect();
  logsComp = useServerLogs(server.value.id);
  logsComp.onLine((line) => {
    logLines.value.push(line);
    if (logLines.value.length > 2000) logLines.value.shift();
    if (autoScroll.value) scrollToBottom(logBox.value);
  });
  logsComp.connect();
}

function setupTerminal() {
  if (termComp) termComp.disconnect();
  termComp = useServerTerminal(server.value.id);
  termComp.onLine((line) => {
    termLines.value.push(line);
    if (termLines.value.length > 2000) termLines.value.shift();
    scrollToBottom(termBox.value);
  });
  termComp.connect();
}

function clearLogs() {
  logLines.value = [];
}

function sendCommand() {
  if (!cmdInput.value.trim() || !termComp) return;
  termComp.sendCommand(cmdInput.value.trim());
  cmdInput.value = "";
}

async function start() {
  loading.value = true;
  try {
    await api.startServer(server.value.id);
    server.value.status = "running";
    ElMessage.success("Server start sequence initiated.");
  } catch (e) {
    ElMessage.error("Failed: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
  }
}

async function stop() {
  loading.value = true;
  try {
    await api.stopServer(server.value.id);
    server.value.status = "stopped";
    logsComp?.disconnect();
    termComp?.disconnect();
    ElMessage.info("Server stopped.");
  } catch (e) {
    ElMessage.error("Failed: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
  }
}

async function restart() {
  loading.value = true;
  server.value.status = "starting";
  try {
    logsComp?.disconnect();
    termComp?.disconnect();
    await api.restartServer(server.value.id);
    server.value.status = "running";
    setupLogs();
    setupTerminal();
    ElMessage.success("Server restarted.");
  } catch (e) {
    ElMessage.error("Failed: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
  }
}

async function kill() {
  try {
    await ElMessageBox.confirm(
      "Force kill the server process? This may result in data loss as the server will not save its state.",
      "Force Kill",
      {
        confirmButtonText: "Kill Process",
        cancelButtonText: "Cancel",
        type: "error",
      }
    );
    loading.value = true;
    await api.killServer(server.value.id);
    server.value.status = "stopped";
    logsComp?.disconnect();
    termComp?.disconnect();
    ElMessage.success("Process terminated.");
  } catch (e) {
    if (e !== "cancel") {
      ElMessage.error("Failed: " + (e.response?.data?.message || e.message));
    }
  } finally {
    loading.value = false;
  }
}

async function confirmDelete() {
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to delete "${server.value.name}"? This action cannot be undone.`,
      "Warning",
      {
        confirmButtonText: "Yes, Delete Forever",
        cancelButtonText: "Cancel",
        type: "warning",
        confirmButtonClass: "el-button--danger",
      },
    );
    loading.value = true;
    await api.deleteServer(server.value.id);
    ElMessage.success("Server instance deleted.");
    router.push("/");
  } catch (e) {
    if (e !== "cancel") {
      ElMessage.error("Failed: " + (e.response?.data?.message || e.message));
      loading.value = false;
    }
  }
}

async function removePlugin(name) {
  try {
    await ElMessageBox.confirm(`Remove this plugin/addon?`, "Confirm", {
      confirmButtonText: "Remove",
      cancelButtonText: "Cancel",
      type: "warning",
    });
    await api.deletePlugin(server.value.id, name);
    ElMessage.success("Plugin removed successfully.");
    loadPlugins();
  } catch (e) {
    if (e !== "cancel") ElMessage.error(e.response?.data?.message || e.message);
  }
}

watch(activeTab, (tab) => {
  if (tab === "config") {
    console.log("Loading config");
    loadConfig();
  }
});

watch(
  () => server.value?.status,
  (newStatus, oldStatus) => {
    if (!server.value) return;
    if (newStatus === "running" && oldStatus !== "running") {
      setupLogs();
      setupTerminal();
    } else if (newStatus === "stopped" && oldStatus === "running") {
      logsComp?.disconnect();
      termComp?.disconnect();
    }
  },
);

function formatDate(iso) {
  if (!iso) return "-";
  return new Date(iso).toLocaleString();
}

function formatSize(bytes) {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / 1048576).toFixed(1) + " MB";
}

function logLineClass(line) {
  if (line.includes("[SYSTEM/INFO]")) return "log-system-info";
  if (line.includes("[SYSTEM/WARN]")) return "log-system-warn";
  if (line.includes("[SYSTEM/ERROR]")) return "log-system-error";
  if (/\[WARN\]|\bWARNING\b/i.test(line)) return "log-warn";
  if (/\[ERROR\]|\bERROR\b|\bFATAL\b/i.test(line)) return "log-error";
  return "";
}

onMounted(async () => {
  await loadServer();
  if (server.value) {
    loadPlugins();
    if (server.value.status === "running") {
      setupLogs();
      setupTerminal();
    }
  }
});

onUnmounted(() => {
  logsComp?.disconnect();
  termComp?.disconnect();
});
</script>

<style scoped>
.detail-page {
  animation: fadeIn 0.3s ease-out;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ─── Header ─── */
.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.server-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.server-name {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.meta-text {
  color: var(--color-text-muted);
  font-weight: 500;
}

.meta-sep {
  color: var(--color-border);
}

.header-actions {
  display: flex;
  gap: 8px;
}

/* ─── Tabs Card ─── */
.detail-tabs {
  padding: 24px;
}

.tab-content {
  padding: 20px 0 0;
}

.tab-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

/* ─── Log Box ─── */
.log-box {
  background: var(--color-white);
  border-radius: var(--radius);
  padding: 20px;
  height: 480px;
  overflow-y: auto;
  font-family: "JetBrains Mono", "Fira Code", "Consolas", monospace;
  font-size: 13px;
  line-height: 1.7;
  border: 1px solid var(--color-border);
}

.log-line {
  color: var(--color-text);
  margin-bottom: 2px;
}

.log-system-info {
  color: #3b82f6;
  font-weight: 600;
}

.log-system-warn {
  color: #f59e0b;
  font-weight: 600;
}

.log-system-error {
  color: #ef4444;
  font-weight: 600;
}

.log-warn {
  color: #f59e0b;
}

.log-error {
  color: #ef4444;
}

.log-empty {
  color: #6b7280;
  font-style: italic;
}

.cmd-row {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.cmd-prompt {
  color: var(--color-primary);
  font-weight: 700;
  font-size: 16px;
}

/* ─── Config ─── */
.config-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 16px;
}

/* ─── States ─── */
.center-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--color-text-secondary);
}

.full-page {
  height: 60vh;
  justify-content: center;
}

.edition-badge {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.edition-badge.java {
  background: #fff7ed;
  color: #ea580c;
}

.edition-badge.bedrock {
  background: #eff6ff;
  color: #2563eb;
}

.path-code {
  background: var(--color-bg);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
}

.spin {
  animation: rotate 1.5s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
