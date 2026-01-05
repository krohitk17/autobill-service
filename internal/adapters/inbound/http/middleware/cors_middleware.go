package Middlewares

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CORSConfig struct {
	AllowOrigins string

	AllowMethods string

	AllowHeaders string

	ExposeHeaders string

	AllowCredentials bool

	MaxAge int
}

func CORSMiddleware(config CORSConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		allowedOrigin := config.AllowOrigins
		if config.AllowOrigins != "*" && origin != "" {
			allowed := false
			origins := strings.Split(config.AllowOrigins, ",")
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					allowed = true
					allowedOrigin = origin
					break
				}
			}
			if !allowed {
				allowedOrigin = ""
			}
		}

		if allowedOrigin != "" {
			c.Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		if config.AllowCredentials {
			c.Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Method() == fiber.MethodOptions {
			c.Set("Access-Control-Allow-Methods", config.AllowMethods)
			c.Set("Access-Control-Allow-Headers", config.AllowHeaders)
			if config.MaxAge > 0 {
				c.Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}
			return c.SendStatus(fiber.StatusNoContent)
		}

		if config.ExposeHeaders != "" {
			c.Set("Access-Control-Expose-Headers", config.ExposeHeaders)
		}

		return c.Next()
	}
}
