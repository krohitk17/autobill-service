package SettlementAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	SettlementDtos "autobill-service/internal/adapters/inbound/http/settlement/dtos"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type SettlementHandler struct {
	service HttpPorts.SettlementUseCase
}

func CreateSettlementHandler(service HttpPorts.SettlementUseCase) SettlementHandler {
	return SettlementHandler{service: service}
}

func (h *SettlementHandler) CreateSettlementHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	reqBody := new(SettlementDtos.CreateSettlementRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.CreateSettlement(ctx, userId, ToCreateSettlementInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(ToSettlementResponseDto(result))
}

func (h *SettlementHandler) GetPendingSettlementsHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetPendingSettlements(ctx, userId, pagination)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToSettlementListResponseDto(result))
}

func (h *SettlementHandler) GetSettlementHistoryHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetSettlementHistory(ctx, userId, pagination)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToSettlementListResponseDto(result))
}

func (h *SettlementHandler) ConfirmSettlementHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	settlementId, err := Helpers.ParseUUID(c.Params("settlementId"))
	if err != nil {
		return err
	}

	err = h.service.ConfirmSettlement(ctx, userId, settlementId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SettlementHandler) DeleteSettlementHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	settlementId, err := Helpers.ParseUUID(c.Params("settlementId"))
	if err != nil {
		return err
	}

	err = h.service.DeleteSettlement(ctx, userId, settlementId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
