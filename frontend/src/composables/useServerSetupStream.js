import { ref } from "vue";
import { getToken } from "@/utils/authStorage";

export function useServerSetupStream() {
  const state = ref("idle"); // idle | connecting | open | closed | error
  const error = ref("");
  const setupEvents = ref([]);
  const setupPercent = ref(0);
  let abortController = null;

  async function start(params) {
    setupEvents.value = [];
    setupPercent.value = 0;
    error.value = "";
    state.value = "connecting";
    abortController = new AbortController();

    try {
      const token = getToken();
      const baseUrl = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";
      const response = await fetch(`${baseUrl}/servers/create-stream`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: token ? `Bearer ${token}` : "",
        },
        body: JSON.stringify(params),
        signal: abortController.signal,
      });

      if (!response.ok) {
        const text = await response.text();
        let message = text;
        try {
          const parsed = JSON.parse(text);
          message = parsed.message || text;
        } catch {
          // use raw text
        }
        throw new Error(message || `Request failed with status ${response.status}`);
      }

      state.value = "open";

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        // Keep the last partial line in the buffer
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const jsonStr = line.substring(6);
            try {
              const event = JSON.parse(jsonStr);
              if (event.percent != null) {
                setupPercent.value = Math.max(setupPercent.value, event.percent);
              }
              setupEvents.value.push(event);

              if (event.status === "failed") {
                error.value = event.message || "Server setup failed";
                state.value = "error";
                return event;
              }

              if (event.status === "success" && event.step === "complete") {
                state.value = "closed";
                return event;
              }
            } catch {
              // Skip malformed JSON lines
            }
          }
        }
      }

      // Stream ended without a terminal event
      state.value = "closed";
      return null;
    } catch (err) {
      if (err.name === "AbortError") {
        state.value = "closed";
        return null;
      }
      error.value = err.message || "SSE stream error";
      state.value = "error";
      return null;
    }
  }

  function disconnect() {
    if (abortController) {
      abortController.abort();
      abortController = null;
    }
    state.value = "closed";
  }

  return {
    state,
    error,
    setupEvents,
    setupPercent,
    start,
    disconnect,
  };
}