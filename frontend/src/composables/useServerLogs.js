import { WS_BASE_URL } from "@/constants";

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_RECONNECT_DELAY = 2000;
const MAX_RECONNECT_DELAY = 30000;

export function useServerLogs(serverId) {
  let ws = null;
  const listeners = new Set();
  let shouldReconnect = false;
  let reconnectTimer = null;
  let reconnectAttempts = 0;

  function getReconnectDelay() {
    const delay = Math.min(BASE_RECONNECT_DELAY * Math.pow(2, reconnectAttempts), MAX_RECONNECT_DELAY);
    return delay + Math.random() * 1000; // jitter
  }

  function connect() {
    disconnect();
    shouldReconnect = true;
    const token = localStorage.getItem("mc_token") || "";
    const url = `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/logs?token=${encodeURIComponent(token)}`;
    ws = new WebSocket(url);

    ws.onmessage = (event) => {
      reconnectAttempts = 0; // reset on successful message
      listeners.forEach((cb) => cb(event.data));
    };

    ws.onerror = () => {};

    ws.onclose = (event) => {
      ws = null;
      // Stop reconnecting on auth failures (4001-4999)
      if (event.code >= 4000 && event.code <= 4999) {
        shouldReconnect = false;
        return;
      }
      if (shouldReconnect && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts++;
        reconnectTimer = setTimeout(() => connect(), getReconnectDelay());
      }
    };
  }

  function onLine(cb) {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function disconnect() {
    shouldReconnect = false;
    reconnectAttempts = 0;
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
