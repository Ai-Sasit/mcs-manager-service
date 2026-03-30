package controllers

import (
	"mc-manage-backend/src/utils"

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

	user, err := state.UserService.Authenticate(req.Username, req.Password)
	if err != nil {
		logger.Warn("[Login] Invalid credentials for: "+req.Username, nil)
		return utils.ErrorResponse(c, "Invalid credentials", fiber.StatusUnauthorized)
	}

	token, err := utils.GenerateJWT(user.Username, user.Role)
	if err != nil {
		logger.Error("[Login] Token generation failed: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Token generation failed", fiber.StatusInternalServerError)
	}

	logger.Info("[Login] User logged in: "+user.Username, nil)
	return c.JSON(fiber.Map{
		"success":  true,
		"token":    token,
		"message":  "Login successful",
		"username": user.Username,
		"role":     user.Role,
	})
}

func GetMe(c fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)
	return c.JSON(fiber.Map{
		"success":  true,
		"username": username,
		"role":     role,
	})
}
