<template>
  <div class="resource-graphs">
    <div class="graphs-header"></div>

    <div v-if="dataPointCount === 0" class="no-data">
      <p>Collecting data...</p>
    </div>

    <div v-else class="charts-grid">
      <div class="chart-card">
        <div class="chart-header">
          <h4>CPU Usage</h4>
          <span class="chart-unit">%</span>
        </div>
        <ResourceChart
          title="CPU Usage"
          :data="history.cpu"
          :time-labels="timeLabels"
          color="#f59e0b"
          :max-value="100"
          :show-data-zoom="true"
        />
      </div>

      <div class="chart-card">
        <div class="chart-header">
          <h4>Memory Usage</h4>
          <span class="chart-unit">%</span>
        </div>
        <ResourceChart
          title="Memory Usage"
          :data="history.memory"
          :time-labels="timeLabels"
          :max-value="100"
          :show-data-zoom="true"
        />
      </div>

      <div class="chart-card">
        <div class="chart-header">
          <h4>Disk Usage</h4>
          <span class="chart-unit">%</span>
        </div>
        <ResourceChart
          title="Disk Usage"
          :data="history.disk"
          :time-labels="timeLabels"
          color="#3b82f6"
          :max-value="100"
          :show-data-zoom="true"
        />
      </div>

      <div class="chart-card">
        <div class="chart-header">
          <h4>Memory</h4>
          <span class="chart-unit">MB</span>
        </div>
        <ResourceChart
          title="Memory (MB)"
          :data="history.memoryUsedMB"
          :time-labels="timeLabels"
          color="#8b5cf6"
          :show-data-zoom="true"
        />
      </div>

      <div class="chart-card">
        <div class="chart-header">
          <h4>Running Servers</h4>
          <span class="chart-unit">count</span>
        </div>
        <ResourceChart
          title="Running Servers"
          :data="history.serverRunning"
          :time-labels="timeLabels"
          color="#ec4899"
          :smooth="false"
          :area-style="false"
          :show-data-zoom="true"
        />
      </div>

      <div class="chart-card">
        <div class="chart-header">
          <h4>Goroutines</h4>
          <span class="chart-unit">count</span>
        </div>
        <ResourceChart
          title="Goroutines"
          :data="history.goroutines"
          :time-labels="timeLabels"
          color="#06b6d4"
          :show-data-zoom="true"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, toRefs } from "vue";
import ResourceChart from "./ResourceChart.vue";

const props = defineProps({
  history: {
    type: Object,
    required: true,
  },
  timeLabels: {
    type: Array,
    required: true,
  },
  dataPointCount: {
    type: Number,
    required: true,
  },
  timeRangeSeconds: {
    type: Number,
    required: true,
  },
});

// Ensure reactivity by using toRefs
const { history, timeLabels, dataPointCount, timeRangeSeconds } = toRefs(props);

const timeRangeLabel = computed(() => {
  const seconds = timeRangeSeconds.value;
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return remainingSeconds > 0
    ? `${minutes}m ${remainingSeconds}s`
    : `${minutes}m`;
});
</script>

<style scoped>
.resource-graphs {
  width: 100%;
}

.graphs-header {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-bottom: 1.25rem;
}

.graphs-info {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.info-label {
  font-weight: 500;
}

.info-separator {
  color: var(--color-text-secondary);
}

.no-data {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--color-text-secondary);
}

.no-data p {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 500;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(480px, 1fr));
  gap: 1.25rem;
}

.chart-card {
  background: var(--color-layer);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1rem;
  height: 360px;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-card);
  transition: border-color 160ms ease, box-shadow 160ms ease;
  overflow: hidden;
}

.chart-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-card-hover);
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding: 0 0.25rem;
  flex-shrink: 0;
}

.chart-header h4 {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text);
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.chart-unit {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-text-muted);
  text-transform: uppercase;
}

.chart-card > .resource-chart {
  flex: 1;
  min-height: 0;
}

@media (max-width: 1024px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .chart-card {
    height: 320px;
    padding: 0.875rem;
  }

  .graphs-header {
    margin-bottom: 1rem;
  }
}
</style>
