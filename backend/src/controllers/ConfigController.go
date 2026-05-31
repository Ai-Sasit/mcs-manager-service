package controllers

import (
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func GetConfig(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	propPath := filepath.Join(srv.ServerDir, "server.properties")
	data, err := os.ReadFile(propPath)
	if err != nil {
		// If file doesn't exist, create a default one and return it
		if os.IsNotExist(err) {
			defaultProps := "server-port=25565\nmax-players=20\nmotd=A Minecraft Server\n"
			os.WriteFile(propPath, []byte(defaultProps), 0644)
			config := parseProperties(defaultProps)
			logger.Info("[GetConfig] Created default config for server="+id, nil)
			return utils.SuccessResponse(c, "OK", config)
		}
		return utils.ErrorResponse(c, "Failed to read config file", fiber.StatusInternalServerError)
	}

	config := parseProperties(string(data))
	logger.Info("[GetConfig] Read config for server="+id, nil)
	return utils.SuccessResponse(c, "OK", config)
}

func UpdateConfig(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	var config map[string]string
	if err := c.Bind().JSON(&config); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	propPath := filepath.Join(srv.ServerDir, "server.properties")
	content := formatProperties(config)
	if err := os.WriteFile(propPath, []byte(content), 0644); err != nil {
		logger.Error("[UpdateConfig] Write error: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Failed to write config", fiber.StatusInternalServerError)
	}

	logger.Info("[UpdateConfig] Updated config for server="+id, nil)
	return utils.SuccessResponse(c, "Config updated", config)
}

func parseProperties(content string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := line[idx+1:]
			result[key] = val
		}
	}
	return result
}

func formatProperties(config map[string]string) string {
	var sb strings.Builder
	sb.WriteString("# Minecraft Server Properties\n")
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(k + "=" + config[k] + "\n")
	}
	return sb.String()
}
