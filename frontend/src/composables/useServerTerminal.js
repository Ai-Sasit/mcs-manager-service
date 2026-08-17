import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useResumableLogStream } from "./useResumableLogStream";

export function useServerTerminal(serverId) {
  const socket = useResumableLogStream(
    () =>
      `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/terminal?token=${encodeURIComponent(getToken() || "")}`,
    ["terminal_output", "log"],
  );

  function sendCommand(cmd) {
    return socket.send({ type: "command", data: cmd });
  }

  return {
    connect: socket.connect,
    disconnect: socket.disconnect,
    resetReplay: socket.resetReplay,
    sendCommand,
    onLine: socket.onLine,
    onControl: socket.onControl,
    onError: socket.onError,
    state: socket.state,
    isOpen: socket.isOpen,
    error: socket.error,
    onStateChange: socket.onStateChange,
  };
}
