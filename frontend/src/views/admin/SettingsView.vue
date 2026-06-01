<template>
  <div class="settings-view page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Global Settings</h2>
        <p class="page-subtitle">
          Configure application-wide parameters and integrations.
        </p>
      </div>
      <div class="header-actions">
        <el-button
          type="primary"
          :icon="Check"
          @click="saveSettings"
          :loading="saving"
        >
          Save Settings
        </el-button>
      </div>
    </div>
    <div class="settings-grid">
      <el-card class="settings-card card">
        <template #header
          ><div class="card-header">
            <span>General Configuration</span>
          </div></template
        >
        <el-form :model="form" label-position="top">
          <el-form-item label="Application Name">
            <el-input
              v-model="form.app_name"
              placeholder="MC Management Dashboard"
            />
          </el-form-item>
          <el-form-item label="Default Server RAM (MB)">
            <el-input-number
              v-model="form.default_ram"
              :min="512"
              :step="512"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item>
            <el-checkbox v-model="form.telemetry_enabled"
              >Enable System Telemetry</el-checkbox
            >
            <p class="helper-text">
              Collect anonymous data to improve the management platform.
            </p>
          </el-form-item>
        </el-form>
      </el-card>
      <el-card class="settings-card card">
        <template #header
          ><div class="card-header">
            <span>External Integrations</span>
          </div></template
        >
        <el-form :model="form" label-position="top">
          <el-form-item label="Discord Webhook URL">
            <el-input
              v-model="form.discord_webhook"
              placeholder="https://discord.com/api/webhooks/..."
              type="password"
              show-password
            />
            <p class="helper-text">
              Used for global notifications (backups, server status changes).
            </p>
          </el-form-item>
          <el-button type="info" plain disabled>Test Notification</el-button>
        </el-form>
      </el-card>
      <el-card class="settings-card card danger-zone">
        <template #header
          ><div class="card-header">
            <span style="color: var(--color-danger)">System Actions</span>
          </div></template
        >
        <div class="danger-actions">
          <div class="danger-item">
            <div class="item-info">
              <h4>Prune Audit Logs</h4>
              <p>Delete audit logs older than 30 days.</p>
            </div>
            <el-button type="danger" plain size="small" disabled
              >Prune</el-button
            >
          </div>
          <el-divider />
          <div class="danger-item">
            <div class="item-info">
              <h4>Factory Reset</h4>
              <p>Clear all server configurations and system settings.</p>
            </div>
            <el-button type="danger" size="small" disabled>Reset</el-button>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { Check } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";

const form = ref({
  app_name: "",
  default_ram: 2048,
  discord_webhook: "",
  telemetry_enabled: true,
});
const loading = ref(false);
const saving = ref(false);

async function fetchSettings() {
  loading.value = true;
  try {
    const { data } = await apiClient.get("/settings");
    form.value = data.data;
  } catch (e) {
    ElMessage.error("Failed to load settings");
  } finally {
    loading.value = false;
  }
}

async function saveSettings() {
  saving.value = true;
  try {
    await apiClient.put("/settings", form.value);
    ElMessage.success("Global settings updated successfully");
  } catch (e) {
    ElMessage.error("Failed to update settings");
  } finally {
    saving.value = false;
  }
}

onMounted(fetchSettings);
</script>

<style scoped>
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
}
.settings-card {
  height: min-content;
}
.helper-text {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
}
.danger-zone {
  border: 1px solid var(--color-danger-light);
}
.danger-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.item-info h4 {
  margin: 0 0 4px;
  font-size: 14px;
}
.item-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
