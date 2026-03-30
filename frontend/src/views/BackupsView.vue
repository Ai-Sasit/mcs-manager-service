<template>
  <div class="backups-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Backups & Snapshots</h2>
        <p class="page-subtitle">Create and manage server backups to prevent data loss.</p>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedServerId" placeholder="Select Server" @change="fetchBackups" style="width: 250px">
          <el-option
            v-for="server in servers"
            :key="server.id"
            :label="server.name"
            :value="server.id" />
        </el-select>
        <el-button :icon="Plus" type="primary" :disabled="!selectedServerId" @click="createBackup" :loading="creating">
          Create Backup
        </el-button>
        <el-button :icon="Refresh" @click="fetchBackups" :loading="loading" :disabled="!selectedServerId">
          Refresh
        </el-button>
      </div>
    </div>

    <el-card v-if="selectedServerId">
      <el-alert 
         title="Backup in progress" 
         type="info" 
         v-if="creating" 
         show-icon 
         description="The backup task is running in the background. Check audit logs for status." 
         style="margin-bottom: 20px" />

      <el-table :data="backups" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="Backup Name" />
        <el-table-column label="Date Created" width="200">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="File Size" width="120">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button-group>
              <el-button
                size="small"
                type="info"
                disabled
                title="Restore not implemented yet"
              >Restore</el-button>

              <el-popconfirm title="Delete this backup forever?" @confirm="deleteBackup(row.name)">
                <template #reference>
                  <el-button size="small" type="danger">Delete</el-button>
                </template>
              </el-popconfirm>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="backups.length === 0 && !loading" description="No backups found for this server" />
    </el-card>

    <div v-else class="center-placeholder card">
      <el-empty description="Please select a server to view backups" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { Plus, Refresh } from "@element-plus/icons-vue";
import { ElMessage, ElNotification } from "element-plus";
import api from "../api";
import { useServersStore } from "../stores/servers";

const store = useServersStore();
const selectedServerId = ref("");
const backups = ref([]);
const loading = ref(false);
const creating = ref(false);

const servers = computed(() => store.servers);

async function fetchBackups() {
  if (!selectedServerId.value) return;
  loading.value = true;
  try {
    const { data } = await api.listBackups(selectedServerId.value);
    backups.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to fetch backups: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
  }
}

async function createBackup() {
  creating.value = true;
  try {
    const { data } = await api.createBackup(selectedServerId.value);
    ElNotification({
      title: 'Success',
      message: data.message || 'Backup started',
      type: 'success',
    });
    // It's backgrounded in Go, so we can't immediately refresh
    setTimeout(fetchBackups, 3000);
  } catch (e) {
    ElMessage.error("Failed to start backup: " + (e.response?.data?.message || e.message));
  } finally {
    creating.value = false;
  }
}

async function deleteBackup(name) {
  loading.value = true;
  try {
    await api.deleteBackup(selectedServerId.value, name);
    ElMessage.success("Backup deleted");
    fetchBackups();
  } catch (e) {
    ElMessage.error("Failed to delete backup: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
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
.center-placeholder {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
</style>
