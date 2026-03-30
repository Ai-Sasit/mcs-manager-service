package controllers

import (
	"mc-manage-backend/src/utils"
	"runtime"

	"github.com/gofiber/fiber/v3"
)

// GetSystemInfo returns basic system/runtime information
func GetSystemInfo(c fiber.Ctx) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"os":              runtime.GOOS,
			"arch":            runtime.GOARCH,
			"go_version":      runtime.Version(),
			"goroutines":      runtime.NumGoroutine(),
			"cpu_count":       runtime.NumCPU(),
			"mem_alloc_mb":    mem.Alloc / 1024 / 1024,
			"mem_sys_mb":      mem.Sys / 1024 / 1024,
			"mem_gc_cycles":   mem.NumGC,
		},
		"message": "OK",
	})
}

// GetAuditLogs returns audit log entries
func GetAuditLogs(c fiber.Ctx) error {
	logs := utils.GetAuditLogs()
	return c.JSON(fiber.Map{
		"success": true,
		"data":    logs,
		"message": "OK",
	})
}
