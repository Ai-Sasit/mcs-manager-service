import { WS_BASE_URL } from "@/constants";

export function useServerLogs(serverId) {
  let ws = null;
  const listeners = new Set();
  let shouldReconnect = false;
  let reconnectTimer = null;

  function connect() {
    disconnect();
    shouldReconnect = true;
    const token = localStorage.getItem("mc_token") || "";
    const url = `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/logs?token=${encodeURIComponent(token)}`;
    ws = new WebSocket(url);

    ws.onmessage = (event) => {
      listeners.forEach((cb) => cb(event.data));
    };

    ws.onerror = () => {};

    ws.onclose = () => {
      ws = null;
      if (shouldReconnect) {
        reconnectTimer = setTimeout(() => connect(), 3000);
      }
    };
  }

  function onLine(cb) {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function disconnect() {
    shouldReconnect = false;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (ws) {
      ws.close();
      ws = null;
    }
    // Do NOT clear listeners here — they must survive reconnects.
    // Callers (onUnmounted) are responsible for releasing refs.
  }

  return { connect, disconnect, onLine };
}
