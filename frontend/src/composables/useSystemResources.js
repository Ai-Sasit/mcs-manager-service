import { ref } from "vue";
import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useReliableWebSocket } from "./useReliableWebSocket";

function parseResourceMessage(raw) {
  const message = JSON.parse(raw);
  if (message.type === "resource_snapshot") return message;
  if (message.type === "error") {
    throw new Error(message.data || "Resource stream error");
  }
  return message;
}

export function useSystemResources() {
  const snapshot = ref(null);
  const lastUpdated = ref("");
  const socket = useReliableWebSocket(
    () => `${WS_BASE_URL}/ws/system/resources?token=${encodeURIComponent(getToken() || "")}`,
    { parseMessage: parseResourceMessage },
  );

  function onSnapshot(cb) {
    return socket.onMessage((message) => {
      if (message?.type !== "resource_snapshot") return;
      snapshot.value = message.data || null;
      lastUpdated.value = message.ts || new Date().toISOString();
      cb?.(snapshot.value, message);
    });
  }

  onSnapshot();

  return {
    snapshot,
    lastUpdated,
    connect: socket.connect,
    disconnect: socket.disconnect,
    onSnapshot,
    onError: socket.onError,
    state: socket.state,
    isOpen: socket.isOpen,
    error: socket.error,
    onStateChange: socket.onStateChange,
  };
}
