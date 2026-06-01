# Resource Monitor Graphs Feature

## Overview

Added real-time historical graph visualization to the Resource Monitor page using Apache ECharts.

## What's New

### 📊 Historical Graphs Tab

- New tab-based interface with "Current Stats" and "Historical Graphs"
- Real-time data collection (updates every 2 seconds)
- Stores last 60 data points (2 minutes of history)
- Smooth animations and interactive charts

### 📈 Available Charts

1. **CPU Usage** - Percentage over time with color-coded thresholds
2. **Memory Usage** - Percentage over time
3. **Disk Usage** - Percentage over time
4. **Memory (MB)** - Absolute memory usage in megabytes
5. **Running Servers** - Count of active servers
6. **Goroutines** - Go runtime goroutine count

### ✨ Chart Features

- **Interactive Tooltips** - Hover to see exact values
- **Zoom & Pan** - Built-in data zoom controls
- **Smooth Animations** - Real-time updates with smooth transitions
- **Area Gradients** - Beautiful gradient fills under lines
- **Responsive Design** - Adapts to screen size
- **Dark Theme** - Matches application theme

## Technical Implementation

### New Files Created

```
frontend/src/
├── composables/
│   └── useResourceHistory.js          # Data buffer management
└── components/charts/
    ├── ResourceChart.vue              # Reusable chart component
    └── ResourceGraphsView.vue         # Graphs container
```

### Modified Files

- `ResourceMonitorView.vue` - Added tabs and history tracking

### Dependencies Added

- `echarts` (^5.x) - Charting library
- `vue-echarts` (^7.x) - Vue 3 wrapper for ECharts

## Usage

### For Users

1. Navigate to **Resource Monitor** page
2. Click **"Historical Graphs"** tab
3. Charts will populate as data is collected
4. Use mouse wheel to zoom, drag to pan
5. Hover over charts for detailed tooltips

### For Developers

**Using the History Composable:**

```javascript
import { useResourceHistory } from "@/composables/useResourceHistory";

const history = useResourceHistory(60); // 60 data points

// Add data point
history.addDataPoint(snapshot);

// Get chart data
const cpuData = history.getChartData("cpu");

// Clear history
history.clear();
```

**Using the Chart Component:**

```vue
<ResourceChart
  title="CPU Usage"
  :data="cpuData"
  :time-labels="timeLabels"
  y-axis-label="Percentage (%)"
  color="#f59e0b"
  :max-value="100"
  :show-data-zoom="true"
/>
```

## Configuration

### Adjust History Duration

In `ResourceMonitorView.vue`:

```javascript
const resourceHistory = useResourceHistory(60); // Change 60 to desired data points
```

**Examples:**

- 30 points = 1 minute (at 2s intervals)
- 60 points = 2 minutes (default)
- 150 points = 5 minutes
- 300 points = 10 minutes

### Customize Chart Colors

In `ResourceGraphsView.vue`, modify the `color` prop:

```vue
<ResourceChart color="#your-color-hex" />
```

## Performance Notes

- **Memory Usage**: ~1-2MB for 60 data points across all metrics
- **Bundle Size**: +570KB (ECharts library)
- **Update Frequency**: 2 seconds (matches WebSocket interval)
- **Optimization**: Uses LTTB sampling for smooth rendering

## Future Enhancements

Potential improvements:

- [ ] Time range selector (2m / 5m / 10m)
- [ ] Export charts as PNG/SVG
- [ ] Download historical data as CSV
- [ ] Alert threshold markers
- [ ] Backend history endpoint for longer ranges
- [ ] Metric comparison (overlay multiple metrics)
- [ ] Custom date range picker

## Browser Compatibility

- Chrome/Edge: ✅ Full support
- Firefox: ✅ Full support
- Safari: ✅ Full support
- Mobile: ✅ Responsive design

## Troubleshooting

**Charts not appearing:**

- Wait 4-6 seconds for data collection to start
- Check WebSocket connection status (should show "LIVE")
- Verify browser console for errors

**Performance issues:**

- Reduce history buffer size (use 30 instead of 60)
- Close other browser tabs
- Check system resources

**Build warnings:**

- Large chunk size warning is expected (ECharts library)
- Consider code-splitting if bundle size is critical

## Credits

- **ECharts**: Apache ECharts visualization library
- **vue-echarts**: Vue 3 integration by ecomfe
