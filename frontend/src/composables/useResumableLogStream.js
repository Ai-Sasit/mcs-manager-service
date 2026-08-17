import { useReliableWebSocket } from "./useReliableWebSocket";

function parseLogEnvelope(raw, envelope) {
  if (envelope && typeof envelope === "object") return envelope;
  return { type: "log", data: raw };
}

export function useResumableLogStream(urlFactory, outputTypes) {
  const acceptedTypes = new Set(outputTypes);
  const socket = useReliableWebSocket(
    ({ streamId, lastEventId }) => {
      const url = new URL(urlFactory());
      if (streamId) {
        url.searchParams.set("stream_id", streamId);
        url.searchParams.set("last_event_id", String(lastEventId));
      }
      return url.toString();
    },
    { resumable: true, parseMessage: parseLogEnvelope },
  );

  function onLine(cb) {
    return socket.onMessage((message) => {
      if (acceptedTypes.has(message?.type)) {
        cb(message.data || "", message);
      }
    });
  }

  return {
    connect: socket.connect,
    disconnect: socket.disconnect,
    resetReplay: socket.resetCursorAndReconnect,
    send: socket.send,
    onLine,
    onControl: socket.onControl,
    onError: socket.onError,
    onStateChange: socket.onStateChange,
    state: socket.state,
    isOpen: socket.isOpen,
    error: socket.error,
    streamId: socket.streamId,
    lastEventId: socket.lastEventId,
  };
}
