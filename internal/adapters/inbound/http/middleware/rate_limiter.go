package Middlewares

import (
	"time"

	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
	Message    string
}

func NewRateLimiter(config RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.Max,
		Expiration: config.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID, ok := c.Locals(string(Helpers.LoggedInUserIDKey)).(string)
			if ok && userID != "" {
				return userID
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
				Message: config.Message,
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}
