<template>
  <div class="audit-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Audit Logs</h2>
        <p class="page-subtitle">Track all system activity and changes.</p>
      </div>
      <div class="header-actions">
        <el-input
          v-model="search"
          placeholder="Search logs..."
          size="default"
          style="width: 250px"
          clearable
        >
          <template #prefix><PhMagnifyingGlass /></template>
        </el-input>
        <el-button @click="fetchLogs" :loading="loading">
          <template #icon><PhArrowsClockwise /></template>
          Refresh
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <PhSpinner :size="32" class="spin" />
      <p>Loading audit trail...</p>
    </div>

    <div v-else-if="filteredLogs.length === 0" class="placeholder-card card">
      <div class="placeholder-icon">📋</div>
      <h3>No Audit Logs</h3>
      <p>Activity records will appear here as actions are performed.</p>
    </div>

    <div v-else class="card" style="padding: 24px">
      <el-table :data="filteredLogs" style="width: 100%">
        <el-table-column label="User" width="120">
          <template #default="{ row }">
            {{ row.user }}
          </template>
        </el-table-column>
        <el-table-column label="Action" width="160">
          <template #default="{ row }">
            <el-tag
              size="small"
              :type="getActionType(row.action)"
              effect="light"
            >
              {{ row.action.toLowerCase().replace(/_/g, " ") }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Target" min-width="180">
          <template #default="{ row }">
            {{ row.target }}
          </template>
        </el-table-column>
        <el-table-column label="Time" width="200">
          <template #default="{ row }">
            {{ formatDate(row.timestamp) }}
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
  PhMagnifyingGlass,
  PhSpinner,
} from "@phosphor-icons/vue";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";
import { ElMessage } from "element-plus";

const logs = ref([]);
const loading = ref(false);
const search = ref("");

const filteredLogs = computed(() => {
  if (!search.value) return logs.value;
  const q = search.value.toLowerCase();
  return logs.value.filter(
    (l) =>
      l.user?.toLowerCase().includes(q) ||
      l.action?.toLowerCase().includes(q) ||
      l.target?.toLowerCase().includes(q),
  );
});

function getActionType(action) {
  if (action.includes("CREATE") || action.includes("START")) return "success";
  if (action.includes("DELETE") || action.includes("STOP")) return "danger";
  if (action.includes("UPDATE") || action.includes("RESTART")) return "warning";
  return "info";
}

function formatDate(iso) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString();
}

async function fetchLogs() {
  loading.value = true;
  try {
    const { data } = await apiClient.get("/audit-logs");
    logs.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to load audit logs: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

onMounted(fetchLogs);
</script>

<style scoped>
.audit-page {
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

.placeholder-card {
  text-align: center;
  padding: 64px 32px;
}

.placeholder-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.loading-state {
  text-align: center;
  padding: 48px 0;
}

.spin {
  animation: rotate 1.5s linear infinite;
}

@keyframes rotate {
  to {
    transform: rotate(360deg);
  }
}
</style>
