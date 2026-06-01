<template>
  <div class="dashboard-page">
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title text-gradient">System Overview</h2>
        <p class="page-subtitle">Real-time status of your Minecraft network.</p>
      </div>
      <div class="header-actions">
        <el-button round :icon="Plus" type="primary" @click="showCreate = true"
          >New Server</el-button
        >
        <el-button
          round
          :icon="Refresh"
          @click="refreshAll"
          :loading="store.loading"
          >Refresh</el-button
        >
      </div>
    </div>

    <!-- Top Stats -->
    <div class="stats-grid">
      <div class="stat-card glass-card float-hover">
        <div class="stat-header">
          <span class="stat-title">Instances</span>
          <div class="stat-icon green">
            <el-icon size="20"><Grid /></el-icon>
          </div>
        </div>
        <div class="stat-main">
          <div class="stat-value">{{ store.servers.length }}</div>
          <div class="stat-badge">{{ uniqueEditions }} Editions</div>
        </div>
        <div class="stat-progress-bg">
          <div class="stat-progress-bar" :style="{ width: '100%' }"></div>
        </div>
      </div>
      <div class="stat-card glass-card float-hover">
        <div class="stat-header">
          <span class="stat-title">Active Now</span>
          <div class="stat-icon emerald">
            <el-icon size="20"><CircleCheck /></el-icon>
          </div>
        </div>
        <div class="stat-main">
          <div class="stat-value">{{ store.runningCount }}</div>
          <div class="stat-badge success">{{ activePercentage }}% Online</div>
        </div>
        <div class="stat-progress-bg">
          <div
            class="stat-progress-bar success"
            :style="{ width: activePercentage + '%' }"
          ></div>
        </div>
      </div>
      <div class="stat-card glass-card float-hover">
        <div class="stat-header">
          <span class="stat-title">Node Load</span>
          <div class="stat-icon amber">
            <el-icon size="20"><Cpu /></el-icon>
          </div>
        </div>
        <div class="stat-main">
          <div class="stat-value">{{ mockCpu }}%</div>
          <div class="stat-badge warning">Optimal</div>
        </div>
        <div class="stat-progress-bg">
          <div
            class="stat-progress-bar warning"
            :style="{ width: mockCpu + '%' }"
          ></div>
        </div>
      </div>
    </div>

    <!-- Main Layout -->
    <div class="dashboard-layout">
      <!-- Left: Server Grid -->
      <div class="layout-main">
        <div class="section-header">
          <h3 class="section-title">Server Fleet</h3>
          <div class="header-tools">
            <el-radio-group v-model="filterEdition" size="small">
              <el-radio-button label="all">All</el-radio-button>
              <el-radio-button label="java">Java</el-radio-button>
              <el-radio-button label="bedrock">Bedrock</el-radio-button>
            </el-radio-group>
          </div>
        </div>

        <div v-if="store.loading" class="center-state">
          <el-icon class="spin" size="32"><Loading /></el-icon>
          <p>Syncing fleet data...</p>
        </div>

        <div class="server-grid" v-else-if="filteredServers.length > 0">
          <ServerCard
            v-for="server in filteredServers"
            :key="server.id"
            :server="server"
            @start="handleStart"
            @stop="handleStop"
          />
        </div>

        <div v-else class="empty-state glass-card">
          <div class="empty-icon">🎮</div>
          <h3>No Fleet Members</h3>
          <p>Your fleet is currently empty. Deploy a new instance to begin.</p>
          <br />
          <el-button type="primary" round @click="showCreate = true"
            >Deploy Instance</el-button
          >
        </div>
      </div>

      <!-- Right: Intelligence Widgets -->
      <div class="layout-side">
        <!-- Activity Timeline -->
        <div class="widget-container glass-card">
          <div class="widget-header">
            <h4>Live Feed</h4>
            <el-button link @click="$router.push('/audit-logs')"
              >Timeline</el-button
            >
          </div>
          <div class="activity-timeline" v-loading="loadingLogs">
            <div v-for="log in recentLogs" :key="log.id" class="timeline-item">
              <div
                class="timeline-node"
                :class="getActionClass(log.action)"
              ></div>
              <div class="timeline-content">
                <div class="timeline-header">
                  <span class="actor">{{ log.user }}</span>
                  <span class="time">{{ formatTime(log.timestamp) }}</span>
                </div>
                <p class="action-desc">
                  {{ log.action.toLowerCase().replace(/_/g, " ") }}
                  <span class="target">{{ log.target }}</span>
                </p>
              </div>
            </div>
            <el-empty
              v-if="recentLogs.length === 0"
              :image-size="40"
              description="Quiet for now"
            />
          </div>
        </div>

        <!-- Resource Mastery -->
        <div class="widget-container glass-card resource-widget">
          <div class="widget-header">
            <h4>Resources</h4>
            <el-tag size="small" type="success" effect="dark" round
              >HEALTHY</el-tag
            >
          </div>
          <div class="resource-meters">
            <div class="meter-item">
              <div class="meter-labels">
                <span>Memory usage</span>
                <span>{{ ramPercentage }}%</span>
              </div>
              <el-progress
                :percentage="ramPercentage"
                :show-text="false"
                stroke-width="8"
                color="#10b981"
              />
            </div>
            <div class="meter-item">
              <div class="meter-labels">
                <span>CPU overhead</span>
                <span>{{ mockCpu }}%</span>
              </div>
              <el-progress
                :percentage="mockCpu"
                :show-text="false"
                stroke-width="8"
                :color="cpuColors"
              />
            </div>
          </div>
        </div>

        <!-- System Actions -->
        <div class="quick-grid">
          <div
            class="quick-box glass-card float-hover"
            @click="$router.push('/schedules')"
          >
            <el-icon><Calendar /></el-icon>
            <span>Tasks</span>
          </div>
          <div
            class="quick-box glass-card float-hover"
            @click="$router.push('/backups')"
          >
            <el-icon><Box /></el-icon>
            <span>Safety</span>
          </div>
        </div>
      </div>
    </div>

    <CreateServerModal
      v-if="showCreate"
      @close="showCreate = false"
      @created="handleCreated"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { ElMessage } from "element-plus";
import {
  Plus,
  Grid,
  CircleCheck,
  Refresh,
  Cpu,
  Loading,
  Calendar,
  Box,
} from "@element-plus/icons-vue";
import { useServersStore } from "@/stores/servers";
import apiClient from "@/api/client";
import ServerCard from "@/components/servers/ServerCard.vue";
import CreateServerModal from "@/components/servers/CreateServerModal.vue";

const store = useServersStore();
const showCreate = ref(false);
const logs = ref([]);
const loadingLogs = ref(false);
const filterEdition = ref("all");

const mockCpu = ref(Math.floor(Math.random() * 15) + 3);
const cpuColors = [
  { color: "#10b981", percentage: 20 },
  { color: "#e6a23c", percentage: 40 },
  { color: "#f56c6c", percentage: 80 },
];

const filteredServers = computed(() => {
  if (filterEdition.value === "all") return store.servers;
  return store.servers.filter((s) => s.edition === filterEdition.value);
});

const recentLogs = computed(() => logs.value.slice(0, 6));

const uniqueEditions = computed(() => {
  const eds = new Set(store.servers.map((s) => s.edition));
  return eds.size;
});

const activePercentage = computed(() => {
  if (store.servers.length === 0) return 0;
  return Math.round((store.runningCount / store.servers.length) * 100);
});

const ramPercentage = computed(() => {
  if (store.servers.length === 0) return 10;
  const totalUsed = store.servers.reduce(
    (acc, s) => (s.status === "running" ? acc + (s.memory_mb || 0) : acc),
    0,
  );
  const max = 16384;
  return Math.min(Math.round((totalUsed / max) * 100), 100);
});

async function fetchLogs() {
  loadingLogs.value = true;
  try {
    const res = await apiClient.get("/audit-logs");
    logs.value = res.data.data || [];
  } catch (e) {
    console.error("Feed error:", e);
  } finally {
    loadingLogs.value = false;
  }
}

async function refreshAll() {
  await store.fetchServers();
  await fetchLogs();
  mockCpu.value = Math.floor(Math.random() * 15) + 3;
}

async function handleStart(id) {
  try {
    await store.startServer(id);
    await fetchLogs();
  } catch (e) {
    ElMessage.error("Activation failed");
  }
}

async function handleStop(id) {
  try {
    await store.stopServer(id);
    await fetchLogs();
  } catch (e) {
    ElMessage.error("Deactivation failed");
  }
}

function handleCreated() {
  store.fetchServers();
  fetchLogs();
}

function formatTime(iso) {
  const d = new Date(iso);
  const now = new Date();
  const diff = (now - d) / 1000;
  if (diff < 60) return "Just now";
  if (diff < 3600) return Math.floor(diff / 60) + "m ago";
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function getActionClass(action) {
  if (action.includes("START") || action.includes("CREATE")) return "success";
  if (action.includes("STOP") || action.includes("DELETE")) return "danger";
  return "info";
}

onMounted(() => {
  store.fetchServers();
  fetchLogs();
});
</script>

<style scoped>
.dashboard-page {
  animation: fadeIn 0.4s cubic-bezier(0.23, 1, 0.32, 1);
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
  font-weight: 500;
}

/* ─── Stats Grid ─── */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 40px;
}

.stat-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.stat-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon.green {
  background: rgba(16, 185, 129, 0.15);
  color: var(--color-primary);
}
.stat-icon.emerald {
  background: rgba(34, 197, 94, 0.15);
  color: #16a34a;
}
.stat-icon.amber {
  background: rgba(245, 158, 11, 0.15);
  color: #d97706;
}

.stat-main {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 16px;
}

.stat-value {
  font-size: 36px;
  font-weight: 900;
  color: var(--color-text);
  line-height: 0.9;
}

.stat-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  background: var(--color-bg);
  border-radius: 6px;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}

.stat-badge.success {
  background: rgba(34, 197, 94, 0.15);
  color: #16a34a;
}
.stat-badge.warning {
  background: rgba(245, 158, 11, 0.15);
  color: #d97706;
}

.stat-progress-bg {
  height: 6px;
  background: var(--color-border-light);
  border-radius: 100px;
  overflow: hidden;
}

.stat-progress-bar {
  height: 100%;
  background: var(--color-primary);
  border-radius: 100px;
}

.stat-progress-bar.success {
  background: #16a34a;
}
.stat-progress-bar.warning {
  background: #d97706;
}

/* ─── Layout ─── */
.dashboard-layout {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 32px;
}

.layout-main .section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.layout-main .section-title {
  font-size: 18px;
  font-weight: 800;
  margin: 0;
}

.server-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
}

/* ─── Timeline ─── */
.layout-side {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.widget-container {
  padding: 24px;
}

.widget-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.widget-header h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
}

.activity-timeline {
  display: flex;
  flex-direction: column;
}

.timeline-item {
  display: flex;
  gap: 16px;
  padding-bottom: 20px;
  position: relative;
}

.timeline-item:not(:last-child)::after {
  content: "";
  position: absolute;
  left: 6px;
  top: 18px;
  bottom: 0px;
  width: 2px;
  background: var(--color-border-light);
}

.timeline-node {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 3px solid var(--color-white);
  box-shadow: 0 0 0 1px var(--color-border);
  margin-top: 4px;
  z-index: 1;
}

.timeline-node.success {
  background: var(--color-success);
}
.timeline-node.danger {
  background: var(--color-danger);
}
.timeline-node.info {
  background: var(--color-primary);
}

.timeline-content {
  flex: 1;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.actor {
  font-size: 13px;
  font-weight: 700;
}

.time {
  font-size: 11px;
  color: var(--color-text-muted);
}

.action-desc {
  font-size: 12px;
  margin: 0;
  color: var(--color-text-secondary);
}

.target {
  font-weight: 700;
  color: var(--color-text);
}

/* ─── Resources ─── */
.resource-meters {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.meter-item {
  display: flex;
  flex-direction: column;
}

.meter-labels {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--color-text-secondary);
}

/* ─── Quick Grid ─── */
.quick-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.quick-box {
  height: 90px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  border-radius: var(--radius-lg);
}

.quick-box .el-icon {
  font-size: 24px;
  color: var(--color-primary);
}

.quick-box span {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-text-secondary);
}

/* ─── States ─── */
.empty-state {
  text-align: center;
  padding: 64px 32px;
}
.empty-icon {
  font-size: 48px;
  margin-bottom: 20px;
}
.center-state {
  text-align: center;
  padding: 48px 0;
  color: var(--color-text-muted);
  font-weight: 500;
}
.spin {
  animation: rotate 1.5s linear infinite;
}
@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1150px) {
  .dashboard-layout {
    grid-template-columns: 1fr;
  }
  .layout-side {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }
  .quick-grid {
    grid-column: span 2;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .layout-side {
    grid-template-columns: 1fr;
  }
  .quick-grid {
    grid-column: span 1;
  }
}
</style>
