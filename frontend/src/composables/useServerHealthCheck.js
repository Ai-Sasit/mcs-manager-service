import { ref, onUnmounted } from "vue";
import apiClient from "@/api/client";
import { useServersStore } from "@/stores/servers";

const DEFAULT_INTERVAL = 15000; // 15 seconds

export function useServerHealthCheck(intervalMs = DEFAULT_INTERVAL) {
  const store = useServersStore();
  const checking = ref(false);
  let timer = null;

  async function checkRunningServers() {
    if (checking.value) return;
    checking.value = true;

    const runningServers = store.servers.filter((s) => s.status === "running");
    if (runningServers.length === 0) {
      checking.value = false;
      return;
    }

    const checks = runningServers.map(async (server) => {
      try {
        await apiClient.get(
          `/system/port-lookup?port=${encodeURIComponent(server.port)}`,
        );
        // Port is alive — server still running
        return { id: server.id, alive: true };
      } catch {
        // Port lookup failed — process likely dead
        return { id: server.id, alive: false };
      }
    });

    const results = await Promise.all(checks);

    for (const result of results) {
      if (!result.alive) {
        store.updateServerStatus(result.id, "stopped");
      }
    }

    checking.value = false;
  }

  function start() {
    stop();
    checkRunningServers();
    timer = setInterval(checkRunningServers, intervalMs);
  }

  function stop() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  // Re-fetch full server list on tab focus to catch external changes
  function onVisibilityChange() {
    if (document.visibilityState === "visible") {
      store.fetchServers();
      checkRunningServers();
    }
  }

  document.addEventListener("visibilitychange", onVisibilityChange);

  onUnmounted(() => {
    stop();
    document.removeEventListener("visibilitychange", onVisibilityChange);
  });

  return { checking, start, stop, runNow: checkRunningServers };
}