package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yuwanadev/mdm-backend/internal/service"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

// JWTAuth creates a middleware that validates JWT tokens from the Authorization header.
func JWTAuth(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Unauthorized(c, "invalid authorization format")
		}

		userID, err := authService.ValidateToken(parts[1])
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}

		// Store user ID in context for downstream handlers
		c.Locals("userID", userID)
		return c.Next()
	}
}
