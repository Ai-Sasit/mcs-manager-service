<template>
  <div class="servers-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Servers Management</h2>
        <p class="page-subtitle">Manage, monitor, and configure all Minecraft server instances.</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" :icon="Plus" @click="showCreate = true">
          Create Server
        </el-button>
        <el-button :icon="Refresh" @click="store.fetchServers" :loading="store.loading">
          Refresh
        </el-button>
      </div>
    </div>

    <el-card class="servers-card">
      <el-table :data="store.servers" v-loading="store.loading" style="width: 100%">
        <el-table-column prop="name" label="Server Name" min-width="150" />
        <el-table-column prop="edition" label="Edition" width="120">
          <template #default="{ row }">
            <el-tag :type="row.edition === 'java' ? 'primary' : 'success'" effect="light">
              {{ row.edition.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="Version" width="100" />
        <el-table-column prop="port" label="Port" width="100" />
        <el-table-column label="Status" width="120">
          <template #default="{ row }">
            <div class="status-indicator" :class="row.status">
              <span class="dot"></span>
              {{ formatStatus(row.status) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="260" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button
                v-if="row.status === 'stopped'"
                type="success"
                size="small"
                @click="handleStart(row.id)"
                :loading="actionLoading[row.id] === 'start'"
              >
                Start
              </el-button>
              <el-button
                v-else
                type="danger"
                size="small"
                @click="handleStop(row.id)"
                :loading="actionLoading[row.id] === 'stop'"
              >
                Stop
              </el-button>
              <el-button
                size="small"
                @click="$router.push(`/server/${row.id}`)"
              >
                Details
              </el-button>
              <el-popconfirm
                title="Are you sure to delete this server?"
                @confirm="handleDelete(row.id)"
              >
                <template #reference>
                  <el-button
                    type="danger"
                    plain
                    size="small"
                    :loading="actionLoading[row.id] === 'delete'"
                  >
                    Delete
                  </el-button>
                </template>
              </el-popconfirm>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <CreateServerModal
      v-if="showCreate"
      @close="showCreate = false"
      @created="handleCreated"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from "vue";
import { ElMessage } from "element-plus";
import { Plus, Refresh } from "@element-plus/icons-vue";
import { useServersStore } from "../stores/servers";
import CreateServerModal from "../components/CreateServerModal.vue";

const store = useServersStore();
const showCreate = ref(false);
const actionLoading = reactive({});

onMounted(() => {
  store.fetchServers();
});

function formatStatus(status) {
  if (!status) return "Unknown";
  return status.charAt(0).toUpperCase() + status.slice(1);
}

async function handleStart(id) {
  actionLoading[id] = "start";
  try {
    await store.startServer(id);
    ElMessage.success("Server started");
  } catch (e) {
    ElMessage.error("Failed to start: " + (e.response?.data?.message || e.message));
  } finally {
    actionLoading[id] = "";
  }
}

async function handleStop(id) {
  actionLoading[id] = "stop";
  try {
    await store.stopServer(id);
    ElMessage.success("Server stopped");
  } catch (e) {
    ElMessage.error("Failed to stop: " + (e.response?.data?.message || e.message));
  } finally {
    actionLoading[id] = "";
  }
}

async function handleDelete(id) {
  actionLoading[id] = "delete";
  try {
    await store.deleteServer(id);
    ElMessage.success("Server deleted");
  } catch (e) {
    ElMessage.error("Failed to delete: " + (e.response?.data?.message || e.message));
  } finally {
    actionLoading[id] = "";
  }
}

function handleCreated() {
  store.fetchServers();
  showCreate.value = false;
}
</script>

<style scoped>
.servers-page {
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

.servers-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
}

.status-indicator .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-text-muted);
}

.status-indicator.running .dot {
  background: var(--color-success);
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.4);
}

.status-indicator.stopped .dot {
  background: var(--color-danger);
}

.status-indicator.starting .dot,
.status-indicator.stopping .dot {
  background: var(--color-warning);
  animation: pulse 1s infinite alternate;
}

@keyframes pulse {
  from { opacity: 0.5; }
  to { opacity: 1; }
}
</style>
