package Middlewares

import (
	Helpers "autobill-service/pkg/helpers"
	Logger "autobill-service/pkg/logger"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

func RequestContextMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		startTime := time.Now()

		c.Locals("requestId", requestID)
		c.Locals("startTime", startTime)

		c.Set("X-Request-ID", requestID)

		Logger.Info().
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Msg("Request started")

		err := c.Next()

		duration := time.Since(startTime)
		Logger.Info().
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Msg("Request completed")

		return err
	}
}

func GetContext(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()

	if requestID, ok := c.Locals("requestId").(string); ok {
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
	}

	if userID, ok := c.Locals(string(Helpers.LoggedInUserIDKey)).(string); ok {
		ctx = context.WithValue(ctx, UserIDKey, userID)
	}

	return ctx
}
