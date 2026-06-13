<template>
  <div class="players-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Online Players</h2>
        <p class="page-subtitle">
          Monitor connected players across all servers.
        </p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreate = true">
          <template #icon><PhPlus /></template>
          Add Player
        </el-button>
        <el-button @click="fetchPlayers" :loading="loading">
          <template #icon><PhArrowsClockwise /></template>
          Refresh
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <PhSpinner :size="32" class="spin" />
      <p>Scanning fleet...</p>
    </div>

    <div v-else-if="players.length === 0" class="empty-state">
      <div class="empty-icon">👤</div>
      <h3>No Active Players</h3>
      <p>No players are currently online across your fleet.</p>
    </div>

    <div v-else class="card" style="padding: 24px">
      <el-table :data="players" style="width: 100%">
        <el-table-column prop="username" label="Player" min-width="150" />
        <el-table-column prop="server_name" label="Server" min-width="150" />
        <el-table-column label="Status" width="120">
          <template #default>
            <el-tag type="success" size="small" effect="light">Online</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="showCreate" title="Add Player" width="400px">
      <el-form :model="form" label-position="top">
        <el-form-item label="Username">
          <el-input v-model="form.username" placeholder="Player name" />
        </el-form-item>
        <el-form-item label="Server">
          <el-select v-model="form.server_id" style="width: 100%">
            <el-option
              v-for="s in store.servers"
              :key="s.id"
              :label="s.name"
              :value="s.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">Cancel</el-button>
        <el-button type="primary" @click="handleCreate">Add</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { PhPlus, PhArrowsClockwise, PhSpinner } from "@phosphor-icons/vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { useServersStore } from "@/stores/servers";
import { getApiErrorMessage } from "@/utils/apiError";

const store = useServersStore();
const players = ref([]);
const loading = ref(false);
const showCreate = ref(false);

const form = ref({ username: "", server_id: "" });

async function fetchPlayers() {
  loading.value = true;
  try {
    if (store.servers.length === 0) await store.fetchServers();

    const results = [];
    for (const server of store.servers) {
      try {
        const { data } = await apiClient.get(
          `/servers/${encodeURIComponent(server.id)}/players`,
        );
        const serverPlayers = (data.data || []).map((p) => ({
          ...p,
          server_name: p.server_name || server.name,
          server_id: server.id,
        }));
        results.push(...serverPlayers);
      } catch {
        // skip servers that don't respond
      }
    }
    players.value = results;
  } catch (e) {
    ElMessage.error("Failed to load players: " + getApiErrorMessage(e));
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  if (!form.value.server_id) {
    ElMessage.warning("Please select a server.");
    return;
  }
  try {
    await apiClient.post(
      `/servers/${encodeURIComponent(form.value.server_id)}/players`,
      {
        username: form.value.username,
      },
    );
    ElMessage.success("Player added.");
    showCreate.value = false;
    form.value = { username: "", server_id: "" };
    fetchPlayers();
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  }
}

onMounted(() => {
  if (store.servers.length === 0) store.fetchServers();
  fetchPlayers();
});
</script>

<style scoped>
.players-page {
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

.empty-state {
  text-align: center;
  padding: 64px 32px;
}

.empty-icon {
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
