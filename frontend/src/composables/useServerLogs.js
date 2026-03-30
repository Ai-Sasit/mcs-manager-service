import { WS_BASE_URL } from "../constants";

export function useServerLogs(serverId) {
  let ws = null;
  const listeners = new Set();

  function connect() {
    const token = localStorage.getItem("mc_token") || "";
    const url = `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/logs?token=${encodeURIComponent(token)}`;
    ws = new WebSocket(url);

    ws.onmessage = (event) => {
      listeners.forEach((cb) => cb(event.data));
    };

    ws.onclose = () => {
      ws = null;
    };
  }

  function onLine(cb) {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function disconnect() {
    if (ws) {
      ws.close();
      ws = null;
    }
    listeners.clear();
  }

  return { connect, disconnect, onLine };
}
