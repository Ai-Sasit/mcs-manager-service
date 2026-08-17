<template>
  <div class="admin-layout">
    <div
      class="nav-scrim"
      :class="{ visible: drawerOpen }"
      aria-hidden="true"
      @click="closeDrawer"
    />

    <aside
      id="primary-navigation"
      class="sidebar"
      :class="{ 'is-open': drawerOpen }"
      aria-label="Primary navigation"
    >
      <div class="sidebar-header">
        <router-link to="/" class="brand" @click="closeDrawer">
          <span class="logo-mark"><PhMonitor :size="20" weight="bold" /></span>
          <span class="logo-text">MC Manage</span>
        </router-link>
        <button
          class="app-icon-button close-nav"
          type="button"
          aria-label="Close navigation"
          @click="closeDrawer"
        >
          <PhX :size="20" />
        </button>
      </div>

      <nav class="sidebar-nav">
        <section v-for="group in navigation" :key="group.label" class="nav-group">
          <h2 class="nav-group-label">{{ group.label }}</h2>
          <router-link
            v-for="item in group.items"
            :key="item.route"
            :to="item.to"
            class="nav-item"
            :class="{ active: route.name === item.route }"
            @click="closeDrawer"
          >
            <component :is="item.icon" :size="19" />
            <span>{{ item.label }}</span>
          </router-link>
        </section>
      </nav>

      <div class="sidebar-footer">
        <button class="nav-item logout" type="button" @click="handleLogout">
          <PhSignOut :size="19" />
          <span>Sign out</span>
        </button>
      </div>
    </aside>

    <div class="main-area">
      <header class="topbar">
        <div class="topbar-start">
          <button
            class="app-icon-button menu-button"
            type="button"
            :aria-expanded="drawerOpen"
            aria-controls="primary-navigation"
            aria-label="Open navigation"
            @click="drawerOpen = true"
          >
            <PhList :size="21" />
          </button>
          <div class="breadcrumb" aria-label="Breadcrumb">
            <span class="crumb-home">MC Manage</span>
            <PhCaretRight :size="13" class="crumb-sep" />
            <span class="crumb-current">{{ pageTitle }}</span>
          </div>
        </div>

        <div class="topbar-actions">
          <button
            class="app-icon-button"
            type="button"
            :aria-label="isDark ? 'Switch to light theme' : 'Switch to dark theme'"
            :title="isDark ? 'Light theme' : 'Dark theme'"
            @click="toggleTheme"
          >
            <PhSun v-if="isDark" :size="19" />
            <PhMoon v-else :size="19" />
          </button>
          <div class="user-info" :title="auth.username || 'admin'">
            <span class="user-avatar"><PhUser :size="16" /></span>
            <span>{{ auth.username || "admin" }}</span>
          </div>
        </div>
      </header>

      <main class="main-content">
        <div class="page-title-bar">
          <h1>{{ pageTitle }}</h1>
        </div>
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  PhArrowsClockwise,
  PhCalendar,
  PhCaretRight,
  PhCopy,
  PhCpu,
  PhFileText,
  PhFolder,
  PhGauge,
  PhGear,
  PhLink,
  PhList,
  PhMonitor,
  PhMoon,
  PhPlug,
  PhPuzzlePiece,
  PhSignOut,
  PhSquaresFour,
  PhSun,
  PhUser,
  PhUserFocus,
  PhX,
} from "@phosphor-icons/vue";
import { useAuthStore } from "@/stores/auth";
import { useTheme } from "@/composables/useTheme";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const drawerOpen = ref(false);
const { isDark, toggleTheme } = useTheme();

const navigation = [
  {
    label: "Management",
    items: [
      { label: "Dashboard", route: "dashboard", to: "/", icon: PhSquaresFour },
      { label: "Servers", route: "servers", to: "/servers", icon: PhList },
      { label: "Players", route: "players", to: "/players", icon: PhUserFocus },
    ],
  },
  {
    label: "Operations",
    items: [
      { label: "Files", route: "files", to: "/files", icon: PhFolder },
      { label: "Console", route: "console", to: "/console", icon: PhCpu },
      { label: "Backups", route: "backups", to: "/backups", icon: PhArrowsClockwise },
      { label: "Schedules", route: "schedules", to: "/schedules", icon: PhCalendar },
    ],
  },
  {
    label: "Extensions",
    items: [
      { label: "Plugins", route: "plugins", to: "/plugins", icon: PhPlug },
      { label: "Mods", route: "mods", to: "/mods", icon: PhPuzzlePiece },
      { label: "Integrations", route: "integrations", to: "/integrations", icon: PhLink },
    ],
  },
  {
    label: "System",
    items: [
      { label: "Audit logs", route: "audit-logs", to: "/audit-logs", icon: PhCopy },
      { label: "Backend logs", route: "backend-logs", to: "/backend-logs", icon: PhFileText },
      { label: "Resources", route: "resources", to: "/resources", icon: PhCpu },
      { label: "Port finder", route: "port-monitor", to: "/port-monitor", icon: PhGauge },
      { label: "Users", route: "users", to: "/users", icon: PhUser },
      { label: "Settings", route: "settings", to: "/settings", icon: PhGear },
    ],
  },
];

const pageTitle = computed(() => {
  const names = {
    dashboard: "Dashboard",
    "server-detail": "Server detail",
    servers: "Servers",
    players: "Players",
    files: "Files",
    console: "Console",
    backups: "Backups",
    schedules: "Schedules",
    plugins: "Plugins",
    mods: "Mods",
    integrations: "Integrations",
    "audit-logs": "Audit logs",
    "backend-logs": "Backend logs",
    resources: "Resources",
    "port-monitor": "Port finder",
    users: "Users",
    settings: "Settings",
  };
  return names[route.name] || "MC Manage";
});

function closeDrawer() {
  drawerOpen.value = false;
}

function handleEscape(event) {
  if (event.key === "Escape") closeDrawer();
}

function handleLogout() {
  auth.logout();
  router.push("/login");
}

watch(() => route.fullPath, closeDrawer);
onMounted(() => document.addEventListener("keydown", handleEscape));
onUnmounted(() => document.removeEventListener("keydown", handleEscape));
</script>

<style scoped>
.admin-layout {
  min-height: 100vh;
  color: var(--color-text);
}

.sidebar {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 30;
  display: flex;
  width: 272px;
  min-width: 272px;
  flex-direction: column;
  background: var(--color-bg-sidebar);
  border-right: 1px solid var(--color-border);
  box-shadow: var(--shadow-sidebar);
}

.sidebar-header {
  display: flex;
  min-height: 56px;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--color-border-light);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--color-text);
}

.logo-mark,
.user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
}

.logo-mark {
  width: 28px;
  height: 28px;
  color: #ffffff;
  background: var(--color-primary);
  border-radius: var(--radius-sm);
}

.logo-text {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.sidebar-nav {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  padding: 20px 12px;
}

.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-group-label {
  margin: 0;
  padding: 0 12px 6px;
  color: var(--color-text-muted);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.nav-item {
  position: relative;
  display: flex;
  min-height: 40px;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 8px 12px;
  color: var(--color-text-secondary);
  font: inherit;
  font-size: 14px;
  font-weight: 500;
  text-align: left;
  text-decoration: none;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: var(--radius-sm);
  transition: background-color 120ms ease, color 120ms ease;
}

.nav-item:hover {
  color: var(--color-text);
  background: var(--color-layer-alt);
}

.nav-item.active {
  color: var(--color-text);
  background: var(--color-primary-bg);
  font-weight: 600;
}

.nav-item.active::before {
  position: absolute;
  left: -12px;
  width: 3px;
  height: 24px;
  content: "";
  background: var(--color-primary);
  border-radius: 0 2px 2px 0;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--color-border-light);
}

.logout:hover {
  color: var(--color-danger);
  background: var(--color-danger-bg);
}

.main-area {
  display: flex;
  min-width: 0;
  min-height: 100vh;
  flex-direction: column;
  margin-left: 272px;
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  min-height: 56px;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: color-mix(in srgb, var(--color-bg) 82%, transparent);
  border-bottom: 1px solid var(--color-border-light);
  backdrop-filter: blur(18px) saturate(120%);
}

.topbar-start,
.topbar-actions,
.breadcrumb,
.user-info {
  display: flex;
  align-items: center;
}

.topbar-start {
  min-width: 0;
  gap: 12px;
}

.topbar-actions {
  gap: 10px;
}

.breadcrumb {
  min-width: 0;
  gap: 7px;
  font-size: 13px;
}

.crumb-home,
.crumb-sep {
  flex: 0 0 auto;
  color: var(--color-text-muted);
}

.crumb-current {
  overflow: hidden;
  color: var(--color-text);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-icon-button {
  display: inline-flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  padding: 0;
  color: var(--color-text-secondary);
  cursor: pointer;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
}

.app-icon-button:hover {
  color: var(--color-text);
  background: var(--color-layer-alt);
  border-color: var(--color-border-light);
}

.menu-button,
.close-nav {
  display: none;
}

.user-info {
  gap: 8px;
  max-width: 180px;
  overflow: hidden;
  color: var(--color-text-secondary);
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-avatar {
  width: 28px;
  height: 28px;
  color: var(--color-primary);
  background: var(--color-primary-bg);
  border: 1px solid color-mix(in srgb, var(--color-primary) 22%, var(--color-border));
  border-radius: 50%;
}

.main-content {
  flex: 1;
  padding: 0 28px 32px;
}

.page-title-bar {
  padding: 24px 0 20px;
}

.page-title-bar h1 {
  margin: 0;
  color: var(--color-text);
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.02em;
}

.nav-scrim {
  display: none;
}

@media (max-width: 960px) {
  .sidebar {
    box-shadow: var(--shadow-flyout);
    transform: translateX(-100%);
    transition: transform 180ms ease;
  }

  .sidebar.is-open {
    transform: translateX(0);
  }

  .main-area {
    margin-left: 0;
  }

  .menu-button,
  .close-nav {
    display: inline-flex;
  }

  .nav-scrim {
    position: fixed;
    inset: 0;
    z-index: 25;
    display: block;
    pointer-events: none;
    background: rgba(11, 18, 15, 0.36);
    opacity: 0;
    transition: opacity 180ms ease;
  }

  .nav-scrim.visible {
    pointer-events: auto;
    opacity: 1;
  }
}

@media (max-width: 640px) {
  .topbar {
    padding: 0 16px;
  }

  .main-content {
    padding: 0 16px 24px;
  }

  .page-title-bar {
    padding: 18px 0 16px;
  }

  .page-title-bar h1 {
    font-size: 24px;
  }

  .crumb-home,
  .crumb-sep {
    display: none;
  }

  .user-info > span:last-child {
    display: none;
  }
}
</style>
