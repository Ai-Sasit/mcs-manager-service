<template>
  <div class="logs-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Backend Logs</h2>
        <p class="page-subtitle">Monitor application-level logs.</p>
      </div>
      <div class="header-actions">
        <el-button @click="fetchLogs" :loading="loading">
          <template #icon><PhArrowsClockwise /></template>Refresh</el-button
        >
        <el-button @click="clearView">
          <template #icon><PhTrash /></template>Clear</el-button
        >
      </div>
    </div>

    <div class="card" style="padding: 24px">
      <div class="toolbar">
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
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
import { PhArrowsClockwise, PhTrash } from "@phosphor-icons/vue";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";
import { ElMessage } from "element-plus";

const lines = ref([]);
const filter = ref("");
const logLevel = ref("");
const loading = ref(false);
const logBox = ref(null);
let socket = null;

const filteredLines = computed(() => {
  let result = lines.value;
  if (logLevel.value) {
    const level = logLevel.value.toUpperCase();
    result = result.filter((l) => l.includes(`[${level}]`));
  }
  if (filter.value) {
    const q = filter.value.toLowerCase();
    result = result.filter((l) => l.toLowerCase().includes(q));
  }
  return result;
});

function logLineClass(line) {
  if (/\[ERROR\]/i.test(line)) return "log-error";
  if (/\[WARN\]/i.test(line)) return "log-warn";
  return "";
}

function scrollToBottom() {
  nextTick(() => {
    if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
  });
}

function connectSocket() {
  const protocol = location.protocol === "https:" ? "wss" : "ws";
  const url = `${protocol}://${location.host}/ws/backend-logs`;
  socket = new WebSocket(url);

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      if (msg.line) {
        lines.value.push(msg.line);
        if (lines.value.length > 1000) lines.value.shift();
        scrollToBottom();
      }
    } catch {
      lines.value.push(event.data);
      if (lines.value.length > 1000) lines.value.shift();
      scrollToBottom();
    }
  };

  socket.onclose = () => {
    setTimeout(() => connectSocket(), 3000);
  };
}

async function fetchLogs() {
  loading.value = true;
  try {
    const { data } = await apiClient.get("/system/logs");
    lines.value = data.data || [];
    scrollToBottom();
  } catch (e) {
    ElMessage.error("Failed to fetch logs: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

function clearView() {
  lines.value = [];
}

onMounted(() => {
  fetchLogs();
  connectSocket();
});

onUnmounted(() => {
  if (socket) socket.close();
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
  gap: 12px;
  margin-bottom: 16px;
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
