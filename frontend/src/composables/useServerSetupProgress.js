import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";

export function useServerSetupProgress(jobId) {
  let ws = null;
  const listeners = new Set();
  const errorListeners = new Set();

  function connect() {
    disconnect();
    const token = getToken() || "";
    const url = `${WS_BASE_URL}/ws/server-setup/${encodeURIComponent(jobId)}?token=${encodeURIComponent(token)}`;
    ws = new WebSocket(url);

    ws.onmessage = (event) => {
      try {
        listeners.forEach((cb) => cb(JSON.parse(event.data)));
      } catch {
        listeners.forEach((cb) =>
          cb({
            type: "setup",
            step: "unknown",
            status: "active",
            message: event.data,
            percent: 0,
          }),
        );
      }
    };

    ws.onerror = () => {
      errorListeners.forEach((cb) => cb("Setup progress connection failed."));
    };
  }

  function onEvent(cb) {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function onError(cb) {
    errorListeners.add(cb);
    return () => errorListeners.delete(cb);
  }

  function disconnect() {
    if (ws) {
      ws.close();
      ws = null;
    }
  }

  return { connect, disconnect, onEvent, onError };
}
