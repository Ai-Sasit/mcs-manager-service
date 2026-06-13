<template>
  <div class="port-monitor-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title text-gradient">Port & Process Monitor</h2>
        <p class="page-subtitle">
          Manage OS processes and network ports for your fleet.
        </p>
      </div>
      <div class="header-actions">
        <el-input
          v-model="searchPortInput"
          placeholder="Lookup Port (e.g. 25565)"
          class="lookup-input"
          round
          clearable
          @keyup.enter="handleSearch"
        >
          <template #prefix><PhMagnifyingGlass /></template>
          <template #append>
            <el-button @click="handleSearch" :loading="searchLoading"
              ><template #icon><PhMagnifyingGlass /></template>Lookup</el-button
            >
          </template>
        </el-input>
        <el-button round @click="refresh" :loading="loading"
          ><template #icon><PhArrowsClockwise /></template>Refresh</el-button
        >
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-grid">
      <div class="stat-card glass-card">
        <div class="stat-header">
          <span class="stat-title">Listening Ports</span>
          <div class="stat-icon primary">
            <PhPlug :size="20" />
          </div>
        </div>
        <div class="stat-main">
          <div class="stat-value">{{ listeningCount }}</div>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-header">
          <span class="stat-title">Active PIDs</span>
          <div class="stat-icon success">
            <PhCpu :size="20" />
          </div>
        </div>
        <div class="stat-main">
          <div class="stat-value">{{ activePidCount }}</div>
        </div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-header">
          <span class="stat-title">Network Type</span>
          <div class="stat-icon warning">
            <PhGauge :size="20" />
          </div>
        </div>
        <div class="stat-main">
          <div class="stat-value">TCP/UDP</div>
        </div>
      </div>
    </div>

    <!-- Quick Lookup Result -->
    <div
      v-if="searchResult || searchError"
      class="search-result-container anim-slide-down"
    >
      <el-alert
        v-if="searchError"
        :title="searchError"
        type="error"
        show-icon
        closable
        @close="searchError = ''"
        class="glass-card"
      />
      <div v-else-if="searchResult" class="found-process-card glass-card">
        <div class="result-header">
          <div class="result-title">
            <PhCompass :size="20" class="pulse-icon" />
            <span>Process Found on Port {{ searchResult.port }}</span>
          </div>
          <el-button link @click="searchResult = null"
            ><template #icon><PhX /></template
          ></el-button>
        </div>
        <div class="result-body">
          <div class="process-meta">
            <div class="meta-item">
              <span class="label">Process Name</span>
              <span class="value">{{ searchResult.name }}</span>
            </div>
            <div class="meta-item">
              <span class="label">PID</span>
              <code class="pid-badge primary">{{ searchResult.pid }}</code>
            </div>
            <div class="meta-item">
              <span class="label">Protocol</span>
              <span class="value">{{ searchResult.protocol }}</span>
            </div>
          </div>
          <el-button
            type="danger"
            round
            @click="handleKillPid(searchResult.pid)"
          >
            <template #icon><PhXCircle /></template>
            Terminate Process
          </el-button>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="monitor-container glass-card">
      <el-table :data="store.servers" style="width: 100%" v-loading="loading">
        <el-table-column label="Instance" min-width="200">
          <template #default="{ row }">
            <div class="instance-cell">
              <div class="instance-avatar" :class="row.edition">
                {{ row.name.charAt(0).toUpperCase() }}
              </div>
              <div class="instance-info">
                <span class="instance-name">{{ row.name }}</span>
                <span class="instance-id">{{ row.id.split("-")[0] }}...</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="Port" width="120">
          <template #default="{ row }">
            <el-tag
              effect="light"
              round
              size="small"
              :type="row.status === 'running' ? 'success' : 'info'"
            >
              {{ row.port }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="Process ID (PID)" width="150">
          <template #default="{ row }">
            <code v-if="row.pid" class="pid-badge">{{ row.pid }}</code>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>

        <el-table-column label="Status" width="120">
          <template #default="{ row }">
            <div class="status-indicator">
              <span class="dot" :class="row.status"></span>
              <span class="status-text">{{ row.status }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="Actions" width="150" align="right">
          <template #default="{ row }">
            <el-dropdown trigger="click">
              <el-button link
                ><template #icon><PhDotsThree /></template
              ></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="$router.push(`/server/${row.id}`)"
                    ><template #icon><PhEye /></template>View
                    Details</el-dropdown-item
                  >
                  <el-dropdown-item
                    v-if="row.status === 'running'"
                    divided
                    class="text-danger"
                    @click="handleKill(row)"
                  >
                    <template #icon><PhXCircle /></template>
                    Force Kill (SIGKILL)
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import {
  PhArrowsClockwise,
  PhPlug,
  PhCpu,
  PhGauge,
  PhDotsThree,
  PhEye,
  PhXCircle,
  PhMagnifyingGlass,
  PhCompass,
  PhX,
} from "@phosphor-icons/vue";
import { useServersStore } from "@/stores/servers";
import { useServerHealthCheck } from "@/composables/useServerHealthCheck";
import { ElMessageBox, ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";

const store = useServersStore();
const loading = ref(false);
const searchPortInput = ref("");
const healthCheck = useServerHealthCheck();
const searchLoading = ref(false);
const searchResult = ref(null);
const searchError = ref("");

const listeningCount = computed(() => store.runningCount);
const activePidCount = computed(
  () => store.servers.filter((s) => !!s.pid).length,
);

async function refresh() {
  loading.value = true;
  await store.fetchServers();
  loading.value = false;
}

async function handleSearch() {
  if (!searchPortInput.value) return;
  searchLoading.value = true;
  searchError.value = "";
  searchResult.value = null;
  try {
    const res = await apiClient.get(
      `/system/port-lookup?port=${searchPortInput.value}`,
    );
    searchResult.value = res.data.data;
  } catch (e) {
    searchError.value = getApiErrorMessage(e, "No process found on this port.");
  } finally {
    searchLoading.value = false;
  }
}

async function handleKillPid(pid) {
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to force kill process ${pid}? This action cannot be undone.`,
      "Force Kill System Process",
      {
        confirmButtonText: "Terminate",
        cancelButtonText: "Cancel",
        type: "warning",
        confirmButtonClass: "el-button--danger",
      },
    );

    loading.value = true;
    await apiClient.post("/system/kill-pid", { pid });
    ElMessage.success(`Process ${pid} killed.`);
    searchResult.value = null;
    await store.fetchServers();
  } catch (e) {
    if (e !== "cancel") {
      ElMessage.error(getApiErrorMessage(e, "Failed to kill process"));
    }
  } finally {
    loading.value = false;
  }
}

async function handleKill(server) {
  try {
    await ElMessageBox.confirm(
      `Force kill process ${server.pid} for "${server.name}"? This will terminate the server immediately without saving data.`,
      "Force Kill Warning",
      {
        confirmButtonText: "Terminate Process",
        cancelButtonText: "Cancel",
        type: "error",
        confirmButtonClass: "el-button--danger",
      },
    );

    loading.value = true;
    await apiClient.post(`/servers/${encodeURIComponent(server.id)}/kill`);
    ElMessage.success("Process terminated successfully.");
    await store.fetchServers();
  } catch (e) {
    if (e !== "cancel") {
      ElMessage.error(getApiErrorMessage(e, "Failed to kill process"));
    }
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  store.fetchServers();
  healthCheck.start();
});
</script>

<style scoped>
.port-monitor-page {
  animation: fadeIn 0.4s cubic-bezier(0.23, 1, 0.32, 1);
  padding: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.page-title {
  margin: 0 0 6px;
  font-size: 28px;
  font-weight: 800;
}

.page-subtitle {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 15px;
}

.lookup-input {
  width: 300px;
  margin-right: 12px;
}

:deep(.lookup-input .el-input-group__append) {
  background-color: var(--color-primary);
  color: white;
  border: none;
  font-weight: 700;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 32px;
}

.stat-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.stat-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0;
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon.primary {
  background: rgba(var(--color-primary-rgb), 0.1);
  color: var(--color-primary);
}
.stat-icon.success {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}
.stat-icon.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.stat-value {
  font-size: 32px;
  font-weight: 900;
}

/* Quick Lookup Result */
.search-result-container {
  margin-bottom: 32px;
}

.found-process-card {
  padding: 20px 24px;
  border-left: 4px solid var(--color-danger);
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.result-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 800;
  font-size: 16px;
  color: var(--color-text);
}

.pulse-icon {
  color: var(--color-danger);
  font-size: 20px;
  animation: pulse-ring 2s infinite;
}

.result-body {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.process-meta {
  display: flex;
  gap: 40px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.meta-item .label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--color-text-muted);
  letter-spacing: 0;
}

.meta-item .value {
  font-size: 16px;
  font-weight: 700;
  color: var(--color-text);
}

.pid-badge.primary {
  background: var(--color-primary-bg);
  color: var(--color-primary);
  font-size: 14px;
  padding: 4px 10px;
}

@keyframes pulse-ring {
  0% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.7;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.anim-slide-down {
  animation: slideDown 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Table Card */
.monitor-container {
  padding: 24px;
}

.instance-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.instance-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 14px;
}

.instance-avatar.java {
  background: #fff7ed;
  color: #ea580c;
}
.instance-avatar.bedrock {
  background: #eff6ff;
  color: #2563eb;
}

.instance-info {
  display: flex;
  flex-direction: column;
}

.instance-name {
  font-weight: 700;
  font-size: 14px;
}

.instance-id {
  font-size: 11px;
  color: var(--color-text-muted);
}

.pid-badge {
  background: #f1f5f9;
  color: #475569;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-weight: 600;
  font-size: 12px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #94a3b8;
}

.dot.running {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}
.dot.starting {
  background: #f59e0b;
  animation: pulse 1s infinite;
}

.status-text {
  font-size: 13px;
  text-transform: capitalize;
  font-weight: 500;
}

.text-danger {
  color: var(--el-color-danger) !important;
}

.text-muted {
  color: var(--color-text-muted);
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
  100% {
    opacity: 1;
  }
}
</style>
