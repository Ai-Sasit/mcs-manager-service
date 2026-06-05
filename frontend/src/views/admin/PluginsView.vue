<template>
  <div class="plugins-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Plugins & Addons</h2>
        <p class="page-subtitle">Manage server plugins and addons.</p>
      </div>
      <div class="header-actions">
        <el-button @click="fetchPlugins" :loading="loading">
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

    <div v-if="loading" class="loading-state">
      <PhSpinner :size="32" class="spin" />
      <p>Loading plugins...</p>
    </div>

    <div v-else-if="!selectedServerId" class="placeholder-card card">
      <div class="placeholder-icon">🔌</div>
      <h3>Select a Server</h3>
      <p>Choose a server to manage its plugins.</p>
    </div>

    <template v-else>
      <div class="card" style="padding: 20px 24px">
        <PluginUpload
          :server-id="selectedServerId"
          :edition="selectedEdition"
          @uploaded="handleUploaded"
        />
      </div>

      <div v-if="plugins.length === 0" class="placeholder-card card">
        <div class="placeholder-icon">📦</div>
        <h3>No Plugins</h3>
        <p>No plugins installed for this server yet.</p>
      </div>

      <div v-else class="card" style="padding: 24px">
        <el-table :data="plugins" style="width: 100%">
          <el-table-column label="Name" min-width="180">
            <template #default="{ row }">
              <div class="plugin-info">
                <PhPlug :size="16" />
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="Size" width="120">
            <template #default="{ row }">
              {{ formatSize(row.size) }}
            </template>
          </el-table-column>
          <el-table-column label="Action" width="100" align="center">
            <template #default="{ row }">
              <el-popconfirm
                title="Delete this plugin?"
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
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import {
  PhArrowsClockwise,
  PhTrash,
  PhPlug,
  PhSpinner,
} from "@phosphor-icons/vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { useServersStore } from "@/stores/servers";
import { getApiErrorMessage } from "@/utils/apiError";
import PluginUpload from "@/components/servers/PluginUpload.vue";

const store = useServersStore();
const plugins = ref([]);
const loading = ref(false);
const selectedServerId = ref("");

const selectedEdition = computed(() => {
  const s = store.servers.find((s) => s.id === selectedServerId.value);
  return s?.edition || "java";
});

async function fetchPlugins() {
  if (!selectedServerId.value) return;
  loading.value = true;
  try {
    const { data } = await apiClient.get(
      `/servers/${encodeURIComponent(selectedServerId.value)}/plugins`,
    );
    plugins.value = data.data || [];
  } catch (e) {
    plugins.value = [];
    ElMessage.error("Failed to load plugins: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

function onServerChange() {
  plugins.value = [];
  fetchPlugins();
}

function handleUploaded() {
  fetchPlugins();
}

async function handleDelete(name) {
  try {
    await apiClient.delete(
      `/servers/${encodeURIComponent(selectedServerId.value)}/plugins/${encodeURIComponent(name)}`,
    );
    ElMessage.success("Plugin deleted.");
    fetchPlugins();
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

onMounted(() => {
  if (store.servers.length === 0) store.fetchServers();
});
</script>

<style scoped>
.plugins-page {
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

.plugin-info {
  display: flex;
  align-items: center;
  gap: 8px;
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

.select-row {
  display: flex;
  align-items: center;
}
</style>
