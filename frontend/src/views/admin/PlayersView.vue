<template>
  <div class="players-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Players Management</h2>
        <p class="page-subtitle">Manage server whitelists, ops, and player lists.</p>
      </div>
      <div class="header-actions">
        <el-select v-model="selectedServerId" placeholder="Select Server" @change="fetchPlayers" style="width: 250px">
          <el-option
            v-for="server in servers"
            :key="server.id"
            :label="server.name"
            :value="server.id" />
        </el-select>
        <el-button :icon="Plus" type="primary" :disabled="!selectedServerId" @click="showAddDialog = true">
          Add Player
        </el-button>
        <el-button :icon="Refresh" @click="fetchPlayers" :loading="loading" :disabled="!selectedServerId">
          Refresh
        </el-button>
      </div>
    </div>

    <el-card v-if="selectedServerId" v-loading="loading">
      <el-table :data="players" stripe style="width: 100%">
        <el-table-column prop="name" label="Player Name" />
        <el-table-column prop="uuid" label="UUID" min-width="180">
          <template #default="{ row }">
            <code class="uuid-text">{{ row.uuid || 'N/A' }}</code>
          </template>
        </el-table-column>
        <el-table-column label="Role" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'success'" size="small">
              {{ row.role === 'admin' ? 'OP' : 'Member' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="220" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button
                v-if="row.role !== 'admin'"
                size="small"
                type="warning"
                @click="updateRole(row.name, 'add_op')"
              >Make OP</el-button>
              <el-button
                v-else
                size="small"
                type="info"
                @click="updateRole(row.name, 'remove_op')"
              >De-OP</el-button>

              <el-popconfirm title="Remove from whitelist?" @confirm="updateRole(row.name, 'remove_whitelist')">
                <template #reference>
                  <el-button size="small" type="danger">Remove</el-button>
                </template>
              </el-popconfirm>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div v-else class="center-placeholder card">
      <el-empty description="Please select a server to manage players" />
    </div>

    <!-- Add Player Dialog -->
    <el-dialog v-model="showAddDialog" title="Add Player to Whitelist" width="400px">
      <el-form :model="addForm" @submit.prevent="handleAddPlayer">
        <el-form-item label="Player Name">
          <el-input v-model="addForm.name" placeholder="Minecraft Username" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">Cancel</el-button>
        <el-button type="primary" @click="handleAddPlayer" :loading="actionLoading">Add</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { Plus, Refresh } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import api from "@/api";
import { useServersStore } from "@/stores/servers";

const store = useServersStore();
const selectedServerId = ref("");
const players = ref([]);
const loading = ref(false);
const actionLoading = ref(false);
const showAddDialog = ref(false);
const addForm = ref({ name: "" });

const servers = computed(() => store.servers);

async function fetchPlayers() {
  if (!selectedServerId.value) return;
  loading.value = true;
  try {
    const { data } = await api.getPlayers(selectedServerId.value);
    players.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to fetch players: " + (e.response?.data?.message || e.message));
  } finally {
    loading.value = false;
  }
}

async function updateRole(name, action) {
  loading.value = true;
  try {
    const server = servers.value.find(s => s.id === selectedServerId.value);
    if (server.status !== "running") {
      throw new Error("Server must be running to execute player commands");
    }
    await api.updatePlayer(selectedServerId.value, { name, action });
    ElMessage.success("Player updated successfully (commands queued)");
    // Since commands are background, wait a bit before refresh or just assume success
    setTimeout(fetchPlayers, 1000);
  } catch (e) {
    ElMessage.error(e.message || "Failed to update player");
  } finally {
    loading.value = false;
  }
}

async function handleAddPlayer() {
  if (!addForm.value.name) return;
  actionLoading.value = true;
  try {
    await updateRole(addForm.value.name, "add_whitelist");
    showAddDialog.value = false;
    addForm.value.name = "";
  } finally {
    actionLoading.value = false;
  }
}

onMounted(async () => {
  if (store.servers.length === 0) {
    await store.fetchServers();
  }
});
</script>

<style scoped>
.uuid-text {
  font-family: monospace;
  font-size: 11px;
  background: var(--color-bg);
  padding: 2px 4px;
  border-radius: 4px;
}
.center-placeholder {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
</style>
