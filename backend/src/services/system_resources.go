package services

import (
	"mc-manage-backend/src/models"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type ResourceSnapshot struct {
	Node    NodeMetrics    `json:"node"`
	Runtime RuntimeMetrics `json:"runtime"`
	Servers ServerMetrics  `json:"servers"`
}

type NodeMetrics struct {
	CPUPercent    float64  `json:"cpu_percent"`
	MemoryUsedMB  uint64   `json:"memory_used_mb"`
	MemoryTotalMB uint64   `json:"memory_total_mb"`
	MemoryPercent float64  `json:"memory_percent"`
	DiskUsedMB    uint64   `json:"disk_used_mb"`
	DiskTotalMB   uint64   `json:"disk_total_mb"`
	DiskPercent   float64  `json:"disk_percent"`
	Load1         *float64 `json:"load_1"`
	Load5         *float64 `json:"load_5"`
	Load15        *float64 `json:"load_15"`
}

type RuntimeMetrics struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	GoVersion   string `json:"go_version"`
	Goroutines  int    `json:"goroutines"`
	CPUCount    int    `json:"cpu_count"`
	MemAllocMB  uint64 `json:"mem_alloc_mb"`
	MemSysMB    uint64 `json:"mem_sys_mb"`
	MemGCCycles uint32 `json:"mem_gc_cycles"`
}

type ServerMetrics struct {
	Total             int    `json:"total"`
	Running           int    `json:"running"`
	AllocatedMemoryMB uint64 `json:"allocated_memory_mb"`
}

func CollectResourceSnapshot(state *AppState) (ResourceSnapshot, error) {
	var runtimeMem runtime.MemStats
	runtime.ReadMemStats(&runtimeMem)

	cpuPercent := 0.0
	cpuPercents, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		return ResourceSnapshot{}, err
	}
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return ResourceSnapshot{}, err
	}

	diskStats, err := disk.Usage(state.DataDir)
	if err != nil {
		return ResourceSnapshot{}, err
	}

	var load1, load5, load15 *float64
	if loadStats, err := load.Avg(); err == nil && loadStats != nil {
		load1 = &loadStats.Load1
		load5 = &loadStats.Load5
		load15 = &loadStats.Load15
	}

	servers := state.ListServers()
	serverMetrics := ServerMetrics{Total: len(servers)}
	for _, server := range servers {
		if server.Status == models.StatusRunning {
			serverMetrics.Running++
			serverMetrics.AllocatedMemoryMB += uint64(server.MemoryMB)
		}
	}

	return ResourceSnapshot{
		Node: NodeMetrics{
			CPUPercent:    cpuPercent,
			MemoryUsedMB:  memStats.Used / 1024 / 1024,
			MemoryTotalMB: memStats.Total / 1024 / 1024,
			MemoryPercent: memStats.UsedPercent,
			DiskUsedMB:    diskStats.Used / 1024 / 1024,
			DiskTotalMB:   diskStats.Total / 1024 / 1024,
			DiskPercent:   diskStats.UsedPercent,
			Load1:         load1,
			Load5:         load5,
			Load15:        load15,
		},
		Runtime: RuntimeMetrics{
			OS:          runtime.GOOS,
			Arch:        runtime.GOARCH,
			GoVersion:   runtime.Version(),
			Goroutines:  runtime.NumGoroutine(),
			CPUCount:    runtime.NumCPU(),
			MemAllocMB:  runtimeMem.Alloc / 1024 / 1024,
			MemSysMB:    runtimeMem.Sys / 1024 / 1024,
			MemGCCycles: runtimeMem.NumGC,
		},
		Servers: serverMetrics,
	}, nil
}
