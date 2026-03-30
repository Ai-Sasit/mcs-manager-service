package main

import (
	"bufio"
	"fmt"
	routes "mc-manage-backend/src"
	middleware "mc-manage-backend/src/middlewares"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

var logger = utils.NewLogger("mc-manage")

// loadEnvFile reads a .env file and sets env vars that are not already set
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	loadEnvFile(".env")
	logger.Init("mc-manage")
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
