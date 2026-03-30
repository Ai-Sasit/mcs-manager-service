package controllers

import (
	"encoding/json"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofiber/fiber/v3"
)

type AppSettings struct {
	AppName         string `json:"app_name"`
	DefaultRAM      int    `json:"default_ram"`
	DiscordWebhook  string `json:"discord_webhook"`
	TelemetryEnabled bool   `json:"telemetry_enabled"`
}

var (
	settings   AppSettings
	settingsMu sync.RWMutex
)

func init() {
	// Root of data/ is where we store global settings
	path := filepath.Join("data", "settings.json")
	if b, err := os.ReadFile(path); err == nil {
		json.Unmarshal(b, &settings)
	} else {
		// Defaults
		settings = AppSettings{
			AppName:    "MC Management Dashboard",
			DefaultRAM: 2048,
		}
	}
}

func GetSettings(c fiber.Ctx) error {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	return c.JSON(fiber.Map{
		"success": true,
		"data":    settings,
	})
}

func UpdateSettings(c fiber.Ctx) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	var newSettings AppSettings
	if err := c.Bind().JSON(&newSettings); err != nil {
		return utils.ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	settings = newSettings
	path := filepath.Join("data", "settings.json")
	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(path, data, 0644)

	utils.LogAudit("admin", "SETTINGS_UPDATE", "system", "Updated global application settings.")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Settings updated",
	})
}
