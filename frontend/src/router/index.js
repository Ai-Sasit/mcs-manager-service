import { createRouter, createWebHistory } from "vue-router";
import { getToken, getUser } from "@/utils/authStorage";
import { ROLES, canAccessRole, getDefaultPath } from "@/utils/roles";

const routes = [
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/auth/LoginView.vue"),
    meta: { layout: "blank", public: true },
  },
  {
    path: "/",
    name: "dashboard",
    component: () => import("@/views/admin/DashboardView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/server/:id",
    name: "server-detail",
    component: () => import("@/views/admin/ServerDetailView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/servers",
    name: "servers",
    component: () => import("@/views/admin/ServersView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/players",
    name: "players",
    component: () => import("@/views/admin/PlayersView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/files",
    name: "files",
    component: () => import("@/views/admin/FilesView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/console",
    name: "console",
    component: () => import("@/views/admin/ConsoleView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/backups",
    name: "backups",
    component: () => import("@/views/admin/BackupsView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/schedules",
    name: "schedules",
    component: () => import("@/views/admin/SchedulesView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/plugins",
    name: "plugins",
    component: () => import("@/views/admin/PluginsView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/integrations",
    name: "integrations",
    component: () => import("@/views/admin/PlaceholderView.vue"),
    props: { title: "Integrations" },
    meta: { requiresAuth: true, roles: [ROLES.ADMIN, ROLES.USER] },
  },
  {
    path: "/audit-logs",
    name: "audit-logs",
    component: () => import("@/views/admin/AuditLogsView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN] },
  },
  {
    path: "/users",
    name: "users",
    component: () => import("@/views/admin/UsersView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN] },
  },
  {
    path: "/backend-logs",
    name: "backend-logs",
    component: () => import("@/views/admin/BackendLogsView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN] },
  },
  {
    path: "/resources",
    name: "resources",
    component: () => import("@/views/admin/ResourceMonitorView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN] },
  },
  {
    path: "/port-monitor",
    name: "port-monitor",
    component: () => import("@/views/admin/PortMonitorView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN] },
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("@/views/admin/SettingsView.vue"),
    meta: { requiresAuth: true, roles: [ROLES.ADMIN] },
  },
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/views/admin/PlaceholderView.vue"),
    props: { title: "Page Not Found" },
    meta: { layout: "blank" },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 };
  },
});

router.beforeEach((to) => {
  if (to.meta.requiresAuth) {
    const token = getToken();
    if (!token) {
      return { name: "login", query: { redirect: to.fullPath } };
    }

    const user = getUser();
    if (!canAccessRole(to.meta.roles, user?.role)) {
      const defaultPath = getDefaultPath(user?.role);
      return to.path === defaultPath ? { name: "login" } : defaultPath;
    }
  }
});

export default router;
