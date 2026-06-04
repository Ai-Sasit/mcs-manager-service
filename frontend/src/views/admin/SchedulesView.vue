<template>
  <div v-if="featureEnabled" class="schedules-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Schedules</h2>
        <p class="page-subtitle">Automate server lifecycle tasks.</p>
      </div>
      <div class="header-actions">
        <el-button @click="showCreateDialog = true" type="primary">
          <template #icon><PhPlus /></template>
          New Schedule
        </el-button>
        <el-button @click="fetchSchedules" :loading="loading">
          <template #icon><PhArrowsClockwise /></template>
          Refresh
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <PhSpinner :size="32" class="spin" />
      <p>Loading schedules...</p>
    </div>

    <div v-else-if="schedules.length === 0" class="placeholder-card card">
      <div class="placeholder-icon">📅</div>
      <h3>No Schedules</h3>
      <p>Create your first schedule to automate server tasks.</p>
    </div>

    <div v-else class="schedules-grid">
      <div v-for="s in schedules" :key="s.id" class="schedule-card card">
        <div class="schedule-header">
          <h4>{{ s.name }}</h4>
          <el-popconfirm
            title="Delete this schedule?"
            @confirm="handleDelete(s.id)"
          >
            <template #reference>
              <el-button type="danger" size="small" plain>
                <template #icon><PhTrash /></template>
              </el-button>
            </template>
          </el-popconfirm>
        </div>
        <div class="schedule-body">
          <div class="schedule-meta">
            <span class="meta-label">Cron</span>
            <code>{{ s.cron }}</code>
          </div>
          <div class="schedule-meta">
            <span class="meta-label">Action</span>
            <el-tag size="small">{{ s.action }}</el-tag>
          </div>
          <div class="schedule-meta">
            <span class="meta-label">Target</span>
            <span>{{ s.server_name || s.target_id }}</span>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showCreateDialog" title="New Schedule" width="480px">
      <el-form :model="form" label-position="top">
        <el-form-item label="Name">
          <el-input v-model="form.name" placeholder="e.g. Nightly Restart" />
        </el-form-item>
        <el-form-item label="Cron Expression">
          <el-input v-model="form.cron" placeholder="0 3 * * *" />
        </el-form-item>
        <el-form-item label="Action">
          <el-select v-model="form.action" style="width: 100%">
            <el-option label="Start" value="start" />
            <el-option label="Stop" value="stop" />
            <el-option label="Restart" value="restart" />
          </el-select>
        </el-form-item>
        <el-form-item label="Target Server">
          <el-select v-model="form.target_id" style="width: 100%">
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
        <el-button @click="showCreateDialog = false">Cancel</el-button>
        <el-button type="primary" @click="handleCreate">Create</el-button>
      </template>
    </el-dialog>
  </div>

  <div v-else class="placeholder-card card">
    <div class="placeholder-icon">📅</div>
    <h2>Task Scheduler</h2>
    <p>Feature is not available.</p>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import {
  PhPlus,
  PhArrowsClockwise,
  PhTrash,
  PhSpinner,
} from "@phosphor-icons/vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { useServersStore } from "@/stores/servers";
import { getApiErrorMessage } from "@/utils/apiError";

const store = useServersStore();
const schedules = ref([]);
const loading = ref(false);
const showCreateDialog = ref(false);
const featureEnabled = ref(true);

const form = ref({ name: "", cron: "", action: "start", target_id: "" });

async function fetchSchedules() {
  loading.value = true;
  try {
    const { data } = await apiClient.get("/schedules");
    schedules.value = data.data || [];
  } catch (e) {
    if (e.response?.status === 404) {
      featureEnabled.value = false;
    } else {
      ElMessage.error("Failed to load schedules: " + getApiErrorMessage(e));
    }
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  try {
    await apiClient.post("/schedules", form.value);
    ElMessage.success("Schedule created.");
    showCreateDialog.value = false;
    form.value = { name: "", cron: "", action: "start", target_id: "" };
    fetchSchedules();
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  }
}

async function handleDelete(id) {
  try {
    await apiClient.delete(`/schedules/${encodeURIComponent(id)}`);
    ElMessage.success("Schedule deleted.");
    fetchSchedules();
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  }
}

onMounted(async () => {
  if (store.servers.length === 0) await store.fetchServers();
  fetchSchedules();
});
</script>

<style scoped>
.schedules-page {
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

.schedules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.schedule-card {
  padding: 24px;
}

.schedule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.schedule-header h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
}

.schedule-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.schedule-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.meta-label {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--color-text-muted);
  width: 60px;
}

.schedule-meta code {
  background: var(--color-bg);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
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
</style>
