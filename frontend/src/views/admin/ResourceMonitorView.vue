<template>
  <div class="resource-page">
    <div class="stats-grid">
      <div class="stat-card card">
        <div class="stat-header">
          <span class="stat-title">Operating System</span>
        </div>
        <div class="stat-value">
          {{ info.os || "—" }} / {{ info.arch || "—" }}
        </div>
      </div>
      <div class="stat-card card">
        <div class="stat-header"><span class="stat-title">CPU Cores</span></div>
        <div class="stat-value">{{ info.cpu_count || "—" }}</div>
      </div>
      <div class="stat-card card">
        <div class="stat-header">
          <span class="stat-title">Memory Usage</span>
        </div>
        <div class="stat-value">{{ info.mem_alloc_mb || "—" }} MB</div>
        <div class="stat-sub">System: {{ info.mem_sys_mb || "—" }} MB</div>
      </div>
    </div>
    <div class="stats-grid" style="margin-top: 20px">
      <div class="stat-card card">
        <div class="stat-header">
          <span class="stat-title">Go Version</span>
        </div>
        <div class="stat-value text-sm">{{ info.go_version || "—" }}</div>
      </div>
      <div class="stat-card card">
        <div class="stat-header">
          <span class="stat-title">Goroutines</span>
        </div>
        <div class="stat-value">{{ info.goroutines || "—" }}</div>
      </div>
      <div class="stat-card card">
        <div class="stat-header"><span class="stat-title">GC Cycles</span></div>
        <div class="stat-value">{{ info.mem_gc_cycles || "—" }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import apiClient from "@/api/client";
import { ElMessage } from "element-plus";
import { getApiErrorMessage } from "@/utils/apiError";

const info = ref({});

async function fetchInfo() {
  try {
    const { data } = await apiClient.get("/system/info");
    info.value = data.data || {};
  } catch (e) {
    info.value = {};
    ElMessage.error("Failed to load system info: " + getApiErrorMessage(e));
  }
}

onMounted(fetchInfo);
</script>

<style scoped>
.resource-page {
  animation: fadeIn 0.3s ease-out;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
.stat-card {
  padding: 20px 24px;
}
.stat-header {
  margin-bottom: 12px;
}
.stat-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0;
}
.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.2;
}
.stat-value.text-sm {
  font-size: 16px;
}
.stat-sub {
  font-size: 13px;
  color: var(--color-text-muted);
  margin-top: 4px;
}
</style>
