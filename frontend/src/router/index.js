import { createRouter, createWebHistory } from "vue-router";

const routes = [
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/auth/LoginView.vue"),
    meta: { layout: "blank", public: true },
  },
  { path: "/", name: "dashboard", component: () => import("@/views/admin/DashboardView.vue") },
  { path: "/server/:id", name: "server-detail", component: () => import("@/views/admin/ServerDetailView.vue") },
  { path: "/servers", name: "servers", component: () => import("@/views/admin/ServersView.vue") },
  { path: "/players", name: "players", component: () => import("@/views/admin/PlayersView.vue") },
  { path: "/files", name: "files", component: () => import("@/views/admin/FilesView.vue") },
  { path: "/console", name: "console", component: () => import("@/views/admin/ConsoleView.vue") },
  { path: "/backups", name: "backups", component: () => import("@/views/admin/BackupsView.vue") },
  { path: "/schedules", name: "schedules", component: () => import("@/views/admin/SchedulesView.vue") },
  { path: "/plugins", name: "plugins", component: () => import("@/views/admin/PluginsView.vue") },
  {
    path: "/integrations",
    name: "integrations",
    component: () => import("@/views/admin/PlaceholderView.vue"),
    props: { title: "Integrations" },
  },
  { path: "/audit-logs", name: "audit-logs", component: () => import("@/views/admin/AuditLogsView.vue") },
  { path: "/users", name: "users", component: () => import("@/views/admin/UsersView.vue") },
  { path: "/backend-logs", name: "backend-logs", component: () => import("@/views/admin/BackendLogsView.vue") },
  { path: "/port-monitor", name: "port-monitor", component: () => import("@/views/admin/PortMonitorView.vue") },
  { path: "/settings", name: "settings", component: () => import("@/views/admin/SettingsView.vue") },
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/views/admin/PlaceholderView.vue"),
    props: { title: "Page Not Found" },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const token = localStorage.getItem("mc_token");
  if (!to.meta.public && !token) {
    return { name: "login" };
  }
});

export default router;
