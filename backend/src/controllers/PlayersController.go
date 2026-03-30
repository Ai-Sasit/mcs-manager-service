package controllers

import (
	"encoding/json"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

type PlayerInfo struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Role string `json:"role"` // "admin" or "member"
}

// GetPlayers reads the whitelist and ops lists from server files
func GetPlayers(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	playersMap := make(map[string]PlayerInfo)

	if srv.Edition == models.EditionJava {
		// Read ops.json
		opsPath := filepath.Join(srv.ServerDir, "ops.json")
		if b, err := os.ReadFile(opsPath); err == nil {
			var ops []struct {
				UUID string `json:"uuid"`
				Name string `json:"name"`
			}
			json.Unmarshal(b, &ops)
			for _, op := range ops {
				playersMap[op.Name] = PlayerInfo{UUID: op.UUID, Name: op.Name, Role: "admin"}
			}
		}
		// Read whitelist.json
		wlPath := filepath.Join(srv.ServerDir, "whitelist.json")
		if b, err := os.ReadFile(wlPath); err == nil {
			var wls []struct {
				UUID string `json:"uuid"`
				Name string `json:"name"`
			}
			json.Unmarshal(b, &wls)
			for _, wl := range wls {
				if _, exists := playersMap[wl.Name]; !exists {
					playersMap[wl.Name] = PlayerInfo{UUID: wl.UUID, Name: wl.Name, Role: "member"}
				}
			}
		}
	} else if srv.Edition == models.EditionBedrock {
		// Read permissions.json
		permPath := filepath.Join(srv.ServerDir, "permissions.json")
		if b, err := os.ReadFile(permPath); err == nil {
			var perms []map[string]interface{}
			json.Unmarshal(b, &perms)
			// Simple fallback if no strong typing
		}
		// Read allowlist.json
		alPath := filepath.Join(srv.ServerDir, "allowlist.json")
		if b, err := os.ReadFile(alPath); err == nil {
			var als []struct {
				Name string `json:"name"`
			}
			json.Unmarshal(b, &als)
			for _, al := range als {
				playersMap[al.Name] = PlayerInfo{Name: al.Name, Role: "member"}
			}
		}
	}

	var result []PlayerInfo
	for _, p := range playersMap {
		result = append(result, p)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

type UpdatePlayerReq struct {
	Name   string `json:"name"`
	Action string `json:"action"` // "add_whitelist", "remove_whitelist", "add_op", "remove_op"
}

// UpdatePlayer sends a command to the server (must be running)
func UpdatePlayer(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	if srv.Status != models.StatusRunning {
		return utils.ErrorResponse(c, "Server must be running to update players", fiber.StatusBadRequest)
	}

	var req UpdatePlayerReq
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	cmdStr := ""
	if srv.Edition == models.EditionJava {
		switch req.Action {
		case "add_whitelist":
			cmdStr = "whitelist add " + req.Name
		case "remove_whitelist":
			cmdStr = "whitelist remove " + req.Name
		case "add_op":
			cmdStr = "op " + req.Name
		case "remove_op":
			cmdStr = "deop " + req.Name
		}
	} else {
		switch req.Action {
		case "add_whitelist":
			cmdStr = "allowlist add " + req.Name
		case "remove_whitelist":
			cmdStr = "allowlist remove " + req.Name
		case "add_op":
			cmdStr = "op " + req.Name
		case "remove_op":
			cmdStr = "deop " + req.Name
		}
	}

	if cmdStr != "" {
		if err := state.SendCommand(id, cmdStr); err != nil {
			return utils.ErrorResponse(c, "Failed to execute command: "+err.Error(), fiber.StatusInternalServerError)
		}
		utils.LogAudit("admin", "PLAYER_UPDATE", id, "Executed: "+cmdStr)
	}

	return c.JSON(fiber.Map{"success": true})
}
