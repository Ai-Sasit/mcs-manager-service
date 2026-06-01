<template>
  <div class="resource-chart">
    <v-chart
      ref="chartRef"
      :option="chartOption"
      :autoresize="true"
      class="chart-instance"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { LineChart } from "echarts/charts";
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
} from "echarts/components";
import VChart from "vue-echarts";

// Register ECharts components
use([
  CanvasRenderer,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
]);

const props = defineProps({
  title: {
    type: String,
    default: "",
  },
  data: {
    type: Array,
    default: () => [],
  },
  timeLabels: {
    type: Array,
    default: () => [],
  },
  yAxisLabel: {
    type: String,
    default: "",
  },
  color: {
    type: String,
    default: "#10b981",
  },
  showDataZoom: {
    type: Boolean,
    default: false,
  },
  maxValue: {
    type: Number,
    default: null,
  },
  minValue: {
    type: Number,
    default: 0,
  },
  smooth: {
    type: Boolean,
    default: true,
  },
  areaStyle: {
    type: Boolean,
    default: true,
  },
});

const chartRef = ref(null);

const chartOption = computed(() => ({
  tooltip: {
    trigger: "axis",
    backgroundColor: "rgba(255, 255, 255, 0.98)",
    borderColor: "#e5e7eb",
    borderWidth: 1,
    padding: [8, 12],
    textStyle: {
      color: "#111827",
      fontSize: 12,
    },
    axisPointer: {
      type: "line",
      lineStyle: {
        color: "#9ca3af",
        width: 1,
        type: "solid",
      },
    },
  },
  grid: {
    left: "2%",
    right: "2%",
    bottom: props.showDataZoom ? "12%" : "2%",
    top: "8%",
    containLabel: true,
  },
  xAxis: {
    type: "category",
    boundaryGap: false,
    data: props.timeLabels,
    axisLine: {
      show: true,
      lineStyle: {
        color: "#e5e7eb",
        width: 1,
      },
    },
    axisTick: {
      show: false,
    },
    axisLabel: {
      color: "#6b7280",
      fontSize: 11,
      interval: Math.floor(props.timeLabels.length / 8) || 0,
      hideOverlap: true,
    },
  },
  yAxis: {
    type: "value",
    min: props.minValue,
    max: props.maxValue,
    axisLine: {
      show: true,
      lineStyle: {
        color: "#e5e7eb",
        width: 1,
      },
    },
    axisTick: {
      show: false,
    },
    axisLabel: {
      color: "#6b7280",
      fontSize: 11,
    },
    splitLine: {
      lineStyle: {
        color: "#e5e7eb",
        width: 1,
        type: "solid",
      },
    },
  },
  dataZoom: props.showDataZoom
    ? [
        {
          type: "inside",
          start: 0,
          end: 100,
          zoomOnMouseWheel: true,
          moveOnMouseMove: true,
        },
      ]
    : [],
  series: [
    {
      name: props.title,
      type: "line",
      smooth: props.smooth,
      symbol: "circle",
      symbolSize: 6,
      sampling: "lttb",
      itemStyle: {
        color: props.color,
        borderWidth: 2,
        borderColor: "#111827",
      },
      lineStyle: {
        width: 2,
        color: props.color,
      },
      areaStyle: props.areaStyle
        ? {
            color: props.color,
            opacity: 0.6,
          }
        : null,
      data: props.data,
    },
  ],
}));

// Watch for chart option changes and update chart incrementally
watch(
  () => props.data,
  (newData, oldData) => {
    if (chartRef.value && chartRef.value.chart) {
      // If data length changed, update the entire chart
      if (!oldData || newData.length !== oldData.length) {
        chartRef.value.setOption(
          {
            xAxis: {
              data: props.timeLabels,
            },
            series: [
              {
                data: newData,
              },
            ],
          },
          {
            replaceMerge: ["xAxis", "series"],
          },
        );
      }
    }
  },
  { deep: false },
);

// Watch for time labels changes
watch(
  () => props.timeLabels,
  (newLabels) => {
    if (chartRef.value && chartRef.value.chart) {
      chartRef.value.setOption({
        xAxis: {
          data: newLabels,
        },
      });
    }
  },
  { deep: false },
);
</script>

<style scoped>
.resource-chart {
  width: 100%;
  height: 280px;
  position: relative;
  overflow: hidden;
}

.chart-instance {
  width: 100% !important;
  height: 100% !important;
}

.resource-chart :deep(.echarts) {
  width: 100% !important;
  height: 100% !important;
}

.resource-chart :deep(canvas) {
  width: 100% !important;
  height: 100% !important;
}
</style>
