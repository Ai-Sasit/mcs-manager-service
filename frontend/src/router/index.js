import { createRouter, createWebHistory } from "vue-router";
import DashboardView from "../views/DashboardView.vue";
import ServerDetailView from "../views/ServerDetailView.vue";
import ServersView from "../views/ServersView.vue";
import ConsoleView from "../views/ConsoleView.vue";
import PlayersView from "../views/PlayersView.vue";
import FilesView from "../views/FilesView.vue";
import BackupsView from "../views/BackupsView.vue";
import SchedulesView from "../views/SchedulesView.vue";
import PluginsView from "../views/PluginsView.vue";
import SettingsView from "../views/SettingsView.vue";
import AuditLogsView from "../views/AuditLogsView.vue";
import UsersView from "../views/UsersView.vue";
import FirewallView from "../views/FirewallView.vue";
import BackendLogsView from "../views/BackendLogsView.vue";
import LoginView from "../views/LoginView.vue";
import PortMonitorView from "../views/PortMonitorView.vue";
import PlaceholderView from "../views/PlaceholderView.vue";

const routes = [
  {
    path: "/login",
    name: "login",
    component: LoginView,
    meta: { layout: "blank", public: true },
  },
  { path: "/", name: "dashboard", component: DashboardView },
  { path: "/server/:id", name: "server-detail", component: ServerDetailView },
  { path: "/servers", name: "servers", component: ServersView },
  { path: "/players", name: "players", component: PlayersView },
  { path: "/files", name: "files", component: FilesView },
  { path: "/console", name: "console", component: ConsoleView },
  { path: "/backups", name: "backups", component: BackupsView },
  { path: "/schedules", name: "schedules", component: SchedulesView },
  { path: "/plugins", name: "plugins", component: PluginsView },
  {
    path: "/integrations",
    name: "integrations",
    component: PlaceholderView,
    props: { title: "Integrations" },
  },
  { path: "/audit-logs", name: "audit-logs", component: AuditLogsView },
  { path: "/users", name: "users", component: UsersView },
  { path: "/firewall", name: "firewall", component: FirewallView },
  { path: "/backend-logs", name: "backend-logs", component: BackendLogsView },
  { path: "/port-monitor", name: "port-monitor", component: PortMonitorView },
  { path: "/settings", name: "settings", component: SettingsView },
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
