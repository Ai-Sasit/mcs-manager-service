# Bugfix Requirements Document

## Introduction

The resource monitoring graphs in the admin panel are experiencing reactivity issues that prevent them from displaying real-time data correctly. Despite WebSocket receiving resource snapshots every 2 seconds, the 6 charts (CPU, Memory, Disk, Memory MB, Running Servers, Goroutines) remain static or update inconsistently. The root cause is a broken Vue 3 reactivity chain caused by manual `.value` unwrapping when passing props to child components, combined with inefficient watch mechanisms and lack of error visibility.

This bugfix ensures that:

- Real-time data flows reactively from WebSocket → useResourceHistory → ResourceGraphsView → ResourceChart
- All 6 charts update smoothly every 2 seconds as new data arrives
- Users receive clear error feedback when WebSocket connection fails
- The Vue 3 reactivity system is properly leveraged throughout the component hierarchy

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN ResourceMonitorView passes `resourceHistory.history.value` to ResourceGraphsView THEN the reactivity chain is broken because manual `.value` unwrapping loses the reactive reference

1.2 WHEN useResourceHistory mutates arrays using `push()` and `shift()` operations THEN the ResourceChart watch mechanism may not detect these mutations consistently

1.3 WHEN the ResourceChart watch only monitors `props.data` THEN it does not trigger when the computed `chartOption` changes due to other prop updates

1.4 WHEN WebSocket connection fails or encounters errors THEN users see "Collecting data..." message indefinitely with no error indication

1.5 WHEN debugging data flow issues THEN developers have no logging or visibility into whether data is being received, processed, or rendered

### Expected Behavior (Correct)

2.1 WHEN ResourceMonitorView passes reactive data to ResourceGraphsView THEN it SHALL pass the reactive ref directly without manual `.value` unwrapping to preserve reactivity

2.2 WHEN useResourceHistory mutates arrays using `push()` and `shift()` operations THEN the changes SHALL trigger reactivity updates in all dependent components

2.3 WHEN ResourceChart receives new data via props THEN it SHALL detect changes and update the chart visualization immediately

2.4 WHEN WebSocket connection fails or encounters errors THEN the system SHALL display a clear error message to users instead of showing "Collecting data..." indefinitely

2.5 WHEN data flows through the system THEN the system SHALL provide optional debug logging to help troubleshoot data flow issues

### Unchanged Behavior (Regression Prevention)

3.1 WHEN WebSocket successfully connects and receives data THEN the system SHALL CONTINUE TO store data in the circular buffer with a maximum of 60 data points

3.2 WHEN displaying the "Current Stats" tab THEN the system SHALL CONTINUE TO show real-time statistics with progress bars and formatted values

3.3 WHEN formatting memory and disk values THEN the system SHALL CONTINUE TO display values in MB or GB with appropriate precision

3.4 WHEN displaying time labels on chart x-axis THEN the system SHALL CONTINUE TO format timestamps as HH:MM:SS in 24-hour format

3.5 WHEN users switch between "Current Stats" and "Historical Graphs" tabs THEN the system SHALL CONTINUE TO maintain data collection in the background

3.6 WHEN charts are rendered with ECharts THEN the system SHALL CONTINUE TO support all existing chart features (tooltips, data zoom, smooth lines, area styles, colors)

3.7 WHEN the component is unmounted THEN the system SHALL CONTINUE TO properly disconnect the WebSocket connection to prevent memory leaks
