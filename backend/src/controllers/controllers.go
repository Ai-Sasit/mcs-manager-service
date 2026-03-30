package controllers

import (
	"fmt"
	"mc-manage-backend/src/interfaces"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"

	"github.com/gofiber/fiber/v3"
)

var state *services.AppState
var logger = utils.NewLogger("mc-manage")

func Init(s *services.AppState) {
	state = s
}

// ListServers returns all servers
func ListServers(c fiber.Ctx) error {
	servers := state.ListServers()
	return c.JSON(interfaces.ApiResponse{
		Success: true,
		Data:    servers,
		Message: "OK",
	})
}

// CreateServer creates a new Minecraft server
func CreateServer(c fiber.Ctx) error {
	var req interfaces.CreateServerRequest
	if err := c.Bind().JSON(&req); err != nil {
		logger.Warn("[CreateServer] Invalid request body: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	if req.Name == "" || req.Edition == "" || req.Version == "" {
		logger.Warn("[CreateServer] Missing required fields", nil)
		return utils.ErrorResponse(c, "name, edition, and version are required", fiber.StatusBadRequest)
	}

	logger.Info(fmt.Sprintf("[CreateServer] name=%s edition=%s version=%s", req.Name, req.Edition, req.Version), nil)

	params := services.CreateServerParams{
		Name:       req.Name,
		Edition:    req.Edition,
		ServerType: req.ServerType,
		Version:    req.Version,
	}
	if req.Port != nil {
		params.Port = *req.Port
	}
	if req.MaxPlayers != nil {
		params.MaxPlayers = *req.MaxPlayers
	}
	if req.MemoryMB != nil {
		params.MemoryMB = *req.MemoryMB
	}

	config, err := services.CreateServer(state, params)
	if err != nil {
		logger.Error("[CreateServer] Failed: "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	logger.Info("[CreateServer] Done id="+config.ID, nil)
	utils.LogAudit("admin", "CREATE_SERVER", req.Name, "Created new server instance.")
	return c.JSON(interfaces.ApiResponse{
		Success: true,
		Data:    config,
		Message: "Server created",
	})
}

// GetServer returns a single server by ID
func GetServer(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Not found", fiber.StatusNotFound)
	}
	return c.JSON(interfaces.ApiResponse{
		Success: true,
		Data:    srv,
		Message: "OK",
	})
}

// DeleteServer deletes a server
func DeleteServer(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[DeleteServer] id="+id, nil)
	if err := services.DeleteServer(state, id); err != nil {
		logger.Error("[DeleteServer] Failed: "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}
	logger.Info("[DeleteServer] Done id="+id, nil)
	utils.LogAudit("admin", "DELETE_SERVER", id, "Deleted server instance.")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Deleted",
	})
}

// StartServer starts a Minecraft server
func StartServer(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[StartServer] id="+id, nil)
	if err := state.StartServer(id); err != nil {
		logger.Error("[StartServer] Failed id="+id+": "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}
	logger.Info("[StartServer] Running id="+id, nil)
	utils.LogAudit("admin", "START_SERVER", id, "Started server instance.")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Started",
	})
}

// StopServer stops a running server
func StopServer(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[StopServer] id="+id, nil)
	if err := state.StopServer(id); err != nil {
		logger.Error("[StopServer] Failed id="+id+": "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}
	logger.Info("[StopServer] Stopped id="+id, nil)
	utils.LogAudit("admin", "STOP_SERVER", id, "Stopped server instance.")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Stopped",
	})
}

// RestartServer restarts a server
func RestartServer(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[RestartServer] id="+id, nil)
	if err := state.RestartServer(id); err != nil {
		logger.Error("[RestartServer] Failed id="+id+": "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}
	logger.Info("[RestartServer] Running id="+id, nil)
	utils.LogAudit("admin", "RESTART_SERVER", id, "Restarted server instance.")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Restarted",
	})
}

// KillServer force stops a server process
func KillServer(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[KillServer] id="+id, nil)
	if err := state.KillServer(id); err != nil {
		logger.Error("[KillServer] Failed id="+id+": "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}
	logger.Info("[KillServer] Killed id="+id, nil)
	utils.LogAudit("admin", "KILL_SERVER", id, "Force killed server process.")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Killed",
	})
}

// ListJavaVersions returns available Java edition versions
func ListJavaVersions(c fiber.Ctx) error {
	logger.Info("[ListJavaVersions] Fetching from Mojang", nil)
	versions, err := services.FetchJavaVersions()
	if err != nil {
		logger.Error("[ListJavaVersions] Failed: "+err.Error(), nil)
		return utils.ErrorResponse(c, fmt.Sprintf("Failed: %s", err.Error()), fiber.StatusInternalServerError)
	}
	logger.Info(fmt.Sprintf("[ListJavaVersions] Got %d versions", len(versions)), nil)
	return c.JSON(interfaces.ApiResponse{
		Success: true,
		Data:    versions,
		Message: "OK",
	})
}

// ListBedrockVersions returns available Bedrock edition versions
func ListBedrockVersions(c fiber.Ctx) error {
	versions := services.GetBedrockVersions()
	return c.JSON(interfaces.ApiResponse{
		Success: true,
		Data:    versions,
		Message: "OK",
	})
}

// ListUsers returns all users
func ListUsers(c fiber.Ctx) error {
	users := state.UserService.ListUsers()
	return c.JSON(interfaces.ApiResponse{Success: true, Data: users, Message: "OK"})
}

// CreateUser creates a new user
func CreateUser(c fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	if role != "admin" {
		return utils.ErrorResponse(c, "Only admins can manage users", fiber.StatusForbidden)
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}
	if req.Username == "" || req.Password == "" {
		return utils.ErrorResponse(c, "Username and password are required", fiber.StatusBadRequest)
	}
	if len(req.Password) < 4 {
		return utils.ErrorResponse(c, "Password must be at least 4 characters", fiber.StatusBadRequest)
	}

	user, err := state.UserService.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit(c.Locals("username").(string), "CREATE_USER", req.Username, "Created new user account.")
	return c.JSON(interfaces.ApiResponse{Success: true, Data: user, Message: "User created"})
}

// UpdateUser updates an existing user
func UpdateUser(c fiber.Ctx) error {
	callerRole, _ := c.Locals("role").(string)
	if callerRole != "admin" {
		return utils.ErrorResponse(c, "Only admins can manage users", fiber.StatusForbidden)
	}

	id := c.Params("id")
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	user, err := state.UserService.UpdateUser(id, req.Username, req.Password, req.Role)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit(c.Locals("username").(string), "UPDATE_USER", user.Username, "Updated user account.")
	return c.JSON(interfaces.ApiResponse{Success: true, Data: user, Message: "User updated"})
}

// DeleteUser deletes a user
func DeleteUser(c fiber.Ctx) error {
	callerRole, _ := c.Locals("role").(string)
	if callerRole != "admin" {
		return utils.ErrorResponse(c, "Only admins can manage users", fiber.StatusForbidden)
	}

	id := c.Params("id")
	if err := state.UserService.DeleteUser(id); err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit(c.Locals("username").(string), "DELETE_USER", id, "Deleted user account.")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "User deleted",
	})
}
