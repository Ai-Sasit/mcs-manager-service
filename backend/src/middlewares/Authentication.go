package middleware

import (
	"mc-manage-backend/src/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func AuthRequired(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid token format",
		})
	}

	claims, err := utils.ValidateJWT(parts[1])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid or expired token",
		})
	}

	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)
	return c.Next()
}
