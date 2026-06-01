package controllers

import (
	"bufio"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// GetSystemInfo returns basic system/runtime information
func GetSystemInfo(c fiber.Ctx) error {
	snapshot, err := services.CollectResourceSnapshot(state)
	if err != nil {
		return utils.ErrorResponse(c, "failed to collect system info: "+err.Error(), fiber.StatusInternalServerError)
	}
	return utils.SuccessResponse(c, "OK", snapshot)
}

// GetAuditLogs returns audit log entries
func GetAuditLogs(c fiber.Ctx) error {
	logs := utils.GetAuditLogs()
	return utils.SuccessResponse(c, "OK", logs)
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
	return utils.SuccessResponse(c, "OK", files)
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

	return utils.SuccessResponse(c, "OK", lines)
}

// LookupPort finds a process by its listening port
func LookupPort(c fiber.Ctx) error {
	portStr := c.Query("port")
	if portStr == "" {
		return utils.ErrorResponse(c, "port query param required", fiber.StatusBadRequest)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return utils.ErrorResponse(c, "invalid port number", fiber.StatusBadRequest)
	}

	info, err := utils.LookupProcessByPort(port)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}

	return utils.SuccessResponse(c, "OK", info)
}

// KillPid force kills a process by PID
func KillPid(c fiber.Ctx) error {
	var req struct {
		Pid int `json:"pid"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	if req.Pid <= 0 {
		return utils.ErrorResponse(c, "invalid PID", fiber.StatusBadRequest)
	}

	if err := utils.KillProcessByPid(req.Pid); err != nil {
		return utils.ErrorResponse(c, "failed to kill process: "+err.Error(), fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "SYSTEM_KILL_PID", strconv.Itoa(req.Pid), "Force killed system process.")
	return utils.SuccessResponse(c, "Process terminated", nil)
}
