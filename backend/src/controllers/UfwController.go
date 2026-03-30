package controllers

import (
	"fmt"
	"mc-manage-backend/src/utils"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// GetUfwStatus returns whether UFW is available, enabled/disabled, and current rules
func GetUfwStatus(c fiber.Ctx) error {
	if runtime.GOOS != "linux" {
		return c.JSON(fiber.Map{
			"success":   true,
			"available": false,
			"message":   "UFW is only available on Linux",
		})
	}

	// Check if ufw is installed
	if _, err := exec.LookPath("ufw"); err != nil {
		return c.JSON(fiber.Map{
			"success":   true,
			"available": false,
			"message":   "UFW is not installed",
		})
	}

	out, err := exec.Command("sudo", "ufw", "status").CombinedOutput()
	if err != nil {
		return c.JSON(fiber.Map{
			"success":   true,
			"available": true,
			"enabled":   false,
			"message":   "Failed to get UFW status: " + string(out),
		})
	}

	output := string(out)
	enabled := strings.Contains(output, "Status: active")

	return c.JSON(fiber.Map{
		"success":   true,
		"available": true,
		"enabled":   enabled,
	})
}

type UfwRule struct {
	Number int    `json:"number"`
	To     string `json:"to"`
	Action string `json:"action"`
	From   string `json:"from"`
	Raw    string `json:"raw"`
}

// ListUfwRules returns parsed UFW rules
func ListUfwRules(c fiber.Ctx) error {
	if runtime.GOOS != "linux" {
		return c.JSON(fiber.Map{
			"success":   true,
			"available": false,
			"data":      []UfwRule{},
		})
	}

	out, err := exec.Command("sudo", "ufw", "status", "numbered").CombinedOutput()
	if err != nil {
		return utils.ErrorResponse(c, "Failed to list rules: "+string(out), fiber.StatusInternalServerError)
	}

	rules := parseUfwRules(string(out))
	return c.JSON(fiber.Map{
		"success":   true,
		"available": true,
		"data":      rules,
	})
}

func parseUfwRules(output string) []UfwRule {
	var rules []UfwRule
	// Match lines like: [ 1] 25565/tcp                  ALLOW IN    Anywhere
	re := regexp.MustCompile(`\[\s*(\d+)\]\s+(.+?)\s+(ALLOW|DENY|REJECT|LIMIT)\s+(IN|OUT|FWD)?\s*(.*)`)

	for _, line := range strings.Split(output, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			num, _ := strconv.Atoi(matches[1])
			to := strings.TrimSpace(matches[2])
			action := strings.TrimSpace(matches[3])
			from := strings.TrimSpace(matches[5])
			if from == "" {
				from = "Anywhere"
			}
			rules = append(rules, UfwRule{
				Number: num,
				To:     to,
				Action: action,
				From:   from,
				Raw:    strings.TrimSpace(line),
			})
		}
	}
	return rules
}

// AllowUfwRule adds an allow rule
func AllowUfwRule(c fiber.Ctx) error {
	return addUfwRule(c, "allow")
}

// DenyUfwRule adds a deny rule
func DenyUfwRule(c fiber.Ctx) error {
	return addUfwRule(c, "deny")
}

func addUfwRule(c fiber.Ctx, action string) error {
	if runtime.GOOS != "linux" {
		return utils.ErrorResponse(c, "UFW is only available on Linux", fiber.StatusBadRequest)
	}

	var req struct {
		Port     int    `json:"port"`
		Protocol string `json:"protocol"` // "tcp", "udp", or "" for both
	}
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	// Strict validation
	if req.Port < 1 || req.Port > 65535 {
		return utils.ErrorResponse(c, "Port must be between 1 and 65535", fiber.StatusBadRequest)
	}

	proto := strings.ToLower(req.Protocol)
	if proto != "" && proto != "tcp" && proto != "udp" {
		return utils.ErrorResponse(c, "Protocol must be 'tcp', 'udp', or empty for both", fiber.StatusBadRequest)
	}

	rule := fmt.Sprintf("%d", req.Port)
	if proto != "" {
		rule = fmt.Sprintf("%d/%s", req.Port, proto)
	}

	out, err := exec.Command("sudo", "ufw", action, rule).CombinedOutput()
	if err != nil {
		return utils.ErrorResponse(c, "Failed: "+string(out), fiber.StatusInternalServerError)
	}

	username, _ := c.Locals("username").(string)
	utils.LogAudit(username, "UFW_"+strings.ToUpper(action), rule, string(out))
	logger.Info(fmt.Sprintf("[UFW] %s %s by %s", action, rule, username), nil)

	return c.JSON(fiber.Map{
		"success": true,
		"message": strings.TrimSpace(string(out)),
	})
}

// DeleteUfwRule deletes a rule by number
func DeleteUfwRule(c fiber.Ctx) error {
	if runtime.GOOS != "linux" {
		return utils.ErrorResponse(c, "UFW is only available on Linux", fiber.StatusBadRequest)
	}

	numStr := c.Params("number")
	num, err := strconv.Atoi(numStr)
	if err != nil || num < 1 {
		return utils.ErrorResponse(c, "Invalid rule number", fiber.StatusBadRequest)
	}

	out, cmdErr := exec.Command("sudo", "ufw", "--force", "delete", fmt.Sprintf("%d", num)).CombinedOutput()
	if cmdErr != nil {
		return utils.ErrorResponse(c, "Failed: "+string(out), fiber.StatusInternalServerError)
	}

	username, _ := c.Locals("username").(string)
	utils.LogAudit(username, "UFW_DELETE_RULE", numStr, string(out))
	logger.Info(fmt.Sprintf("[UFW] Deleted rule #%d by %s", num, username), nil)

	return c.JSON(fiber.Map{
		"success": true,
		"message": strings.TrimSpace(string(out)),
	})
}

// ToggleUfw enables or disables UFW
func ToggleUfw(c fiber.Ctx) error {
	if runtime.GOOS != "linux" {
		return utils.ErrorResponse(c, "UFW is only available on Linux", fiber.StatusBadRequest)
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	action := "disable"
	if req.Enabled {
		action = "enable"
	}

	args := []string{"ufw", "--force", action}
	out, err := exec.Command("sudo", args...).CombinedOutput()
	if err != nil {
		return utils.ErrorResponse(c, "Failed: "+string(out), fiber.StatusInternalServerError)
	}

	username, _ := c.Locals("username").(string)
	utils.LogAudit(username, "UFW_TOGGLE", action, string(out))
	logger.Info(fmt.Sprintf("[UFW] %s by %s", action, username), nil)

	return c.JSON(fiber.Map{
		"success": true,
		"message": strings.TrimSpace(string(out)),
	})
}
