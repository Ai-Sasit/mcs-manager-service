package src

import (
	"mc-manage-backend/src/controllers"
	middleware "mc-manage-backend/src/middlewares"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, state *services.AppState) {
	controllers.Init(state)

	// Public auth
	app.Post("/api/auth/login", controllers.Login)

	// WebSocket auth middleware (checks ?token= query param)
	app.Use("/ws", func(c fiber.Ctx) error {
		token := c.Query("token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).SendString("Missing token")
		}
		if _, err := utils.ValidateJWT(token); err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}
		return c.Next()
	})

	// WebSocket routes (plain Fiber v3 handlers, upgrade handled inside)
	app.Get("/ws/servers/:id/logs", controllers.WsLogs)
	app.Get("/ws/servers/:id/terminal", controllers.WsTerminal)

	// Protected API routes
	api := app.Group("/api", middleware.AuthRequired)

	api.Get("/auth/me", controllers.GetMe)

	// Server CRUD
	api.Get("/servers", controllers.ListServers)
	api.Post("/servers", controllers.CreateServer)
	api.Get("/servers/:id", controllers.GetServer)
	api.Delete("/servers/:id", controllers.DeleteServer)

	// Server actions
	api.Post("/servers/:id/start", controllers.StartServer)
	api.Post("/servers/:id/stop", controllers.StopServer)
	api.Post("/servers/:id/restart", controllers.RestartServer)

	// Config
	api.Get("/servers/:id/config", controllers.GetConfig)
	api.Put("/servers/:id/config", controllers.UpdateConfig)

	// Plugins
	api.Get("/servers/:id/plugins", controllers.ListPlugins)
	api.Post("/servers/:id/plugins", controllers.UploadPlugin)
	api.Delete("/servers/:id/plugins/:name", controllers.DeletePlugin)

	// Versions
	api.Get("/versions/java", controllers.ListJavaVersions)
	api.Get("/versions/bedrock", controllers.ListBedrockVersions)

	// System
	api.Get("/system/info", controllers.GetSystemInfo)
	api.Get("/audit-logs", controllers.GetAuditLogs)

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
	api.Get("/settings", controllers.GetSettings)
	api.Put("/settings", controllers.UpdateSettings)
}
