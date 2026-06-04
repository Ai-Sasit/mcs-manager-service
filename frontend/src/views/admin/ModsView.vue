<template>
  <div class="mods-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Mods</h2>
        <p class="page-subtitle">
          Manage mods for Forge & Fabric servers. Upload <code>.jar</code> files
          to the <code>mods/</code> directory.
        </p>
      </div>
      <div class="header-actions">
        <el-select
          v-model="selectedServerId"
          placeholder="Select a Forge/Fabric Server"
          @change="fetchMods"
          style="width: 280px"
        >
          <el-option
            v-for="server in moddedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          >
            <span class="server-option">
              <el-tag
                size="small"
                :type="server.server_type === 'forge' ? 'warning' : 'primary'"
              >
                {{ server.server_type.toUpperCase() }}
              </el-tag>
              {{ server.name }}
            </span>
          </el-option>
        </el-select>
        <el-upload
          v-if="selectedServerId"
          :action="uploadUrl"
          :headers="uploadHeaders"
          name="file"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :show-file-list="false"
          accept=".jar"
        >
          <el-button :icon="Upload" type="primary">Upload Mod</el-button>
        </el-upload>
        <el-button
          :icon="Refresh"
          @click="fetchMods"
          :loading="loading"
          :disabled="!selectedServerId"
        >
          Refresh
        </el-button>
      </div>
    </div>

    <el-card v-if="selectedServerId">
      <el-table :data="mods" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="Mod Name">
          <template #default="{ row }">
            <div class="mod-info">
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
              title="Delete this mod?"
              @confirm="deleteMod(row.name)"
            >
              <template #reference>
                <el-button :icon="Delete" type="danger" size="small" plain />
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <el-empty
        v-if="mods.length === 0 && !loading"
        description="No mods found on this server"
      />
    </el-card>

    <div v-else class="center-placeholder card">
      <el-empty
        description="Please select a Forge or Fabric server to manage its mods"
      />
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
import { getApiErrorMessage } from "@/utils/apiError";

const store = useServersStore();
const selectedServerId = ref("");
const mods = ref([]);
const loading = ref(false);

const servers = computed(() => store.servers);

const moddedServers = computed(() =>
  servers.value.filter(
    (s) => s.server_type === "forge" || s.server_type === "fabric",
  ),
);

const uploadUrl = computed(() => {
  return `${API_BASE_URL}/servers/${selectedServerId.value}/mods`;
});

const uploadHeaders = computed(() => {
  const token = getToken();
  return {
    Authorization: `Bearer ${token}`,
  };
});

async function fetchMods() {
  if (!selectedServerId.value) return;
  loading.value = true;
  try {
    const { data } = await apiClient.get(
      `/servers/${encodeURIComponent(selectedServerId.value)}/mods`,
    );
    mods.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to fetch mods: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

function handleUploadSuccess() {
  ElMessage.success("Mod uploaded and will be loaded on next server start");
  fetchMods();
}

function handleUploadError(error) {
  ElMessage.error("Failed to upload mod: " + getApiErrorMessage(error));
}

async function deleteMod(name) {
  try {
    await apiClient.delete(
      `/servers/${encodeURIComponent(selectedServerId.value)}/mods/${encodeURIComponent(name)}`,
    );
    ElMessage.success("Mod deleted");
    fetchMods();
  } catch (e) {
    ElMessage.error("Failed to delete mod: " + getApiErrorMessage(e));
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
.mod-info {
  display: flex;
  align-items: center;
  gap: 12px;
}
.server-option {
  display: flex;
  align-items: center;
  gap: 8px;
}
.center-placeholder {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
</style>
