package response

import "github.com/gofiber/fiber/v2"

// JSON sends a successful JSON response.
func JSON(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// Error sends an error JSON response.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

// Created sends a 201 Created response.
func Created(c *fiber.Ctx, data interface{}) error {
	return JSON(c, fiber.StatusCreated, data)
}

// OK sends a 200 OK response.
func OK(c *fiber.Ctx, data interface{}) error {
	return JSON(c, fiber.StatusOK, data)
}

// NoContent sends a 204 No Content response.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// BadRequest sends a 400 Bad Request error.
func BadRequest(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized error.
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, message)
}

// NotFound sends a 404 Not Found error.
func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, message)
}

// InternalError sends a 500 Internal Server Error.
func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, message)
}
