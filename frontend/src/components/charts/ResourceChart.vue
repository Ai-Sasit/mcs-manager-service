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
import { useTheme } from "@/composables/useTheme";

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
    default: "#00856f",
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
const { theme } = useTheme();

const palette = computed(() => {
  // The reactive theme read makes ECharts redraw when the app appearance changes.
  theme.value;
  const styles = getComputedStyle(document.documentElement);
  const color = (name) => styles.getPropertyValue(name).trim();
  return {
    layer: color("--color-layer"),
    text: color("--color-text"),
    secondaryText: color("--color-text-secondary"),
    mutedText: color("--color-text-muted"),
    border: color("--color-border"),
    borderLight: color("--color-border-light"),
  };
});

const chartOption = computed(() => {
  const areaColor = props.areaStyle ? props.color : undefined;
  const colors = palette.value;
  return {
    tooltip: {
      trigger: "axis",
      backgroundColor: colors.layer,
      borderColor: colors.border,
      borderWidth: 1,
      padding: [8, 12],
      textStyle: {
        color: colors.text,
        fontSize: 12,
      },
      axisPointer: {
        type: "line",
        lineStyle: {
          color: colors.mutedText,
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
          color: colors.borderLight,
          width: 1,
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: colors.secondaryText,
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
          color: colors.borderLight,
          width: 1,
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: colors.secondaryText,
        fontSize: 11,
      },
      splitLine: {
        lineStyle: {
          color: colors.borderLight,
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
          borderColor: colors.layer,
        },
        lineStyle: {
          width: 2,
          color: props.color,
        },
        areaStyle: props.areaStyle
          ? {
              color: areaColor,
              opacity: 0.6,
            }
          : null,
        data: props.data,
      },
    ],
  };
});

// Force chart to re-render on data changes (deep watch on data + labels)
watch(
  () => [props.data, props.timeLabels, theme.value],
  () => {
    const chart = chartRef.value?.chart;
    if (chart) {
      chart.setOption(chartOption.value, { notMerge: true });
    }
  },
  { deep: true },
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
