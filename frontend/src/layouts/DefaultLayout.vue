<template>
  <div class="admin-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="logo-mark">
          <el-icon size="20" color="#ffffff"><Monitor /></el-icon>
        </div>
        <span class="logo-text">MC Manage</span>
      </div>

      <nav class="sidebar-nav">
        <div class="nav-group">
          <div class="nav-group-label">Management</div>
          <router-link
            to="/"
            class="nav-item"
            :class="{ active: route.name === 'dashboard' }">
            <el-icon size="18"><Grid /></el-icon>
            <span>Dashboard</span>
          </router-link>
          <router-link
            to="/servers"
            class="nav-item"
            :class="{ active: route.name === 'servers' }">
            <el-icon size="18"><List /></el-icon>
            <span>Servers</span>
          </router-link>
          <router-link
            to="/players"
            class="nav-item"
            :class="{ active: route.name === 'players' }">
            <el-icon size="18"><UserFilled /></el-icon>
            <span>Players</span>
          </router-link>
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Operations</div>
          <router-link
            to="/files"
            class="nav-item"
            :class="{ active: route.name === 'files' }">
            <el-icon size="18"><Folder /></el-icon>
            <span>Files</span>
          </router-link>
          <router-link
            to="/console"
            class="nav-item"
            :class="{ active: route.name === 'console' }">
            <el-icon size="18"><Cpu /></el-icon>
            <span>Console</span>
          </router-link>
          <router-link
            to="/backups"
            class="nav-item"
            :class="{ active: route.name === 'backups' }">
            <el-icon size="18"><Refresh /></el-icon>
            <span>Backups</span>
          </router-link>
          <router-link
            to="/schedules"
            class="nav-item"
            :class="{ active: route.name === 'schedules' }">
            <el-icon size="18"><Calendar /></el-icon>
            <span>Schedules</span>
          </router-link>
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Extensions</div>
          <router-link
            to="/plugins"
            class="nav-item"
            :class="{ active: route.name === 'plugins' }">
            <el-icon size="18"><Connection /></el-icon>
            <span>Plugins</span>
          </router-link>
          <router-link
            to="/integrations"
            class="nav-item"
            :class="{ active: route.name === 'integrations' }">
            <el-icon size="18"><Link /></el-icon>
            <span>Integrations</span>
          </router-link>
        </div>

        <div class="nav-group">
          <div class="nav-group-label">System</div>
          <router-link
            to="/audit-logs"
            class="nav-item"
            :class="{ active: route.name === 'audit-logs' }">
            <el-icon size="18"><DocumentCopy /></el-icon>
            <span>Audit Logs</span>
          </router-link>
          <router-link
            to="/backend-logs"
            class="nav-item"
            :class="{ active: route.name === 'backend-logs' }">
            <el-icon size="18"><Document /></el-icon>
            <span>Backend Logs</span>
          </router-link>
          <router-link
            to="/resources"
            class="nav-item"
            :class="{ active: route.name === 'resources' }">
            <el-icon size="18"><Cpu /></el-icon>
            <span>Resources</span>
          </router-link>
          <router-link
            to="/port-monitor"
            class="nav-item"
            :class="{ active: route.name === 'port-monitor' }">
            <el-icon size="18"><Odometer /></el-icon>
            <span>Port Monitor</span>
          </router-link>
          <router-link
            to="/users"
            class="nav-item"
            :class="{ active: route.name === 'users' }">
            <el-icon size="18"><User /></el-icon>
            <span>Users</span>
          </router-link>
          <router-link
            to="/settings"
            class="nav-item"
            :class="{ active: route.name === 'settings' }">
            <el-icon size="18"><Setting /></el-icon>
            <span>Settings</span>
          </router-link>
        </div>
      </nav>

      <div class="sidebar-footer">
        <button class="nav-item logout" @click="handleLogout">
          <el-icon size="18"><SwitchButton /></el-icon>
          <span>Logout</span>
        </button>
      </div>
    </aside>

    <!-- Main Area -->
    <div class="main-area">
      <!-- Top Bar -->
      <header class="topbar">
        <div class="breadcrumb">
          <span class="crumb-home">Home</span>
          <span class="crumb-sep">/</span>
          <span class="crumb-current">{{ pageTitle }}</span>
        </div>
        <div class="topbar-actions">
          <div class="user-info">
            <el-icon size="16"><User /></el-icon>
            <span>{{ auth.username || "admin" }}</span>
          </div>
        </div>
      </header>

      <!-- Page Title -->
      <div class="page-title-bar">
        <h1>{{ pageTitle }}</h1>
      </div>

      <!-- Content -->
      <main class="main-content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  Grid,
  List,
  UserFilled,
  Folder,
  Cpu,
  Refresh,
  Calendar,
  Connection,
  Link,
  DocumentCopy,
  Document,
  Odometer,
  User,
  Setting,
  SwitchButton,
  Monitor,
} from "@element-plus/icons-vue";
import { useAuthStore } from "@/stores/auth";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const pageTitle = computed(() => {
  const names = {
    dashboard: "Dashboard",
    "server-detail": "Server Detail",
    servers: "Servers",
    players: "Players",
    files: "Files",
    console: "Console",
    backups: "Backups",
    schedules: "Schedules",
    plugins: "Plugins",
    integrations: "Integrations",
    "audit-logs": "Audit Logs",
    "backend-logs": "Backend Logs",
    resources: "Resources",
    "port-monitor": "Port Monitor",
    users: "Users",
    settings: "Settings",
  };
  return names[route.name] || "MC Manage";
});

function handleLogout() {
  auth.logout();
  router.push("/login");
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}

/* ─── Sidebar ─── */
.sidebar {
  width: 240px;
  min-width: 240px;
  background: var(--color-white);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
  box-shadow: var(--shadow-sidebar);
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 20px;
  border-bottom: 1px solid var(--color-border);
}

.logo-mark {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-text {
  font-size: 16px;
  font-weight: 700;
  color: var(--color-text);
  letter-spacing: 0;
}

.sidebar-nav {
  flex: 1;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-group-label {
  padding: 0 12px 8px;
  font-size: 11px;
  font-weight: 700;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius);
  color: var(--color-text-secondary);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  border: none;
  background: transparent;
  width: 100%;
  font-family: inherit;
  transition: all 0.15s ease;
}

.nav-item:hover {
  background: var(--color-bg);
  color: var(--color-text);
}

.nav-item.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
  font-weight: 600;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--color-border);
}

.nav-item.logout:hover {
  background: rgba(239, 68, 68, 0.06);
  color: var(--color-danger);
}

/* ─── Main Area ─── */
.main-area {
  flex: 1;
  margin-left: 240px;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.topbar {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  background: var(--color-white);
  border-bottom: 1px solid var(--color-border);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.crumb-home {
  color: var(--color-text-muted);
}

.crumb-sep {
  color: var(--color-text-muted);
}

.crumb-current {
  color: var(--color-text);
  font-weight: 500;
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.page-title-bar {
  padding: 24px 28px 0;
}

.page-title-bar h1 {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
}

.main-content {
  flex: 1;
  padding: 20px 28px 28px;
  animation: fadeIn 0.3s ease-out;
}
</style>
