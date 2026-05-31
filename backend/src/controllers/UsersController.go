package controllers

import (
	"mc-manage-backend/src/interfaces"
	"mc-manage-backend/src/utils"

	"github.com/gofiber/fiber/v3"
)

// ListUsers returns all users (admin only).
func ListUsers(c fiber.Ctx) error {
	users := state.UserService.ListUsers()
	return utils.SuccessResponse(c, "OK", users)
}

// CreateUser creates a new user (admin only).
func CreateUser(c fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	if role != "admin" {
		return utils.ErrorResponse(c, "Only admins can manage users", fiber.StatusForbidden)
	}

	var req interfaces.CreateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}
	if req.Username == "" || req.Password == "" {
		return utils.ErrorResponse(c, "Username and password are required", fiber.StatusBadRequest)
	}
	if len(req.Password) < 8 {
		return utils.ErrorResponse(c, "Password must be at least 8 characters", fiber.StatusBadRequest)
	}

	user, err := state.UserService.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit(c.Locals("username").(string), "CREATE_USER", req.Username, "Created new user account.")
	return utils.SuccessResponse(c, "User created", user, fiber.StatusCreated)
}

// UpdateUser updates an existing user (admin only).
func UpdateUser(c fiber.Ctx) error {
	callerRole, _ := c.Locals("role").(string)
	if callerRole != "admin" {
		return utils.ErrorResponse(c, "Only admins can manage users", fiber.StatusForbidden)
	}

	id := c.Params("id")
	var req interfaces.UpdateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	user, err := state.UserService.UpdateUser(id, req.Username, req.Password, req.Role)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit(c.Locals("username").(string), "UPDATE_USER", user.Username, "Updated user account.")
	return utils.SuccessResponse(c, "User updated", user)
}

// DeleteUser deletes a user (admin only).
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
	return utils.SuccessResponse(c, "User deleted", nil)
}
