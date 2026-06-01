import { computed, ref } from "vue";

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_RECONNECT_DELAY = 1200;
const MAX_RECONNECT_DELAY = 30000;

export function useReliableWebSocket(urlFactory, options = {}) {
  let ws = null;
  let reconnectTimer = null;
  let reconnectAttempts = 0;
  let manuallyClosed = false;
  const listeners = new Set();
  const errorListeners = new Set();
  const stateListeners = new Set();

  const state = ref("idle");
  const error = ref("");
  const lastUrl = ref("");
  const isOpen = computed(() => state.value === "open");

  function setState(nextState) {
    state.value = nextState;
    stateListeners.forEach((cb) => cb(nextState));
  }

  function getUrl() {
    const url = typeof urlFactory === "function"
      ? urlFactory(options.context?.() || {})
      : urlFactory;
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
    const delay = Math.min(
      BASE_RECONNECT_DELAY * Math.pow(2, reconnectAttempts),
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

  function scheduleReconnect(event) {
    if (manuallyClosed) return;
    if (event?.code >= 4000 && event.code <= 4999) {
      setState("closed");
      error.value = "WebSocket authorization failed. Please sign in again.";
      errorListeners.forEach((cb) => cb(error.value, { url: describeUrl(), closeCode: event.code }));
      return;
    }
    if (reconnectAttempts >= (options.maxReconnectAttempts || MAX_RECONNECT_ATTEMPTS)) {
      setState("error");
      error.value = `WebSocket connection failed for ${describeUrl()}. Check production proxy /ws upgrade routing.`;
      errorListeners.forEach((cb) => cb(error.value, { url: describeUrl(), closeCode: event?.code }));
      return;
    }

    reconnectAttempts += 1;
    setState("reconnecting");
    const retry = () => {
      if (document.hidden) {
        reconnectTimer = setTimeout(retry, 2000);
        return;
      }
      connect();
    };
    reconnectTimer = setTimeout(retry, reconnectDelay());
  }

  function connect() {
    clearReconnectTimer();
    if (ws) {
      ws.onclose = null;
      ws.close();
      ws = null;
    }

    manuallyClosed = false;
    setState(reconnectAttempts > 0 ? "reconnecting" : "connecting");
    error.value = "";
    ws = new WebSocket(getUrl());

    ws.onopen = () => {
      reconnectAttempts = 0;
      setState("open");
    };

    ws.onmessage = (event) => {
      const parser = options.parseMessage || ((value) => value);
      try {
        const message = parser(event.data);
        listeners.forEach((cb) => cb(message));
      } catch {
        error.value = "Failed to parse WebSocket message.";
        errorListeners.forEach((cb) => cb(error.value));
      }
    };

    ws.onerror = () => {
      error.value = `WebSocket connection error for ${describeUrl()}.`;
      errorListeners.forEach((cb) => cb(error.value, { url: describeUrl() }));
    };

    ws.onclose = (event) => {
      ws = null;
      if (manuallyClosed) {
        setState("closed");
        return;
      }
      scheduleReconnect(event);
    };
  }

  function disconnect() {
    manuallyClosed = true;
    clearReconnectTimer();
    reconnectAttempts = 0;
    if (ws) {
      ws.close();
      ws = null;
    }
    setState("closed");
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

  return { state, error, lastUrl, isOpen, connect, disconnect, send, onMessage, onError, onStateChange };
}
