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
      <div class="placeholder-icon"><PhClipboardText :size="48" /></div>
      <h3>No Audit Logs</h3>
      <p>Activity records will appear here as actions are performed.</p>
    </div>

    <div v-else class="card" style="padding: 24px">
      <el-table :data="paginatedLogs" style="width: 100%">
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
            <div class="target-cell">
              <span>{{ row.target }}</span>
              <span
                v-if="row.details"
                class="audit-description"
                :title="row.details"
              >
                {{ row.details }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Time" width="200">
          <template #default="{ row }">
            {{ formatDate(row.timestamp) }}
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[25, 50, 100]"
          :total="filteredLogs.length"
          layout="total, sizes, prev, pager, next"
          background
          @size-change="handlePageSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from "vue";
import {
  PhArrowsClockwise,
  PhMagnifyingGlass,
  PhSpinner,
  PhClipboardText,
} from "@phosphor-icons/vue";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";
import { ElMessage } from "element-plus";

const logs = ref([]);
const loading = ref(false);
const search = ref("");
const currentPage = ref(1);
const pageSize = ref(25);

const filteredLogs = computed(() => {
  if (!search.value) return logs.value;
  const q = search.value.toLowerCase();
  return logs.value.filter(
    (l) =>
      l.user?.toLowerCase().includes(q) ||
      l.action?.toLowerCase().includes(q) ||
      l.target?.toLowerCase().includes(q) ||
      l.details?.toLowerCase().includes(q),
  );
});

const paginatedLogs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredLogs.value.slice(start, start + pageSize.value);
});

watch(search, () => {
  currentPage.value = 1;
});

watch(filteredLogs, (items) => {
  const lastPage = Math.max(1, Math.ceil(items.length / pageSize.value));
  if (currentPage.value > lastPage) currentPage.value = lastPage;
});

function handlePageSizeChange() {
  currentPage.value = 1;
}

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
    currentPage.value = 1;
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
  display: inline-flex;
  margin-bottom: 16px;
  color: var(--color-primary);
}

.target-cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
  color: var(--color-text);
}

.audit-description {
  display: -webkit-box;
  overflow: hidden;
  color: var(--color-text-muted);
  font-size: 12px;
  line-height: 1.35;
  text-overflow: ellipsis;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
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
