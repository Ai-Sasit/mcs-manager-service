import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useReliableWebSocket } from "./useReliableWebSocket";

function parseSetupMessage(raw) {
  try {
    return JSON.parse(raw);
  } catch {
    return {
      type: "setup",
      step: "unknown",
      status: "active",
      message: raw,
      percent: 0,
    };
  }
}

export function useServerSetupProgress(jobId) {
  let lastEventId = 0;
  const socket = useReliableWebSocket(
    () => {
      const params = new URLSearchParams({
        token: getToken() || "",
      });
      if (lastEventId > 0) params.set("last_event_id", String(lastEventId));
      return `${WS_BASE_URL}/ws/server-setup/${encodeURIComponent(jobId)}?${params.toString()}`;
    },
    { parseMessage: parseSetupMessage },
  );

  function onEvent(cb) {
    return socket.onMessage((event) => {
      if (event.id) lastEventId = Math.max(lastEventId, event.id);
      cb(event);
    });
  }

  return {
    connect: socket.connect,
    disconnect: socket.disconnect,
    onEvent,
    onError: socket.onError,
    state: socket.state,
    isOpen: socket.isOpen,
    error: socket.error,
    onStateChange: socket.onStateChange,
  };
}
