package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mc-manage-backend/src/interfaces"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"time"

	"github.com/gofiber/fiber/v3"
)

var state *services.AppState

func Init(s *services.AppState) {
	state = s
}

// ListServers returns all servers
func ListServers(c fiber.Ctx) error {
	servers := state.ListServers()
	return utils.SuccessResponse(c, "OK", servers)
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
	if !isValidServerEdition(req.Edition) {
		return utils.ErrorResponse(c, "edition must be java or bedrock", fiber.StatusBadRequest)
	}

	logger.Info(fmt.Sprintf("[CreateServer] name=%s edition=%s version=%s", req.Name, req.Edition, req.Version), nil)

	params := createServerParamsFromRequest(req)

	config, err := services.CreateServer(state, params)
	if err != nil {
		logger.Error("[CreateServer] Failed: "+err.Error(), nil)
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	logger.Info("[CreateServer] Done id="+config.ID, nil)
	utils.LogAudit("admin", "CREATE_SERVER", req.Name, "Created new server instance.")
	return utils.SuccessResponse(c, "Server created", config, fiber.StatusCreated)
}

func CreateServerSetupJob(c fiber.Ctx) error {
	var req interfaces.CreateServerRequest
	if err := c.Bind().JSON(&req); err != nil {
		logger.Warn("[CreateServerSetupJob] Invalid request body: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	if req.Name == "" || req.Edition == "" || req.Version == "" {
		logger.Warn("[CreateServerSetupJob] Missing required fields", nil)
		return utils.ErrorResponse(c, "name, edition, and version are required", fiber.StatusBadRequest)
	}
	if !isValidServerEdition(req.Edition) {
		return utils.ErrorResponse(c, "edition must be java or bedrock", fiber.StatusBadRequest)
	}

	params := createServerParamsFromRequest(req)
	job := state.SetupJobs().Create()
	logger.Info(fmt.Sprintf("[CreateServerSetupJob] job=%s name=%s edition=%s version=%s", job.ID, req.Name, req.Edition, req.Version), nil)

	go func() {
		config, err := services.CreateServerWithProgress(state, params, job.Publish)
		if err != nil {
			job.Publish(services.SetupEvent{
				Type:    "setup",
				Step:    "failed",
				Status:  "failed",
				Message: err.Error(),
				Percent: 100,
			})
			state.SetupJobs().DeleteAfter(job.ID, 15*time.Minute)
			return
		}
		utils.LogAudit("admin", "CREATE_SERVER", req.Name, "Created new server instance.")
		logger.Info("[CreateServerSetupJob] Done job="+job.ID+" id="+config.ID, nil)
		state.SetupJobs().DeleteAfter(job.ID, 15*time.Minute)
	}()

	return utils.SuccessResponse(c, "Server setup started", fiber.Map{"job_id": job.ID}, fiber.StatusCreated)
}

// CreateServerStream creates a new server and streams progress via SSE (Server-Sent Events).
// This replaces the WebSocket-based /servers/setup-jobs + /ws/server-setup/:job_id flow.
func CreateServerStream(c fiber.Ctx) error {
	var req interfaces.CreateServerRequest
	if err := c.Bind().JSON(&req); err != nil {
		logger.Warn("[CreateServerStream] Invalid request body: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	if req.Name == "" || req.Edition == "" || req.Version == "" {
		logger.Warn("[CreateServerStream] Missing required fields", nil)
		return utils.ErrorResponse(c, "name, edition, and version are required", fiber.StatusBadRequest)
	}
	if !isValidServerEdition(req.Edition) {
		return utils.ErrorResponse(c, "edition must be java or bedrock", fiber.StatusBadRequest)
	}

	params := createServerParamsFromRequest(req)
	logger.Info(fmt.Sprintf("[CreateServerStream] name=%s edition=%s version=%s", req.Name, req.Edition, req.Version), nil)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
		flush := func() { w.Flush() }

		writeSSE := func(event services.SetupEvent) {
			data, err := json.Marshal(event)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flush()
		}

		config, err := services.CreateServerWithProgress(state, params, writeSSE)
		if err != nil {
			writeSSE(services.SetupEvent{
				Type:    "setup",
				Step:    "failed",
				Status:  "failed",
				Message: err.Error(),
				Percent: 100,
			})
			return
		}
		utils.LogAudit("admin", "CREATE_SERVER", req.Name, "Created new server instance.")
		logger.Info("[CreateServerStream] Done id="+config.ID, nil)
	})

	return nil
}

func createServerParamsFromRequest(req interfaces.CreateServerRequest) services.CreateServerParams {
	params := services.CreateServerParams{
		Name:       req.Name,
		Edition:    req.Edition,
		ServerType: req.ServerType,
		ModLoader:  req.ModLoader,
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
	return params
}

func isValidServerEdition(edition models.ServerEdition) bool {
	return edition == models.EditionJava || edition == models.EditionBedrock
}

// GetServer returns a single server by ID
func GetServer(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Not found", fiber.StatusNotFound)
	}
	return utils.SuccessResponse(c, "OK", srv)
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
	return utils.SuccessResponse(c, "Deleted", nil)
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
	return utils.SuccessResponse(c, "Started", nil)
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
	return utils.SuccessResponse(c, "Stopped", nil)
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
	return utils.SuccessResponse(c, "Restarted", nil)
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
	return utils.SuccessResponse(c, "Killed", nil)
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
	return utils.SuccessResponse(c, "OK", versions)
}

// ListBedrockVersions returns available Bedrock edition versions
func ListBedrockVersions(c fiber.Ctx) error {
	versions := services.GetBedrockVersions()
	return utils.SuccessResponse(c, "OK", versions)
}
