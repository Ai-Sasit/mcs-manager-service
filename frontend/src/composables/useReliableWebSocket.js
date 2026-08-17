import { computed, ref } from "vue";
import { clearAuthSession } from "@/utils/authStorage";

const BASE_RECONNECT_DELAY = 1200;
const MAX_RECONNECT_DELAY = 30000;
const STABLE_CONNECTION_MS = 10000;

export function useReliableWebSocket(urlFactory, options = {}) {
  let ws = null;
  let reconnectTimer = null;
  let stableTimer = null;
  let reconnectAttempts = 0;
  let generation = 0;
  let shouldBeConnected = false;
  let lifecycleBound = false;

  const listeners = new Set();
  const errorListeners = new Set();
  const stateListeners = new Set();
  const controlListeners = new Set();

  const state = ref("idle");
  const error = ref("");
  const lastUrl = ref("");
  const streamId = ref("");
  const lastEventId = ref(0);
  const isOpen = computed(() => state.value === "open");

  function setState(nextState) {
    if (state.value === nextState) return;
    state.value = nextState;
    stateListeners.forEach((cb) => cb(nextState));
  }

  function emitError(message, details = {}) {
    error.value = message;
    errorListeners.forEach((cb) => cb(message, details));
  }

  function emitControl(message) {
    controlListeners.forEach((cb) => cb(message));
  }

  function getUrl() {
    const context = {
      streamId: streamId.value,
      lastEventId: lastEventId.value,
    };
    const url =
      typeof urlFactory === "function" ? urlFactory(context) : urlFactory;
    lastUrl.value = url;
    return url;
  }

  function describeUrl() {
    try {
      const parsed = new URL(lastUrl.value);
      return `${parsed.origin}${parsed.pathname}`;
    } catch {
      return lastUrl.value || "websocket endpoint";
    }
  }

  function reconnectDelay() {
    const exponent = Math.max(0, reconnectAttempts - 1);
    const delay = Math.min(
      BASE_RECONNECT_DELAY * Math.pow(2, exponent),
      MAX_RECONNECT_DELAY,
    );
    return delay + Math.floor(Math.random() * 900);
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  }

  function clearStableTimer() {
    if (stableTimer) {
      clearTimeout(stableTimer);
      stableTimer = null;
    }
  }

  function isOffline() {
    return typeof navigator !== "undefined" && navigator.onLine === false;
  }

  function bindLifecycle() {
    if (lifecycleBound || typeof window === "undefined") return;
    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);
    document.addEventListener("visibilitychange", handleVisibilityChange);
    lifecycleBound = true;
  }

  function unbindLifecycle() {
    if (!lifecycleBound || typeof window === "undefined") return;
    window.removeEventListener("online", handleOnline);
    window.removeEventListener("offline", handleOffline);
    document.removeEventListener("visibilitychange", handleVisibilityChange);
    lifecycleBound = false;
  }

  function handleOnline() {
    if (!shouldBeConnected || isOpen.value) return;
    clearReconnectTimer();
    openSocket();
  }

  function handleOffline() {
    if (!shouldBeConnected) return;
    clearReconnectTimer();
    clearStableTimer();
    generation += 1;
    const socket = ws;
    ws = null;
    if (socket) socket.close(1000, "Browser went offline");
    setState("offline");
  }

  function handleVisibilityChange() {
    if (
      document.hidden ||
      !shouldBeConnected ||
      isOpen.value ||
      state.value === "connecting"
    ) {
      return;
    }
    clearReconnectTimer();
    openSocket();
  }

  function handleControlEnvelope(message) {
    if (message.type === "stream_reset") {
      streamId.value = message.stream_id || "";
      lastEventId.value = 0;
      emitControl(message);
      return true;
    }
    if (message.type === "stream_gap") {
      if (message.stream_id) streamId.value = message.stream_id;
      emitControl(message);
      return true;
    }
    return false;
  }

  function handleIncoming(raw) {
    let envelope = null;
    let acceptedEventId = 0;
    try {
      envelope = JSON.parse(raw);
    } catch {
      // Legacy/plain-text messages are still supported.
    }

    if (options.resumable && envelope && typeof envelope === "object") {
      if (handleControlEnvelope(envelope)) return;

      const incomingStreamId = envelope.stream_id || "";
      const incomingEventId = Number(envelope.event_id || 0);
      if (incomingStreamId) {
        if (streamId.value && incomingStreamId !== streamId.value) {
          streamId.value = incomingStreamId;
          lastEventId.value = 0;
          emitControl({
            type: "stream_reset",
            stream_id: incomingStreamId,
            data: "The log stream changed; replaying retained output.",
          });
        } else if (!streamId.value) {
          streamId.value = incomingStreamId;
        }
      }
      if (
        incomingEventId > 0 &&
        incomingStreamId === streamId.value &&
        incomingEventId <= lastEventId.value
      ) {
        return;
      }
      acceptedEventId = incomingEventId;
    }

    const parser = options.parseMessage || ((value) => value);
    try {
      const message = parser(raw, envelope);
      listeners.forEach((cb) => cb(message));
      if (acceptedEventId > 0) lastEventId.value = acceptedEventId;
    } catch {
      emitError("Failed to parse WebSocket message.", {
        url: describeUrl(),
      });
    }
  }

  function scheduleReconnect(event) {
    if (!shouldBeConnected) return;
    if (event?.code === 4000 || event?.code === 4001) {
      shouldBeConnected = false;
      setState("closed");
      unbindLifecycle();
      return;
    }
    if (event?.code === 4401) {
      shouldBeConnected = false;
      setState("auth-error");
      emitError("Your session expired. Please sign in again.", {
        url: describeUrl(),
        closeCode: event.code,
      });
      clearAuthSession();
      unbindLifecycle();
      if (
        typeof window !== "undefined" &&
        window.location.pathname !== "/login"
      ) {
        window.location.href = "/login";
      }
      return;
    }
    if (isOffline()) {
      setState("offline");
      return;
    }

    reconnectAttempts += 1;
    setState("reconnecting");
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      openSocket();
    }, reconnectDelay());
  }

  function openSocket() {
    if (!shouldBeConnected) return;
    clearReconnectTimer();
    clearStableTimer();

    if (isOffline()) {
      setState("offline");
      return;
    }
    if (
      ws &&
      (ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    const socketGeneration = ++generation;
    const socket = new WebSocket(getUrl());
    ws = socket;
    setState(reconnectAttempts > 0 ? "reconnecting" : "connecting");
    error.value = "";

    socket.onopen = () => {
      if (socketGeneration !== generation || ws !== socket) return;
      setState("open");
      stableTimer = setTimeout(() => {
        if (socketGeneration === generation && ws === socket) {
          reconnectAttempts = 0;
        }
      }, STABLE_CONNECTION_MS);
    };

    socket.onmessage = (event) => {
      if (socketGeneration !== generation || ws !== socket) return;
      handleIncoming(event.data);
    };

    socket.onerror = () => {
      if (socketGeneration !== generation || ws !== socket) return;
      emitError(`WebSocket connection error for ${describeUrl()}.`, {
        url: describeUrl(),
      });
    };

    socket.onclose = (event) => {
      if (socketGeneration !== generation || ws !== socket) return;
      ws = null;
      clearStableTimer();
      scheduleReconnect(event);
    };
  }

  function connect() {
    shouldBeConnected = true;
    bindLifecycle();
    openSocket();
  }

  function disconnect() {
    shouldBeConnected = false;
    clearReconnectTimer();
    clearStableTimer();
    reconnectAttempts = 0;
    generation += 1;
    const socket = ws;
    ws = null;
    if (socket) socket.close(1000, "Client disconnected");
    unbindLifecycle();
    setState("closed");
  }

  function resetCursorAndReconnect() {
    streamId.value = "";
    lastEventId.value = 0;
    shouldBeConnected = true;
    bindLifecycle();
    clearReconnectTimer();
    clearStableTimer();
    reconnectAttempts = 0;
    generation += 1;
    const socket = ws;
    ws = null;
    if (socket) socket.close(1000, "Replay requested");
    openSocket();
  }

  function send(payload) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return false;
    ws.send(typeof payload === "string" ? payload : JSON.stringify(payload));
    return true;
  }

  function onMessage(cb) {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function onError(cb) {
    errorListeners.add(cb);
    return () => errorListeners.delete(cb);
  }

  function onStateChange(cb) {
    stateListeners.add(cb);
    return () => stateListeners.delete(cb);
  }

  function onControl(cb) {
    controlListeners.add(cb);
    return () => controlListeners.delete(cb);
  }

  return {
    state,
    error,
    lastUrl,
    streamId,
    lastEventId,
    isOpen,
    connect,
    disconnect,
    resetCursorAndReconnect,
    send,
    onMessage,
    onError,
    onStateChange,
    onControl,
  };
}
