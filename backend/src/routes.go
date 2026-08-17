package src

import (
	"mc-manage-backend/src/controllers"
	middleware "mc-manage-backend/src/middlewares"
	"mc-manage-backend/src/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, state *services.AppState) {
	controllers.Init(state)
	middleware.SetTokenValidator(state.UserService.IsTokenValid)

	// Public auth
	app.Post("/api/v1/auth/login", controllers.Login)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("V.1.0.0")
	})

	// WebSocket routes authenticate during upgrade so clients receive an
	// application close code instead of an opaque failed handshake.
	app.Get("/ws/servers/:id/logs", controllers.WsLogs)
	app.Get("/ws/servers/:id/terminal", controllers.WsTerminal)
	app.Get("/ws/server-setup/:job_id", controllers.WsServerSetup)
	app.Get("/ws/backend-logs", controllers.WsBackendLogs)
	app.Get("/ws/system/resources", controllers.WsSystemResources)

	// Protected API routes
	api := app.Group("/api/v1", middleware.AuthRequired)

	api.Get("/auth/me", controllers.GetMe)
	api.Post("/auth/logout", controllers.Logout)
	api.Put("/auth/me/password", controllers.ChangePassword)

	// Server CRUD
	api.Get("/servers", controllers.ListServers)
	api.Post("/servers/setup-jobs", controllers.CreateServerSetupJob)
	api.Post("/servers/create-stream", controllers.CreateServerStream)
	api.Post("/servers", controllers.CreateServer)
	api.Get("/servers/:id", controllers.GetServer)
	api.Delete("/servers/:id", controllers.DeleteServer)

	// Server actions
	api.Post("/servers/:id/start", controllers.StartServer)
	api.Post("/servers/:id/stop", controllers.StopServer)
	api.Post("/servers/:id/restart", controllers.RestartServer)
	api.Post("/servers/:id/kill", controllers.KillServer)

	// Config
	api.Get("/servers/:id/config", controllers.GetConfig)
	api.Put("/servers/:id/config", controllers.UpdateConfig)

	// Plugins
	api.Get("/servers/:id/plugins", controllers.ListPlugins)
	api.Post("/servers/:id/plugins", controllers.UploadPlugin)
	api.Delete("/servers/:id/plugins/:name", controllers.DeletePlugin)

	// Mods (Forge/Fabric)
	api.Get("/servers/:id/mods", controllers.ListMods)
	api.Post("/servers/:id/mods", controllers.UploadMod)
	api.Delete("/servers/:id/mods/:name", controllers.DeleteMod)

	// Versions
	api.Get("/versions/java", controllers.ListJavaVersions)
	api.Get("/versions/bedrock", controllers.ListBedrockVersions)

	// System
	api.Get("/system/info", controllers.GetSystemInfo)
	api.Get("/system/port-lookup", controllers.LookupPort)
	api.Post("/system/kill-pid", controllers.KillPid)
	api.Get("/audit-logs", controllers.GetAuditLogs)
	api.Get("/backend-logs", controllers.ListBackendLogFiles)
	api.Get("/backend-logs/file", controllers.GetBackendLogFile)

	// Additional Stubs
	api.Get("/servers/:id/players", controllers.GetPlayers)
	api.Post("/servers/:id/players", controllers.UpdatePlayer)

	api.Get("/servers/:id/files", controllers.ListFiles)
	api.Get("/servers/:id/files/read", controllers.ReadFile)
	api.Post("/servers/:id/files/write", controllers.WriteFile)

	api.Get("/servers/:id/backups", controllers.ListBackups)
	api.Post("/servers/:id/backups", controllers.CreateBackup)
	api.Delete("/servers/:id/backups/:name", controllers.DeleteBackup)

	api.Get("/schedules", controllers.ListSchedules)
	api.Post("/schedules", controllers.CreateSchedule)
	api.Put("/schedules/:id/toggle", controllers.ToggleSchedule)
	api.Delete("/schedules/:id", controllers.DeleteSchedule)

	api.Get("/users", controllers.ListUsers)
	api.Post("/users", controllers.CreateUser)
	api.Put("/users/:id", controllers.UpdateUser)
	api.Delete("/users/:id", controllers.DeleteUser)
	api.Get("/settings", controllers.GetSettings)
	api.Put("/settings", controllers.UpdateSettings)

}
