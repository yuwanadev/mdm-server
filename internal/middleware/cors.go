package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns a configured CORS middleware.
func CORS(origins string) fiber.Handler {
	allowedOrigins := "http://localhost:3000"
	if origins != "" {
		allowedOrigins = origins
	}

	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodOptions,
		}, ","),
		AllowHeaders:     "Origin, Content-Type, Authorization",
		AllowCredentials: true,
	})
}
