package Middlewares

import (
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"
	JWTUtil "autobill-service/pkg/jwt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(util JWTUtil.JWTUtil) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtToken := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if jwtToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, Errors.ErrNoToken)
		}
		parsedToken, isValid := util.Parse(jwtToken)
		if !isValid {
			return fiber.NewError(fiber.StatusUnauthorized, Errors.ErrInvalidToken)
		}
		c.Locals(string(Helpers.LoggedInUserIDKey), parsedToken)
		return c.Next()
	}
}
