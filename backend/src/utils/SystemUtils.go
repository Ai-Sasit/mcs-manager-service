package utils

import (
	"fmt"
	"mc-manage-backend/src/models"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func LookupProcessByPort(port int) (*models.ProcessInfo, error) {
	if runtime.GOOS == "windows" {
		return lookupProcessWindows(port)
	} else {
		return lookupProcessLinux(port)
	}
}

func lookupProcessWindows(port int) (*models.ProcessInfo, error) {
	// Find PID using netstat
	cmd := exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano | findstr LISTENING | findstr :%d", port))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("no process found listening on port %d", port)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		// The address field is fields[1], e.g. "0.0.0.0:25565"
		if strings.HasSuffix(fields[1], fmt.Sprintf(":%d", port)) {
			pidStr := fields[len(fields)-1]
			pid, _ := strconv.Atoi(strings.TrimSpace(pidStr))
			name := getProcessNameWindows(pid)
			return &models.ProcessInfo{
				Pid:      pid,
				Name:     name,
				Port:     port,
				Protocol: fields[0],
			}, nil
		}
	}

	return nil, fmt.Errorf("no process found listening on port %d", port)
}

func getProcessNameWindows(pid int) string {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
	out, _ := cmd.Output()
	fields := strings.Fields(string(out))
	if len(fields) > 0 {
		return fields[0]
	}
	return "Unknown"
}

func lookupProcessLinux(port int) (*models.ProcessInfo, error) {
	cmd := exec.Command("sudo", "lsof", "-t", fmt.Sprintf("-i:%d", port))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("no process found listening on port %d", port)
	}

	pidStr := strings.TrimSpace(string(out))
	if pidStr == "" {
		return nil, fmt.Errorf("no process found listening on port %d", port)
	}

	// Take the first PID if multiple lines are returned
	pids := strings.Split(pidStr, "\n")
	firstPid, err := strconv.Atoi(strings.TrimSpace(pids[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PID for port %d: %w", port, err)
	}

	name := getProcessNameLinux(firstPid)
	return &models.ProcessInfo{
		Pid:      firstPid,
		Name:     name,
		Port:     port,
		Protocol: "TCP",
	}, nil
}

func getProcessNameLinux(pid int) string {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func KillProcessByPid(pid int) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	} else {
		cmd = exec.Command("sudo", "kill", "-9", strconv.Itoa(pid))
	}
	return cmd.Run()
}
