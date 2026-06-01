<template>
  <div class="schedules-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Automated Schedules</h2>
        <p class="page-subtitle">
          Configure recurring tasks like backups and restarts.
        </p>
      </div>
      <div class="header-actions">
        <el-button :icon="Plus" type="primary" @click="showCreateDialog = true">
          New Schedule
        </el-button>
        <el-button :icon="Refresh" @click="fetchSchedules" :loading="loading">
          Refresh
        </el-button>
      </div>
    </div>

    <el-card v-loading="loading">
      <el-table :data="schedules" stripe style="width: 100%">
        <el-table-column label="Status" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="(val) => handleToggle(row.id, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="Server" width="200">
          <template #default="{ row }">
            {{ getServerName(row.server_id) }}
          </template>
        </el-table-column>
        <el-table-column prop="task" label="Task" width="150">
          <template #default="{ row }">
            <el-tag :type="getTaskTag(row.task)">{{
              row.task.toUpperCase()
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cron" label="Schedule (Cron)" />
        <el-table-column label="Actions" width="120" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              title="Delete this schedule?"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button :icon="Delete" type="danger" size="small" plain />
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <el-empty
        v-if="schedules.length === 0 && !loading"
        description="No schedules configured"
      />
    </el-card>

    <!-- Create Schedule Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      title="Create Automated Task"
      width="500px"
    >
      <el-form :model="form" label-width="120px">
        <el-form-item label="Target Server">
          <el-select
            v-model="form.server_id"
            placeholder="Select Server"
            style="width: 100%"
          >
            <el-option
              v-for="s in servers"
              :key="s.id"
              :label="s.name"
              :value="s.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Task Type">
          <el-select
            v-model="form.task"
            placeholder="Select Task"
            style="width: 100%"
          >
            <el-option label="Backup" value="backup" />
            <el-option label="Restart" value="restart" />
            <el-option label="Run Command" value="command" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.task === 'command'" label="Command">
          <el-input v-model="form.command" placeholder="e.g. /say Hello" />
        </el-form-item>
        <el-form-item label="Schedule (Cron)">
          <el-input
            v-model="form.cron"
            placeholder="0 0 * * * (Every midnight)"
          />
          <div class="helper-text">Format: Minute Hour Day Month DayOfWeek</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">Cancel</el-button>
        <el-button type="primary" @click="handleCreate" :loading="actionLoading"
          >Create</el-button
        >
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { Plus, Refresh, Delete } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { useServersStore } from "@/stores/servers";

const store = useServersStore();
const schedules = ref([]);
const loading = ref(false);
const actionLoading = ref(false);
const showCreateDialog = ref(false);

const form = ref({
  server_id: "",
  task: "backup",
  cron: "0 0 * * *",
  command: "",
});

const servers = computed(() => store.servers);

async function fetchSchedules() {
  loading.value = true;
  try {
    const { data } = await apiClient.get("/schedules");
    schedules.value = data.data || [];
  } catch (e) {
    ElMessage.error("Failed to load schedules");
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  if (!form.value.server_id || !form.value.cron) {
    return ElMessage.warning("Please fill all fields");
  }
  actionLoading.value = true;
  try {
    await apiClient.post("/schedules", form.value);
    ElMessage.success("Schedule created");
    showCreateDialog.value = false;
    fetchSchedules();
  } catch (e) {
    ElMessage.error("Failed to create schedule");
  } finally {
    actionLoading.value = false;
  }
}

async function handleToggle(id, enabled) {
  try {
    await apiClient.put(`/schedules/${encodeURIComponent(id)}/toggle`, {
      enabled,
    });
  } catch (e) {
    ElMessage.error("Failed to update schedule status");
    fetchSchedules();
  }
}

async function handleDelete(id) {
  try {
    await apiClient.delete(`/schedules/${encodeURIComponent(id)}`);
    ElMessage.success("Schedule removed");
    fetchSchedules();
  } catch (e) {
    ElMessage.error("Failed to delete schedule");
  }
}

function getServerName(id) {
  const s = servers.value.find((s) => s.id === id);
  return s ? s.name : "System";
}

function getTaskTag(task) {
  if (task === "backup") return "success";
  if (task === "restart") return "danger";
  return "warning";
}

onMounted(async () => {
  if (store.servers.length === 0) {
    await store.fetchServers();
  }
  fetchSchedules();
});
</script>

<style scoped>
.helper-text {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
}
</style>
