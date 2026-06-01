<template>
  <div class="audit-logs-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Audit Logs</h2>
        <p class="page-subtitle">
          Track all administrative actions and system events.
        </p>
      </div>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          placeholder="Search logs..."
          clearable
          :prefix-icon="Search"
          style="width: 250px"
        />
        <el-button :icon="Refresh" @click="fetchLogs" :loading="loading">
          Refresh
        </el-button>
      </div>
    </div>

    <el-card class="logs-card">
      <el-table
        :data="filteredLogs"
        v-loading="loading"
        style="width: 100%"
        stripe
        height="calc(100vh - 240px)"
      >
        <el-table-column prop="timestamp" label="Time" width="180">
          <template #default="{ row }">
            {{ new Date(row.timestamp).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="user" label="User" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.user }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="Action" width="160">
          <template #default="{ row }">
            <el-tag :type="getActionColor(row.action)" effect="light">
              {{ row.action.replace(/_/g, " ") }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target" label="Target / Resource" width="200" />
        <el-table-column
          prop="details"
          label="Details"
          min-width="250"
          show-overflow-tooltip
        />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { Refresh, Search } from "@element-plus/icons-vue";
import apiClient from "@/api/client";

const logs = ref([]);
const loading = ref(false);
const searchQuery = ref("");

const filteredLogs = computed(() => {
  if (!searchQuery.value) return logs.value;
  const q = searchQuery.value.toLowerCase();
  return logs.value.filter(
    (log) =>
      log.action.toLowerCase().includes(q) ||
      log.target.toLowerCase().includes(q) ||
      log.details.toLowerCase().includes(q) ||
      log.user.toLowerCase().includes(q),
  );
});

async function fetchLogs() {
  loading.value = true;
  try {
    const res = await apiClient.get("/audit-logs");
    logs.value = res.data.data || [];
  } catch (e) {
    ElMessage.error(
      "Failed to load audit logs: " + (e.response?.data?.message || e.message),
    );
  } finally {
    loading.value = false;
  }
}

function getActionColor(action) {
  if (action.includes("CREATE") || action.includes("START")) return "success";
  if (action.includes("DELETE") || action.includes("STOP")) return "danger";
  if (action.includes("RESTART") || action.includes("UPDATE")) return "warning";
  return "info";
}

onMounted(() => {
  fetchLogs();
});
</script>

<style scoped>
.audit-logs-page {
  animation: fadeIn 0.3s ease-out;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-title {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
}
.page-subtitle {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}
.header-actions {
  display: flex;
  gap: 12px;
}
.logs-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
:deep(.logs-card .el-card__body) {
  padding: 0;
}
</style>
