import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useReliableWebSocket } from "./useReliableWebSocket";

function parseLogMessage(raw) {
  try {
    const message = JSON.parse(raw);
    if (message.type === "log") return message.data || "";
  } catch {
    return raw;
  }
  return raw;
}

export function useServerLogs(serverId) {
  const socket = useReliableWebSocket(
    () =>
      `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/logs?token=${encodeURIComponent(getToken() || "")}`,
    { parseMessage: parseLogMessage },
  );

  function onLine(cb) {
    return socket.onMessage(cb);
  }

  return {
    connect: socket.connect,
    disconnect: socket.disconnect,
    onLine,
    state: socket.state,
    isOpen: socket.isOpen,
    error: socket.error,
    onStateChange: socket.onStateChange,
  };
}
