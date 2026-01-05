package BalanceAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type BalanceHandler struct {
	service HttpPorts.BalanceUseCase
}

func CreateBalanceHandler(service HttpPorts.BalanceUseCase) BalanceHandler {
	return BalanceHandler{service: service}
}

func (h *BalanceHandler) GetMyBalanceHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	result, err := h.service.GetMyBalance(ctx, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToUserBalanceResponseDto(result))
}

func (h *BalanceHandler) GetUserBalanceHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	requesterId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	targetUserId, err := Helpers.ParseUUID(c.Params("userId"))
	if err != nil {
		return err
	}

	if requesterId != targetUserId {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrBalanceNotFound)
	}

	result, err := h.service.GetMyBalance(ctx, requesterId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToUserBalanceResponseDto(result))
}

func (h *BalanceHandler) GetGroupBalanceHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	result, err := h.service.GetGroupBalance(ctx, userId, groupId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToGroupBalanceResponseDto(result))
}

func (h *BalanceHandler) RecalculateGroupBalanceHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	result, err := h.service.RecalculateGroupBalance(ctx, userId, groupId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToGroupBalanceResponseDto(result))
}

func (h *BalanceHandler) GetSimplifiedDebtsHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	result, err := h.service.GetSimplifiedDebts(ctx, userId, groupId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToSimplifiedDebtsResponseDto(result))
}
