package controllers

import (
	"mc-manage-backend/src/interfaces"
	"mc-manage-backend/src/utils"

	"github.com/gofiber/fiber/v3"
)

func Login(c fiber.Ctx) error {
	var req interfaces.LoginRequest
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

	if err := state.UserService.AddToken(user.ID, token); err != nil {
		logger.Error("[Login] Failed to persist token for: "+user.Username, nil)
	}

	logger.Info("[Login] User logged in: "+user.Username, nil)
	return utils.SuccessResponse(c, "Login successful", fiber.Map{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
	})
}

func Logout(c fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	token, _ := c.Locals("token").(string)

	user, ok := state.UserService.GetUserByUsername(username)
	if ok && token != "" {
		_ = state.UserService.RemoveToken(user.ID, token)
	}

	logger.Info("[Logout] User logged out: "+username, nil)
	return utils.SuccessResponse(c, "Logged out successfully", nil)
}

func GetMe(c fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)
	return utils.SuccessResponse(c, "OK", fiber.Map{
		"username": username,
		"role":     role,
	})
}

// ChangePassword lets any authenticated user change their own password.
func ChangePassword(c fiber.Ctx) error {
	username, _ := c.Locals("username").(string)

	var req interfaces.ChangePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return utils.ErrorResponse(c, "current_password and new_password are required", fiber.StatusBadRequest)
	}

	user, ok := state.UserService.GetUserByUsername(username)
	if !ok {
		return utils.ErrorResponse(c, "User not found", fiber.StatusNotFound)
	}

	if err := state.UserService.ChangeOwnPassword(user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	logger.Info("[ChangePassword] Password updated for: "+username, nil)
	utils.LogAudit(username, "CHANGE_PASSWORD", username, "User changed their own password.")
	return utils.SuccessResponse(c, "Password updated", nil)
}
