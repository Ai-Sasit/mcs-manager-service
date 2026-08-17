<template>
  <div class="logs-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Backend Logs</h2>
        <p class="page-subtitle">Monitor application-level logs.</p>
      </div>
      <div class="header-actions">
        <el-button @click="refreshLogs" :loading="loading">
          <template #icon><PhArrowsClockwise /></template>Refresh</el-button
        >
        <el-button @click="clearView">
          <template #icon><PhTrash /></template>Clear</el-button
        >
      </div>
    </div>

    <div class="card" style="padding: 24px">
      <div class="toolbar">
        <div class="filter-tools">
          <el-input
            v-model="filter"
            placeholder="Filter lines..."
            size="small"
            clearable
            style="width: 300px"
          />
          <el-select v-model="logLevel" size="small" style="width: 140px">
            <el-option label="All" value="" />
            <el-option label="INFO" value="info" />
            <el-option label="WARN" value="warn" />
            <el-option label="ERROR" value="error" />
          </el-select>
        </div>
        <el-tag size="small" :type="connectionTagType">
          {{ connectionLabel }}
        </el-tag>
      </div>
      <el-alert
        v-if="streamNotice"
        :title="streamNotice"
        type="warning"
        show-icon
        closable
        @close="streamNotice = ''"
      />
      <div class="log-box" ref="logBox">
        <div
          v-for="(line, i) in filteredLines"
          :key="i"
          :class="['log-line', logLineClass(line)]"
        >
          {{ line }}
        </div>
        <div v-if="filteredLines.length === 0" class="log-empty">
          No log entries match the current filter.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from "vue";
import { PhArrowsClockwise, PhTrash } from "@phosphor-icons/vue";
import { useBackendLogs } from "@/composables/useBackendLogs";

const lines = ref([]);
const filter = ref("");
const logLevel = ref("");
const loading = ref(false);
const logBox = ref(null);
const connectionState = ref("idle");
const streamNotice = ref("");
let socket = null;

const connectionLabel = computed(() => {
  if (connectionState.value === "open") return "LIVE";
  return connectionState.value.toUpperCase();
});

const connectionTagType = computed(() => {
  if (connectionState.value === "open") return "success";
  if (
    connectionState.value === "error" ||
    connectionState.value === "auth-error"
  ) {
    return "danger";
  }
  return "warning";
});

const filteredLines = computed(() => {
  let result = lines.value;
  if (logLevel.value) {
    const level = logLevel.value.toUpperCase();
    const pattern = new RegExp(`\\b${level}\\b`, "i");
    result = result.filter((line) => pattern.test(line));
  }
  if (filter.value) {
    const q = filter.value.toLowerCase();
    result = result.filter((l) => l.toLowerCase().includes(q));
  }
  return result;
});

function logLineClass(line) {
  if (/\bERROR\b|\bFATAL\b/i.test(line)) return "log-error";
  if (/\bWARN\b|\bWARNING\b/i.test(line)) return "log-warn";
  return "";
}

function scrollToBottom() {
  nextTick(() => {
    if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
  });
}

function connectSocket() {
  socket = useBackendLogs();
  socket.onStateChange((state) => {
    connectionState.value = state;
    loading.value = state === "connecting" || state === "reconnecting";
  });
  socket.onLine((line) => {
    lines.value.push(line);
    if (lines.value.length > 10000) lines.value.shift();
    scrollToBottom();
  });
  socket.onControl((message) => {
    if (message.type === "stream_reset") lines.value = [];
    streamNotice.value =
      message.data || "Some backend log entries could not be recovered.";
  });
  socket.onError((message) => {
    if (connectionState.value === "error") streamNotice.value = message;
  });
  socket.connect();
  connectionState.value = socket.state.value;
}

function refreshLogs() {
  streamNotice.value = "";
  lines.value = [];
  loading.value = true;
  socket?.resetReplay();
  if (socket) {
    connectionState.value = socket.state.value;
    loading.value =
      socket.state.value === "connecting" ||
      socket.state.value === "reconnecting";
  }
}

function clearView() {
  lines.value = [];
}

onMounted(() => {
  connectSocket();
});

onUnmounted(() => {
  socket?.disconnect();
});
</script>

<style scoped>
.logs-page {
  animation: fadeIn 0.3s ease-out;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
}

.page-subtitle {
  margin: 0;
  color: var(--color-text-secondary);
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.filter-tools {
  display: flex;
  gap: 12px;
}

.el-alert {
  margin-bottom: 12px;
}

.log-box {
  background: #111815;
  border: 1px solid #25312d;
  border-radius: 8px;
  padding: 16px;
  height: 500px;
  overflow-y: auto;
  font-family: "JetBrains Mono", "Fira Code", monospace;
  font-size: 13px;
  line-height: 1.7;
}

.log-line {
  color: #e2e8e5;
}

.log-warn {
  color: #ffd87a;
}

.log-error {
  color: #ff99a4;
}

.log-empty {
  color: #9aa8a2;
  font-style: italic;
  text-align: center;
  padding-top: 40px;
}
</style>
