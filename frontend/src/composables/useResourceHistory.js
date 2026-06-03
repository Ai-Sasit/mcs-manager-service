import { ref, computed } from "vue";

/**
 * Manages historical resource data for charting
 * @param {number} maxDataPoints - Maximum number of data points to retain (default: 60 = 2 minutes at 2s intervals)
 */
export function useResourceHistory(maxDataPoints = 60) {
  const history = ref({
    timestamps: [],
    cpu: [],
    memory: [],
    disk: [],
    memoryUsedMB: [],
    memoryTotalMB: [],
    diskUsedMB: [],
    diskTotalMB: [],
    serverRunning: [],
    serverTotal: [],
    goroutines: [],
  });

  /**
   * Add a new data point to the history
   * @param {Object} snapshot - Resource snapshot from useSystemResources
   */
  function addDataPoint(snapshot) {
    if (!snapshot) return;

    const timestamp = new Date();
    const node = snapshot.node || {};
    const runtime = snapshot.runtime || {};
    const servers = snapshot.servers || {};

    // Clone existing arrays and append new data points
    const updated = {
      timestamps: [...history.value.timestamps, timestamp],
      cpu: [...history.value.cpu, node.cpu_percent ?? 0],
      memory: [...history.value.memory, node.memory_percent ?? 0],
      disk: [...history.value.disk, node.disk_percent ?? 0],
      memoryUsedMB: [...history.value.memoryUsedMB, node.memory_used_mb ?? 0],
      memoryTotalMB: [...history.value.memoryTotalMB, node.memory_total_mb ?? 0],
      diskUsedMB: [...history.value.diskUsedMB, node.disk_used_mb ?? 0],
      diskTotalMB: [...history.value.diskTotalMB, node.disk_total_mb ?? 0],
      serverRunning: [...history.value.serverRunning, servers.running ?? 0],
      serverTotal: [...history.value.serverTotal, servers.total ?? 0],
      goroutines: [...history.value.goroutines, runtime.goroutines ?? 0],
    };

    // Trim to max data points (circular buffer behavior)
    if (updated.timestamps.length > maxDataPoints) {
      Object.keys(updated).forEach((key) => {
        updated[key] = updated[key].slice(1);
      });
    }

    // Replace entire object to ensure reactivity through prop chains
    history.value = updated;
  }

  /**
   * Clear all historical data
   */
  function clear() {
    Object.keys(history.value).forEach((key) => {
      history.value[key] = [];
    });
  }

  /**
   * Get chart data for a specific metric
   * @param {string} metric - Metric name (cpu, memory, disk, etc.)
   * @returns {Array} Array of [timestamp, value] pairs
   */
  function getChartData(metric) {
    if (!history.value[metric]) return [];
    return history.value.timestamps.map((ts, i) => [ts, history.value[metric][i]]);
  }

  /**
   * Get formatted time labels for x-axis
   */
  const timeLabels = computed(() => {
    return history.value.timestamps.map((ts) => {
      const date = new Date(ts);
      return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
      });
    });
  });

  /**
   * Get data point count
   */
  const dataPointCount = computed(() => history.value.timestamps.length);

  /**
   * Get time range in seconds
   */
  const timeRangeSeconds = computed(() => {
    if (history.value.timestamps.length < 2) return 0;
    const first = history.value.timestamps[0];
    const last = history.value.timestamps[history.value.timestamps.length - 1];
    return Math.floor((last - first) / 1000);
  });

  return {
    history,
    addDataPoint,
    clear,
    getChartData,
    timeLabels,
    dataPointCount,
    timeRangeSeconds,
  };
}
