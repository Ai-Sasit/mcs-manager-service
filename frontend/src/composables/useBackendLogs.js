import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useReliableWebSocket } from "./useReliableWebSocket";

function parseBackendLogMessage(raw) {
  try {
    const message = JSON.parse(raw);
    if (message.type === "log") return message.data || "";
  } catch {
    return raw;
  }
  return raw;
}

export function useBackendLogs() {
  const socket = useReliableWebSocket(
    () => `${WS_BASE_URL}/ws/backend-logs?token=${encodeURIComponent(getToken() || "")}`,
    { parseMessage: parseBackendLogMessage },
  );

  return {
    connect: socket.connect,
    disconnect: socket.disconnect,
    onLine: socket.onMessage,
    onError: socket.onError,
    state: socket.state,
    isOpen: socket.isOpen,
    error: socket.error,
    onStateChange: socket.onStateChange,
  };
}
