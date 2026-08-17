<template>
  <div class="servers-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Servers Management</h2>
        <p class="page-subtitle">
          Manage, monitor, and configure all Minecraft server instances.
        </p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreate = true">
          <template #icon><PhPlus /></template>
          Create Server
        </el-button>
        <el-button @click="store.fetchServers" :loading="store.loading">
          <template #icon><PhArrowsClockwise /></template>
          Refresh
        </el-button>
      </div>
    </div>

    <el-card class="servers-card">
      <el-table
        :data="store.servers"
        v-loading="store.loading"
        style="width: 100%"
      >
        <el-table-column label="Server" min-width="240">
          <template #default="{ row }">
            <div class="server-identity">
              <span class="server-avatar" :class="row.edition">
                <PhCoffee v-if="row.edition === 'java'" :size="19" />
                <PhCube v-else :size="19" />
              </span>
              <span class="server-name-cell">
                <strong>{{ row.name }}</strong>
                <small>{{ row.id }}</small>
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Edition" width="130">
          <template #default="{ row }">
            <span class="edition-label" :class="row.edition">
              {{ row.edition === "java" ? "Java Edition" : "Bedrock Edition" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="Version" width="120">
          <template #default="{ row }">v{{ row.version }}</template>
        </el-table-column>
        <el-table-column label="Port" width="110">
          <template #default="{ row }"><code class="port-value">:{{ row.port }}</code></template>
        </el-table-column>
        <el-table-column label="Status" width="120">
          <template #default="{ row }">
            <StatusBadge :status="row.status" />
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="286" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button
                v-if="row.status === 'stopped'"
                class="server-command"
                type="primary"
                size="small"
                @click="handleStart(row.id)"
                :loading="actionLoading[row.id] === 'start'"
              >
                <template #icon><PhPlay /></template>
                Start
              </el-button>
              <el-button
                v-else
                class="server-command"
                type="danger"
                size="small"
                @click="handleStop(row.id)"
                :loading="actionLoading[row.id] === 'stop'"
              >
                <template #icon><PhStop /></template>
                Stop
              </el-button>
              <el-button
                class="server-manage-button"
                size="small"
                @click="$router.push(`/server/${row.id}`)"
              >
                Manage
                <template #icon><PhArrowRight /></template>
              </el-button>
              <el-popconfirm
                title="Are you sure to delete this server?"
                @confirm="handleDelete(row.id)"
              >
                <template #reference>
                  <el-button
                    class="server-delete-button"
                    type="danger"
                    plain
                    size="small"
                    aria-label="Delete server"
                    title="Delete server"
                    :loading="actionLoading[row.id] === 'delete'"
                  >
                    <template #icon><PhTrash /></template>
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
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
import {
  PhArrowRight,
  PhArrowsClockwise,
  PhCoffee,
  PhCube,
  PhPlay,
  PhPlus,
  PhStop,
  PhTrash,
} from "@phosphor-icons/vue";
import { useServersStore } from "@/stores/servers";
import CreateServerModal from "@/components/servers/CreateServerModal.vue";
import StatusBadge from "@/components/common/StatusBadge.vue";
import { useServerHealthCheck } from "@/composables/useServerHealthCheck";
import { getApiErrorMessage } from "@/utils/apiError";

const store = useServersStore();
const showCreate = ref(false);
const actionLoading = reactive({});
const healthCheck = useServerHealthCheck();

onMounted(() => {
  store.fetchServers();
  healthCheck.start();
});

async function handleStart(id) {
  actionLoading[id] = "start";
  try {
    await store.startServer(id);
    ElMessage.success("Server started");
  } catch (e) {
    ElMessage.error("Failed to start: " + getApiErrorMessage(e));
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
    ElMessage.error("Failed to stop: " + getApiErrorMessage(e));
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
    ElMessage.error("Failed to delete: " + getApiErrorMessage(e));
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

:deep(.servers-card .el-table) {
  border: 0 !important;
  border-radius: 0;
}

:deep(.servers-card .el-table__inner-wrapper::before) {
  display: none;
}

.server-identity,
.row-actions {
  display: flex;
  align-items: center;
}

.server-identity {
  gap: 12px;
  min-width: 0;
}

.server-avatar {
  display: inline-flex;
  width: 36px;
  height: 36px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
}

.server-avatar.java {
  color: var(--color-warning);
  background: var(--color-warning-bg);
}

.server-avatar.bedrock {
  color: var(--color-info);
  background: var(--color-info-bg);
}

.server-name-cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.server-name-cell strong {
  overflow: hidden;
  color: var(--color-text);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-name-cell small {
  overflow: hidden;
  color: var(--color-text-muted);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.edition-label {
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

.edition-label.java {
  color: var(--color-warning);
}

.edition-label.bedrock {
  color: var(--color-info);
}

.port-value {
  padding: 3px 6px;
  color: var(--color-text-secondary);
  background: var(--color-layer-alt);
  border-radius: var(--radius-sm);
  font-family: "Cascadia Code", Consolas, monospace;
  font-size: 12px;
  font-weight: 600;
}

.row-actions {
  justify-content: flex-end;
  gap: 6px;
}

.server-command {
  min-width: 72px;
}

.server-manage-button {
  min-width: 82px;
}

.server-delete-button {
  width: 32px;
  min-height: 32px;
  padding: 7px !important;
  color: var(--color-danger) !important;
  background: var(--color-danger-bg) !important;
  border-color: color-mix(in srgb, var(--color-danger) 32%, transparent) !important;
}

.server-delete-button:hover,
.server-delete-button:focus-visible {
  color: #ffffff !important;
  background: var(--color-danger) !important;
  border-color: var(--color-danger) !important;
}
</style>
