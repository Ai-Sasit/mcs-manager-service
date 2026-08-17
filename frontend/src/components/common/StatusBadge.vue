<template>
  <span class="status-badge" :class="status">
    <span class="dot"></span>
    <span class="label">{{ displayText }}</span>
  </span>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({ status: String });

const displayText = computed(() => {
  const map = {
    running: "Running",
    stopped: "Stopped",
    starting: "Starting",
  };
  return map[props.status] || props.status;
});
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0;
}

.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.running {
  background: var(--green-pale);
  color: var(--green);
  border-color: color-mix(in srgb, var(--green) 28%, transparent);
}

.running .dot {
  background: var(--green);
  animation: pulse-dot 2s infinite;
}

.stopped {
  background: var(--color-layer-alt);
  color: var(--text-muted);
  border-color: var(--color-border);
}

.stopped .dot {
  background: var(--text-muted);
}

.starting {
  background: var(--yellow-pale);
  color: var(--yellow);
  border-color: color-mix(in srgb, var(--yellow) 30%, transparent);
}

.starting .dot {
  background: var(--yellow);
  animation: pulse-dot 1s infinite;
}
</style>
