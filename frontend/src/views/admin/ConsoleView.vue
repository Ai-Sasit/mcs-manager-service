<template>
  <div class="console-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">System Console</h2>
        <p class="page-subtitle">Live overview of all server instances.</p>
      </div>
      <div class="header-actions">
        <el-button @click="store.fetchServers" :loading="store.loading">
          <template #icon><PhArrowsClockwise /></template>
          Refresh List
        </el-button>
      </div>
    </div>

    <div
      v-if="store.servers.length === 0 && !store.loading"
      class="center-state card"
    >
      <PhCpu :size="48" color="var(--color-text-muted)" />
      <h3>No Server Selected</h3>
      <p>Deploy a server to start managing it.</p>
    </div>

    <div v-else class="console-grid">
      <div
        v-for="server in store.servers"
        :key="server.id"
        class="console-card card"
        @click="$router.push(`/server/${server.id}`)"
      >
        <div class="console-header">
          <div class="server-identity">
            <div class="server-avatar" :class="server.edition">
              {{ server.name.charAt(0).toUpperCase() }}
            </div>
            <div class="server-details">
              <h4>{{ server.name }}</h4>
              <p>Port {{ server.port }} · {{ server.version }}</p>
            </div>
          </div>
          <div class="status-area">
            <span class="dot" :class="server.status"></span>
            <span class="status-label">{{ server.status }}</span>
          </div>
        </div>
        <div class="console-body">
          <div class="resource-meters">
            <div v-if="server.edition === 'java'" class="meter">
              <span>RAM</span>
              <el-progress
                :percentage="Math.round((server.memory_mb / 4096) * 100)"
                :show-text="false"
                stroke-width="6"
                color="#10b981"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { PhArrowsClockwise, PhCpu } from "@phosphor-icons/vue";
import { useServersStore } from "@/stores/servers";

const store = useServersStore();

onMounted(() => {
  store.fetchServers();
});
</script>

<style scoped>
.console-page {
  animation: fadeIn 0.3s ease-out;
  padding: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.page-title {
  margin: 0 0 6px;
  font-size: 28px;
  font-weight: 800;
}

.page-subtitle {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 15px;
}

.console-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 20px;
}

.console-card {
  padding: 20px;
  cursor: pointer;
}

.console-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.server-identity {
  display: flex;
  align-items: center;
  gap: 12px;
}

.server-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 16px;
}

.server-avatar.java {
  background: #fff7ed;
  color: #ea580c;
}

.server-avatar.bedrock {
  background: #eff6ff;
  color: #2563eb;
}

.server-details h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
}

.server-details p {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--color-text-muted);
}

.status-area {
  display: flex;
  align-items: center;
  gap: 6px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #94a3b8;
}

.dot.running {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}

.status-label {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.center-state {
  text-align: center;
  padding: 80px 32px;
  color: var(--color-text-muted);
}
</style>
