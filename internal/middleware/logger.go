package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger logs incoming requests with method, path, status, and duration.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		log.Printf("[%s] %s %s %d %s",
			c.Method(),
			c.Path(),
			c.IP(),
			c.Response().StatusCode(),
			duration,
		)

		return err
	}
}
