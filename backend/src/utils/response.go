package utils

import "github.com/gofiber/fiber/v3"

func SuccessResponse(c fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func ErrorResponse(c fiber.Ctx, message string, statusCode int) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}
