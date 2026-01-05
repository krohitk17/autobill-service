package Helpers

import (
	Errors "autobill-service/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContextKey string

const (
	LoggedInUserIDKey ContextKey = "loggedInUserId"
)

func GetUserIdFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIdLocal := c.Locals(string(LoggedInUserIDKey))
	if userIdLocal == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, Errors.ErrNoToken)
	}
	id, err := uuid.Parse(userIdLocal.(string))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidToken)
	}
	return id, nil
}

func ParseUUID(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidId)
	}
	return id, nil
}
