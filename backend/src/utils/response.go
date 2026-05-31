package utils

import "github.com/gofiber/fiber/v3"

func SuccessResponse(c fiber.Ctx, message string, data interface{}, statusCode ...int) error {
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	return c.Status(code).JSON(fiber.Map{
		"status":  "ok",
		"message": message,
		"data":    data,
	})
}

func ErrorResponse(c fiber.Ctx, message string, statusCode int) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

func PaginatedResponse(c fiber.Ctx, message string, data interface{}, total, page, limit int) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": message,
		"data":    data,
		"pagination": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
