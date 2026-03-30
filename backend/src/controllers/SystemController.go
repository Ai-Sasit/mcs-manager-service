package controllers

import (
	"bufio"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// GetSystemInfo returns basic system/runtime information
func GetSystemInfo(c fiber.Ctx) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
			"go_version":    runtime.Version(),
			"goroutines":    runtime.NumGoroutine(),
			"cpu_count":     runtime.NumCPU(),
			"mem_alloc_mb":  mem.Alloc / 1024 / 1024,
			"mem_sys_mb":    mem.Sys / 1024 / 1024,
			"mem_gc_cycles": mem.NumGC,
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

// ListBackendLogFiles returns available backend log file names (newest first)
func ListBackendLogFiles(c fiber.Ctx) error {
	entries, err := os.ReadDir("logs")
	if err != nil {
		return c.JSON(fiber.Map{"success": true, "data": []string{}})
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			files = append(files, e.Name())
		}
	}
	// Sort descending (newest date first)
	sort.Slice(files, func(i, j int) bool { return files[i] > files[j] })
	return c.JSON(fiber.Map{"success": true, "data": files})
}

// GetBackendLogFile returns the lines of a specific backend log file.
// Query params:
//   - file  : filename (basename only, e.g. "2026-03-30-mc-manage.log")
//   - tail  : number of lines from the end (default 500, max 5000)
func GetBackendLogFile(c fiber.Ctx) error {
	filename := filepath.Base(c.Query("file"))
	if filename == "" || filename == "." {
		return utils.ErrorResponse(c, "file query param required", fiber.StatusBadRequest)
	}
	// Reject path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.HasPrefix(filename, ".") {
		return utils.ErrorResponse(c, "invalid filename", fiber.StatusBadRequest)
	}

	tail, _ := strconv.Atoi(c.Query("tail", "500"))
	if tail < 1 {
		tail = 500
	}
	if tail > 5000 {
		tail = 5000
	}

	path := filepath.Join("logs", filename)
	f, err := os.Open(path)
	if err != nil {
		return utils.ErrorResponse(c, "log file not found", fiber.StatusNotFound)
	}
	defer f.Close()

	// Read all lines into a ring buffer of size `tail`
	buf := make([]string, tail)
	pos := 0
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		buf[pos%tail] = scanner.Text()
		pos++
		count++
	}

	var lines []string
	if count <= tail {
		lines = buf[:count]
	} else {
		// Oldest entry is at pos%tail; wrap around
		lines = make([]string, tail)
		for i := 0; i < tail; i++ {
			lines[i] = buf[(pos+i)%tail]
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": lines})
}
