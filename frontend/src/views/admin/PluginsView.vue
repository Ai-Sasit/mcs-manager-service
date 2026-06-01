<template>
  <div class="plugins-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Plugins & Addons</h2>
        <p class="page-subtitle">
          Extend your server with custom features and modifications.
        </p>
      </div>
      <div class="header-actions">
        <el-select
          v-model="selectedServerId"
          placeholder="Select Server"
          @change="fetchPlugins"
          style="width: 250px"
        >
          <el-option
            v-for="server in servers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-upload
          v-if="selectedServerId"
          :action="uploadUrl"
          :headers="uploadHeaders"
          name="file"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :show-file-list="false"
          accept=".jar,.zip,.mcpack"
        >
          <el-button :icon="Upload" type="primary">Upload Plugin</el-button>
        </el-upload>
        <el-button
          :icon="Refresh"
          @click="fetchPlugins"
          :loading="loading"
          :disabled="!selectedServerId"
        >
          Refresh
        </el-button>
      </div>
    </div>

    <el-card v-if="selectedServerId">
      <el-table :data="plugins" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="Plugin Name">
          <template #default="{ row }">
            <div class="plugin-info">
              <el-icon><Connection /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Size" width="120">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column
          label="Actions"
          width="120"
          fixed="right"
          align="center"
        >
          <template #default="{ row }">
            <el-popconfirm
              title="Delete this plugin?"
              @confirm="deletePlugin(row.name)"
            >
              <template #reference>
                <el-button :icon="Delete" type="danger" size="small" plain />
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <el-empty
        v-if="plugins.length === 0 && !loading"
        description="No plugins found on this server"
      />
    </el-card>

    <div v-else class="center-placeholder card">
      <el-empty description="Please select a server to manage its plugins" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { Upload, Refresh, Delete, Connection } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { API_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useServersStore } from "@/stores/servers";

const store = useServersStore();
const selectedServerId = ref("");
const plugins = ref([]);
const loading = ref(false);

const servers = computed(() => store.servers);

const uploadUrl = computed(() => {
  return `${API_BASE_URL}/servers/${selectedServerId.value}/plugins`;
});

const uploadHeaders = computed(() => {
  const token = getToken();
  return {
    Authorization: `Bearer ${token}`,
  };
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
    ElMessage.error("Failed to fetch plugins");
  } finally {
    loading.value = false;
  }
}

function handleUploadSuccess() {
  ElMessage.success("Plugin uploaded and ready for next server start");
  fetchPlugins();
}

function handleUploadError() {
  ElMessage.error("Failed to upload plugin");
}

async function deletePlugin(name) {
  try {
    await apiClient.delete(
      `/servers/${encodeURIComponent(selectedServerId.value)}/plugins/${encodeURIComponent(name)}`,
    );
    ElMessage.success("Plugin deleted");
    fetchPlugins();
  } catch (e) {
    ElMessage.error("Failed to delete plugin");
  }
}

function formatSize(bytes) {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / 1048576).toFixed(1) + " MB";
}

onMounted(async () => {
  if (store.servers.length === 0) {
    await store.fetchServers();
  }
});
</script>

<style scoped>
.plugin-info {
  display: flex;
  align-items: center;
  gap: 12px;
}
.center-placeholder {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
</style>
