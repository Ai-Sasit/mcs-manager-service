package middleware

import (
	"mc-manage-backend/src/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// tokenValidator is an optional callback to check server-side token revocation.
// Set via SetTokenValidator from routes.go.
var tokenValidator func(username, token string) bool

// SetTokenValidator wires up the server-side token validity check used for logout tracking.
func SetTokenValidator(fn func(username, token string) bool) {
	tokenValidator = fn
}

func AuthRequired(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token format",
		})
	}

	tokenStr := parts[1]
	claims, err := utils.ValidateJWT(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid or expired token",
		})
	}

	if tokenValidator != nil && !tokenValidator(claims.Username, tokenStr) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Token has been revoked",
		})
	}

	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)
	c.Locals("token", tokenStr)
	return c.Next()
}
