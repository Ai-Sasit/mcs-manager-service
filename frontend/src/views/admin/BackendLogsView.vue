<template>
  <div class="backend-logs-view">
    <div class="toolbar card">
      <div class="toolbar-left">
        <el-select
          v-model="selectedFile"
          placeholder="Select log file"
          style="width: 280px"
          @change="loadLines"
        >
          <el-option v-for="f in files" :key="f" :label="f" :value="f" />
        </el-select>
        <el-select v-model="tailCount" style="width: 140px" @change="loadLines">
          <el-option label="Last 100 lines" :value="100" />
          <el-option label="Last 500 lines" :value="500" />
          <el-option label="Last 1000 lines" :value="1000" />
          <el-option label="Last 2000 lines" :value="2000" />
          <el-option label="Last 5000 lines" :value="5000" />
        </el-select>
        <el-input
          v-model="search"
          placeholder="Filter lines..."
          clearable
          style="width: 220px"
        />
      </div>
      <div class="toolbar-right">
        <el-switch
          v-model="isLive"
          active-text="Live Logs"
          inactive-text=""
          @change="toggleLive"
        />
        <el-checkbox v-model="autoScroll">Auto-scroll</el-checkbox>
        <el-button @click="loadLines" :loading="loading"
          ><el-icon><Refresh /></el-icon>Refresh</el-button
        >
        <el-button @click="clearView"
          ><el-icon><Delete /></el-icon>Clear</el-button
        >
      </div>
    </div>
    <div class="stats-bar" v-if="selectedFile">
      <span class="stat"
        ><strong>{{ filteredLines.length }}</strong>
        {{ search ? "matching" : "total" }} lines</span
      >
      <span class="stat" v-if="search"
        >of <strong>{{ lines.length }}</strong> loaded</span
      >
      <span class="stat-sep" v-if="selectedFile">·</span>
      <span class="stat">{{ selectedFile }}</span>
      <el-tag
        v-if="isLive"
        type="danger"
        size="small"
        effect="dark"
        class="live-tag"
        >LIVE</el-tag
      >
    </div>
    <div class="log-card card">
      <div
        class="log-box"
        ref="logBox"
        v-loading="loading"
        element-loading-text="Loading logs..."
      >
        <template v-if="filteredLines.length > 0">
          <div
            v-for="(line, i) in filteredLines"
            :key="i"
            :class="['log-line', lineClass(line)]"
          >
            <span class="line-num">{{ lineOffset + i + 1 }}</span>
            <span class="line-text">{{ line }}</span>
          </div>
        </template>
        <div v-else-if="!loading" class="log-empty">
          {{
            selectedFile
              ? "No lines match the current filter."
              : "Select a log file above to view its contents."
          }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
import { Refresh, Delete } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { getToken } from "@/utils/authStorage";
import { WS_BASE_URL } from "@/constants";
import { getApiErrorMessage } from "@/utils/apiError";

const files = ref([]);
const selectedFile = ref("");
const tailCount = ref(500);
const search = ref("");
const lines = ref([]);
const loading = ref(false);
const autoScroll = ref(true);
const isLive = ref(false);
const socket = ref(null);
const logBox = ref(null);
const lineOffset = computed(() => 0);

const filteredLines = computed(() => {
  if (!search.value.trim()) return lines.value;
  const q = search.value.toLowerCase();
  return lines.value.filter((l) => l.toLowerCase().includes(q));
});

function lineClass(line) {
  if (/\bERROR\b|\bFATAL\b/i.test(line)) return "log-error";
  if (/\bWARN\b|\bWARNING\b/i.test(line)) return "log-warn";
  if (/\bINFO\b/i.test(line)) return "log-info";
  if (/\bDEBUG\b/i.test(line)) return "log-debug";
  return "";
}

async function loadFiles() {
  try {
    const { data } = await apiClient.get("/backend-logs");
    files.value = data.data || [];
    if (files.value.length > 0 && !selectedFile.value) {
      selectedFile.value = files.value[0];
      await loadLines();
    }
  } catch (e) {
    ElMessage.error("Failed to load log file list: " + getApiErrorMessage(e));
  }
}

async function loadLines() {
  if (!selectedFile.value) return;
  loading.value = true;
  try {
    const { data } = await apiClient.get(
      `/backend-logs/file?file=${encodeURIComponent(selectedFile.value)}&tail=${tailCount.value}`,
    );
    lines.value = data.data || [];
    scrollToBottom();
  } catch (e) {
    ElMessage.error("Failed to load log file: " + getApiErrorMessage(e));
    lines.value = [];
  } finally {
    loading.value = false;
  }
}

function scrollToBottom() {
  if (autoScroll.value) {
    nextTick(() => {
      if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
    });
  }
}

function toggleLive() {
  if (isLive.value) {
    connectLive();
  } else {
    disconnectLive();
  }
}

function connectLive() {
  const token = getToken();
  if (!token) {
    isLive.value = false;
    return;
  }
  const url = `${WS_BASE_URL}/ws/backend-logs?token=${token}`;
  socket.value = new WebSocket(url);
  socket.value.onopen = () => {
    ElMessage.success("Connected to live logs");
  };
  socket.value.onmessage = (event) => {
    lines.value.push(event.data);
    if (lines.value.length > tailCount.value) lines.value.shift();
    scrollToBottom();
  };
  socket.value.onclose = () => {
    isLive.value = false;
    socket.value = null;
  };
  socket.value.onerror = () => {
    isLive.value = false;
  };
}

function disconnectLive() {
  if (socket.value) {
    socket.value.close();
    socket.value = null;
  }
}

function clearView() {
  lines.value = [];
}
watch(filteredLines, scrollToBottom);
onMounted(loadFiles);
onUnmounted(disconnectLive);
</script>

<style scoped>
.backend-logs-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px 20px;
}
.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.stats-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  padding: 0 4px;
}
.stat strong {
  color: var(--el-text-color-primary);
}
.stat-sep {
  color: var(--el-border-color);
}
.live-tag {
  margin-left: 8px;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
  100% {
    opacity: 1;
  }
}
.log-card {
  padding: 0;
  overflow: hidden;
}
.log-box {
  height: calc(100vh - 280px);
  min-height: 400px;
  overflow-y: auto;
  padding: 16px 20px;
  font-family: "JetBrains Mono", "Fira Code", "Consolas", monospace;
  font-size: 12.5px;
  line-height: 1.65;
  background: #0f1117;
  color: #d1d5db;
}
.log-line {
  display: flex;
  gap: 12px;
  padding: 1px 0;
}
.log-line:hover {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 3px;
}
.line-num {
  color: #4b5563;
  user-select: none;
  min-width: 44px;
  text-align: right;
  flex-shrink: 0;
}
.line-text {
  white-space: pre-wrap;
  word-break: break-all;
  flex: 1;
}
.log-empty {
  color: #4b5563;
  font-style: italic;
  text-align: center;
  margin-top: 60px;
}
.log-error .line-text {
  color: #f87171;
}
.log-warn .line-text {
  color: #fbbf24;
}
.log-info .line-text {
  color: #d1d5db;
}
.log-debug .line-text {
  color: #6b7280;
}
</style>
