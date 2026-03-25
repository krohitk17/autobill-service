package Middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

type TimeoutConfig struct {
	Timeout time.Duration
	Message string
}

func TimeoutMiddleware(config TimeoutConfig) fiber.Handler {
	return timeout.NewWithContext(func(c *fiber.Ctx) error {
		return c.Next()
	}, config.Timeout)
}
