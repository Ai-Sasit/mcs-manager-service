package controllers

import (
	"mc-manage-backend/src/utils"
	"os"

	"github.com/gofiber/fiber/v3"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminUser == "" {
		adminUser = "admin"
	}
	if adminPass == "" {
		adminPass = "admin"
	}

	if req.Username != adminUser || req.Password != adminPass {
		logger.Warn("[Login] Invalid credentials for: "+req.Username, nil)
		return utils.ErrorResponse(c, "Invalid credentials", fiber.StatusUnauthorized)
	}

	token, err := utils.GenerateJWT(req.Username)
	if err != nil {
		logger.Error("[Login] Token generation failed: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Token generation failed", fiber.StatusInternalServerError)
	}

	logger.Info("[Login] User logged in: "+req.Username, nil)
	return c.JSON(fiber.Map{
		"success":  true,
		"token":    token,
		"message":  "Login successful",
		"username": req.Username,
	})
}

func GetMe(c fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	return c.JSON(fiber.Map{
		"success":  true,
		"username": username,
	})
}
