package main

import (
	"fmt"
	routes "mc-manage-backend/src"
	middleware "mc-manage-backend/src/middlewares"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

var logger = utils.NewLogger("mc-manage")

func main() {
	_ = godotenv.Load()
	logger.Init("mc-manage")
	db := utils.ConnectDB()
	defer utils.DisconnectDB(db)

	state := services.NewAppState()

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 100, // 100MB for server jar uploads
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
	app.Use(middleware.Timer)

	routes.RegisterRoutes(app, state)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("MC Manage backend running at http://127.0.0.1:%s\n", port)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			fmt.Printf("Error starting server: %s\n", err.Error())
			shutdown <- syscall.SIGTERM
		}
	}()

	<-shutdown
	fmt.Println("Shutting down server...")

	state.StopAllServers()

	if err := app.Shutdown(); err != nil {
		fmt.Printf("Error during server shutdown: %s\n", err.Error())
		os.Exit(1)
	}
}
