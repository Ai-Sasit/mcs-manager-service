import { WS_BASE_URL } from "../constants";

export function useServerTerminal(serverId) {
  let ws = null;
  const listeners = new Set();
  let shouldReconnect = false;
  let reconnectTimer = null;

  function connect() {
    disconnect();
    shouldReconnect = true;
    const token = localStorage.getItem("mc_token") || "";
    const url = `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/terminal?token=${encodeURIComponent(token)}`;
    ws = new WebSocket(url);

    ws.onmessage = (event) => {
      listeners.forEach((cb) => cb(event.data));
    };

    ws.onerror = () => {
      // error handled by onclose
    };

    ws.onclose = () => {
      ws = null;
      if (shouldReconnect) {
        reconnectTimer = setTimeout(() => connect(), 3000);
      }
    };
  }

  function sendCommand(cmd) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(cmd);
    }
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
    listeners.clear();
  }

  return { connect, disconnect, sendCommand, onLine };
}
