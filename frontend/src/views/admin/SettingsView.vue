<template>
  <div class="settings-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">Settings</h2>
        <p class="page-subtitle">Application configuration.</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="saveSettings" :loading="saving">
          <template #icon><PhCheck /></template>
          Save
        </el-button>
      </div>
    </div>

    <div class="card" style="padding: 24px">
      <el-form :model="form" label-position="top" style="max-width: 600px">
        <el-form-item label="Base Server Directory">
          <el-input
            v-model="form.base_dir"
            placeholder="/opt/minecraft/servers"
          />
        </el-form-item>
        <el-form-item label="Java Binary Path">
          <el-input v-model="form.java_path" placeholder="java" />
        </el-form-item>
        <el-form-item label="Default Java Args">
          <el-input v-model="form.java_args" placeholder="-Xms512M -Xmx2G" />
        </el-form-item>
        <el-form-item label="Max Concurrent Instances">
          <el-input-number v-model="form.max_instances" :min="1" :max="50" />
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { PhCheck } from "@phosphor-icons/vue";
import { ElMessage } from "element-plus";
import apiClient from "@/api/client";
import { getApiErrorMessage } from "@/utils/apiError";

const saving = ref(false);

const form = ref({
  base_dir: "",
  java_path: "",
  java_args: "",
  max_instances: 10,
});

async function loadSettings() {
  try {
    const { data } = await apiClient.get("/system/settings");
    Object.assign(form.value, data.data || {});
  } catch (e) {
    ElMessage.error("Failed to load settings: " + getApiErrorMessage(e));
  }
}

async function saveSettings() {
  saving.value = true;
  try {
    await apiClient.put("/system/settings", form.value);
    ElMessage.success("Settings updated.");
  } catch (e) {
    ElMessage.error(getApiErrorMessage(e));
  } finally {
    saving.value = false;
  }
}

onMounted(loadSettings);
</script>

<style scoped>
.settings-page {
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
</style>
