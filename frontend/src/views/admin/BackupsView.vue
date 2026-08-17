<template>
  <div class="backups-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Backups</h2>
        <p class="page-subtitle">Manage server backups.</p>
      </div>
      <div class="header-actions">
        <el-button
          type="primary"
          @click="createBackup"
          :loading="creating"
          :disabled="!selectedServerId"
        >
          <template #icon><PhPlus /></template>
          Create Backup
        </el-button>
        <el-button
          @click="fetchBackups"
          :loading="loading"
          :disabled="!selectedServerId"
        >
          <template #icon><PhArrowsClockwise /></template>
          Refresh
        </el-button>
      </div>
    </div>

    <div class="select-row">
      <el-select
        v-model="selectedServerId"
        placeholder="Select Server"
        @change="onServerChange"
        style="width: 250px"
      >
        <el-option
          v-for="s in store.servers"
          :key="s.id"
          :label="s.name"
          :value="s.id"
        />
      </el-select>
    </div>

    <div v-if="creating" class="info-alert">
      <el-alert
        type="info"
        show-icon
        :closable="false"
        description="The backup task is running in the background. Check audit logs for status."
      />
    </div>

    <div v-if="loading" class="loading-state">
      <PhSpinner :size="32" class="spin" />
      <p>Loading backups...</p>
    </div>

    <div v-else-if="!selectedServerId" class="placeholder-card card">
      <div class="placeholder-icon"><PhFloppyDisk :size="48" /></div>
      <h3>Select a Server</h3>
      <p>Choose a server to view its backups.</p>
    </div>

    <div v-else-if="backups.length === 0" class="placeholder-card card">
      <div class="placeholder-icon"><PhFloppyDisk :size="48" /></div>
      <h3>No Backups</h3>
      <p>Create your first backup to protect your server data.</p>
    </div>

    <div v-else class="card" style="padding: 24px">
      <el-table :data="backups" style="width: 100%">
        <el-table-column prop="name" label="Backup Name" min-width="200" />
        <el-table-column label="Size" width="120">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column label="Created" width="200">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="100" align="center">
          <template #default="{ row }">
            <el-popconfirm
              title="Delete this backup?"
              @confirm="handleDelete(row.name)"
            >
              <template #reference>
                <el-button type="danger" size="small" plain>
                  <template #icon><PhTrash /></template>
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import {
  PhPlus,
  PhArrowsClockwise,
  PhTrash,
  PhSpinner,
  PhFloppyDisk,
} from "@phosphor-icons/vue";
import { ElMessage, ElNotification } from "element-plus";
import apiClient from "@/api/client";
import { useServersStore } from "@/stores/servers";
import { getApiErrorMessage } from "@/utils/apiError";

const store = useServersStore();
const backups = ref([]);
const loading = ref(false);
const creating = ref(false);
const selectedServerId = ref("");

async function fetchBackups() {
  if (!selectedServerId.value) return;
  loading.value = true;
  try {
    const { data } = await apiClient.get(
      `/servers/${encodeURIComponent(selectedServerId.value)}/backups`,
    );
    backups.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to load backups: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

function onServerChange() {
  backups.value = [];
  fetchBackups();
}

async function createBackup() {
  if (!selectedServerId.value) return;
  creating.value = true;
  try {
    await apiClient.post(
      `/servers/${encodeURIComponent(selectedServerId.value)}/backups`,
    );
    ElNotification.success(
      "Backup task initiated. It will run in the background.",
    );
    fetchBackups();
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  } finally {
    creating.value = false;
  }
}

async function handleDelete(name) {
  if (!selectedServerId.value) return;
  try {
    await apiClient.delete(
      `/servers/${encodeURIComponent(selectedServerId.value)}/backups/${encodeURIComponent(name)}`,
    );
    ElMessage.success("Backup deleted.");
    fetchBackups();
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  }
}

function formatSize(bytes) {
  if (!bytes) return "—";
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / 1048576).toFixed(1) + " MB";
}

function formatDate(iso) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString();
}

onMounted(() => {
  if (store.servers.length === 0) store.fetchServers();
});
</script>

<style scoped>
.backups-page {
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

.lines-box {
  max-height: 400px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 13px;
  background: var(--color-bg);
  padding: 16px;
  border-radius: 8px;
}

.log-row {
  margin-bottom: 4px;
}

.text-muted {
  color: var(--color-text-muted);
}

.info-alert {
  margin-bottom: 8px;
}
</style>
