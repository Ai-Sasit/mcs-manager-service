<template>
  <div class="server-card card float-hover" :class="{ 'live-card': server.status === 'running' }" @click="$router.push(`/server/${server.id}`)">
    <div class="card-top">
      <div class="card-title-row">
        <h3 class="server-name">{{ server.name }}</h3>
        <div v-if="server.status === 'running'" class="live-indicator pulse-live"></div>
        <span v-else class="status-dot" :class="server.status"></span>
      </div>
      <div class="server-meta">
        <el-tag size="small" :type="getStatusType(server.status)" effect="plain" class="status-tag">
          {{ server.status.toUpperCase() }}
        </el-tag>
        <div class="meta-pills">
          <span class="meta-pill">v{{ server.version }}</span>
          <span class="meta-pill">{{ server.edition === 'java' ? 'Java' : 'Bedrock' }}</span>
        </div>
      </div>
    </div>

    <div class="card-stats">
      <div class="stat-item">
        <el-icon><Monitor /></el-icon>
        <span>{{ server.port }}</span>
      </div>
      <div class="stat-sep"></div>
      <div class="stat-item" v-if="server.edition === 'java'">
        <el-icon><Cpu /></el-icon>
        <span>{{ Math.round(server.memory_mb / 1024 * 10) / 10 }}GB</span>
      </div>
    </div>

    <div class="card-actions" @click.stop>
      <div class="action-group">
        <el-tooltip :content="server.status === 'stopped' ? 'Start Server' : 'Stop Server'" placement="top">
          <el-button
            circle
            :type="server.status === 'stopped' ? 'primary' : 'danger'"
            :icon="server.status === 'stopped' ? VideoPlay : VideoPause"
            @click="$emit('start', server.id)"
          />
        </el-tooltip>
        <el-tooltip content="Terminal" placement="top">
          <el-button
            circle
            :icon="Cpu"
            @click="$router.push(`/server/${server.id}?tab=terminal`)"
          />
        </el-tooltip>
      </div>
      <el-button
        class="details-btn"
        round
        @click="$router.push(`/server/${server.id}`)"
      >
        Manage
        <el-icon class="el-icon--right"><ArrowRight /></el-icon>
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { VideoPlay, VideoPause, ArrowRight, Monitor, Cpu } from "@element-plus/icons-vue";

defineProps({
  server: {
    type: Object,
    required: true,
  },
});

defineEmits(["start", "stop"]);

function getStatusType(status) {
  if (status === 'running') return 'success';
  if (status === 'starting') return 'warning';
  return 'info';
}
</script>

<style scoped>
.server-card {
  cursor: pointer;
  padding: 24px;
  position: relative;
  overflow: hidden;
  background: var(--color-white);
}

.live-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  background: var(--color-success);
}

.card-top {
  margin-bottom: 16px;
}

.card-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.server-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
  letter-spacing: -0.01em;
}

.live-indicator {
  width: 10px;
  height: 10px;
  background: var(--color-success);
  border-radius: 50%;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-text-muted);
}

.status-dot.starting {
  background: var(--color-warning);
  animation: pulse-dot 1.5s infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.server-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-tag {
  font-weight: 700 !important;
  font-size: 10px !important;
  letter-spacing: 0.05em;
  padding: 0 8px !important;
  height: 20px !important;
  line-height: 20px !important;
}

.meta-pills {
  display: flex;
  gap: 6px;
}

.meta-pill {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: 2px 8px;
  border-radius: 100px;
}

.card-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding: 12px;
  background: var(--color-bg);
  border-radius: var(--radius);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.stat-item .el-icon {
  font-size: 14px;
  color: var(--color-primary);
}

.stat-sep {
  width: 1px;
  height: 16px;
  background: var(--color-border);
}

.card-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.action-group {
  display: flex;
  gap: 8px;
}

.details-btn {
  font-weight: 600 !important;
  padding: 0 16px !important;
}

.details-btn:hover {
  background: var(--color-primary) !important;
  color: white !important;
  border-color: var(--color-primary) !important;
}
</style>
