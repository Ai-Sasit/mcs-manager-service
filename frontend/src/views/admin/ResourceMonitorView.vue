<template>
  <div class="resource-page">
    <div class="resource-header">
      <div>
        <h2>Resource Monitor</h2>
        <p>Live host and backend runtime metrics.</p>
      </div>
      <el-tag
        :type="connectionTagType"
        effect="dark"
        round
        :title="resources.error.value"
      >
        {{ connectionLabel }}
      </el-tag>
    </div>

    <el-tabs v-model="activeTab" class="resource-tabs">
      <el-tab-pane label="Current Stats" name="stats">
        <div class="stats-grid">
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Node Load</span>
              <span class="stat-meta">{{ cpuPercent }}%</span>
            </div>
            <div class="stat-value">{{ cpuPercent }}%</div>
            <el-progress
              :percentage="cpuPercent"
              :show-text="false"
              stroke-width="8"
              :color="cpuColors"
            />
          </div>

          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Memory Usage</span>
              <span class="stat-meta">{{ memoryPercent }}%</span>
            </div>
            <div class="stat-value">{{ formatMB(node.memory_used_mb) }}</div>
            <div class="stat-sub">of {{ formatMB(node.memory_total_mb) }}</div>
            <el-progress
              :percentage="memoryPercent"
              :show-text="false"
              stroke-width="8"
              color="#10b981"
            />
          </div>

          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Disk Usage</span>
              <span class="stat-meta">{{ diskPercent }}%</span>
            </div>
            <div class="stat-value">{{ formatMB(node.disk_used_mb) }}</div>
            <div class="stat-sub">of {{ formatMB(node.disk_total_mb) }}</div>
            <el-progress
              :percentage="diskPercent"
              :show-text="false"
              stroke-width="8"
              color="#3b82f6"
            />
          </div>
        </div>

        <div class="stats-grid runtime-grid">
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Operating System</span>
            </div>
            <div class="stat-value text-sm">
              {{ runtime.os || "—" }} / {{ runtime.arch || "—" }}
            </div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">CPU Cores</span>
            </div>
            <div class="stat-value">{{ runtime.cpu_count || "—" }}</div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Load Average</span>
            </div>
            <div class="stat-value text-sm">{{ loadAverage }}</div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Go Version</span>
            </div>
            <div class="stat-value text-sm">
              {{ runtime.go_version || "—" }}
            </div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Goroutines</span>
            </div>
            <div class="stat-value">{{ runtime.goroutines || "—" }}</div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">GC Cycles</span>
            </div>
            <div class="stat-value">{{ runtime.mem_gc_cycles ?? "—" }}</div>
            <div class="stat-sub">
              Go memory: {{ formatMB(runtime.mem_alloc_mb) }} allocated /
              {{ formatMB(runtime.mem_sys_mb) }} system
            </div>
          </div>
        </div>

        <div class="stats-grid runtime-grid">
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Servers</span>
            </div>
            <div class="stat-value">
              {{ servers.running || 0 }} / {{ servers.total || 0 }}
            </div>
            <div class="stat-sub">running / total</div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Allocated Server Memory</span>
            </div>
            <div class="stat-value">
              {{ formatMB(servers.allocated_memory_mb) }}
            </div>
          </div>
          <div class="stat-card card">
            <div class="stat-header">
              <span class="stat-title">Updated</span>
            </div>
            <div class="stat-value text-sm">{{ lastUpdatedLabel }}</div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="Historical Graphs" name="graphs">
        <ResourceGraphsView
          :history="resourceHistory.history"
          :time-labels="resourceHistory.timeLabels"
          :data-point-count="resourceHistory.dataPointCount"
          :time-range-seconds="resourceHistory.timeRangeSeconds"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import apiClient from "@/api/client";
import { ElMessage } from "element-plus";
import { useSystemResources } from "@/composables/useSystemResources";
import { useResourceHistory } from "@/composables/useResourceHistory";
import ResourceGraphsView from "@/components/charts/ResourceGraphsView.vue";
import { getApiErrorMessage } from "@/utils/apiError";

const activeTab = ref("stats");

const resources = useSystemResources();
const resourceHistory = useResourceHistory(60); // Keep 60 data points (2 minutes at 2s intervals)

const cpuColors = [
  { color: "#10b981", percentage: 40 },
  { color: "#e6a23c", percentage: 70 },
  { color: "#f56c6c", percentage: 90 },
];

// Track resource snapshots in history
resources.onSnapshot((snapshot) => {
  if (snapshot) {
    console.log("[ResourceMonitor] Received snapshot:", {
      cpu: snapshot.node?.cpu_percent,
      memory: snapshot.node?.memory_percent,
      timestamp: new Date().toISOString(),
    });
    resourceHistory.addDataPoint(snapshot);
  }
});

// Log errors
resources.onError((error) => {
  console.error("[ResourceMonitor] WebSocket error:", error);
});

function clampPercent(value) {
  const number = Number(value);
  if (!Number.isFinite(number)) return 0;
  return Math.min(Math.max(Math.round(number), 0), 100);
}

function formatMB(value) {
  const number = Number(value);
  if (!Number.isFinite(number) || number <= 0) return "—";
  if (number >= 1024) return `${(number / 1024).toFixed(1)} GB`;
  return `${Math.round(number)} MB`;
}

const snapshot = computed(() => resources.snapshot.value || {});
const node = computed(() => snapshot.value.node || {});
const runtime = computed(() => snapshot.value.runtime || {});
const servers = computed(() => snapshot.value.servers || {});
const cpuPercent = computed(() => clampPercent(node.value.cpu_percent));
const memoryPercent = computed(() => clampPercent(node.value.memory_percent));
const diskPercent = computed(() => clampPercent(node.value.disk_percent));

const loadAverage = computed(() => {
  const values = [node.value.load_1, node.value.load_5, node.value.load_15];
  if (values.some((value) => value === null || value === undefined)) return "—";
  return values.map((value) => Number(value).toFixed(2)).join(" / ");
});

const connectionTagType = computed(() => {
  if (resources.state.value === "open") return "success";
  if (resources.state.value === "error") return "danger";
  return "warning";
});

const connectionLabel = computed(() => {
  if (resources.state.value === "open") return "LIVE";
  return resources.state.value.toUpperCase();
});

const lastUpdatedLabel = computed(() => {
  if (!resources.lastUpdated.value) return "—";
  return new Date(resources.lastUpdated.value).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
});

async function fetchFallbackInfo() {
  try {
    const { data } = await apiClient.get("/system/info");
    resources.snapshot.value = data.data || null;
    resources.lastUpdated.value = new Date().toISOString();
  } catch (e) {
    ElMessage.error("Failed to load system info: " + getApiErrorMessage(e));
  }
}

onMounted(() => {
  fetchFallbackInfo();
  resources.connect();
});

onUnmounted(() => {
  resources.disconnect();
});
</script>

<style scoped>
.resource-page {
  animation: fadeIn 0.3s ease-out;
}

.resource-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.resource-header h2 {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 800;
  color: var(--color-text);
}

.resource-header p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.runtime-grid {
  margin-top: 20px;
}

.stat-card {
  padding: 20px 24px;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.stat-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0;
}

.stat-meta {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text-muted);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.2;
  margin-bottom: 8px;
}

.stat-value.text-sm {
  font-size: 16px;
}

.stat-sub {
  font-size: 13px;
  color: var(--color-text-muted);
  margin: -2px 0 12px;
}

@media (max-width: 1000px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

.resource-tabs {
  margin-top: 1.5rem;
}

.resource-tabs :deep(.el-tabs__header) {
  margin-bottom: 1.5rem;
  border-bottom: 1px solid rgba(75, 85, 99, 0.3);
}

.resource-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none;
}

.resource-tabs :deep(.el-tabs__item) {
  font-size: 0.875rem;
  font-weight: 600;
  color: #6b7280;
  padding: 0 1rem;
  height: 40px;
  line-height: 40px;
}

.resource-tabs :deep(.el-tabs__item:hover) {
  color: #9ca3af;
}

.resource-tabs :deep(.el-tabs__item.is-active) {
  color: #e5e7eb;
}

.resource-tabs :deep(.el-tabs__active-bar) {
  height: 2px;
  background-color: var(--color-primary, #10b981);
}
</style>
