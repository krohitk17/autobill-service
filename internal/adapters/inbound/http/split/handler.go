package SplitAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	SplitDtos "autobill-service/internal/adapters/inbound/http/split/dtos"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type SplitHandler struct {
	service HttpPorts.SplitUseCase
}

func CreateSplitHandler(service HttpPorts.SplitUseCase) SplitHandler {
	return SplitHandler{service: service}
}

func (h *SplitHandler) CreateSplitHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	reqBody := new(SplitDtos.CreateSplitRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.CreateSplit(ctx, userId, ToCreateSplitInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(ToSplitResponseDto(result))
}

func (h *SplitHandler) GetSplitHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	splitId, err := Helpers.ParseUUID(c.Params("splitId"))
	if err != nil {
		return err
	}

	result, err := h.service.GetSplit(ctx, userId, splitId)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(ToSplitResponseDto(result))
}

func (h *SplitHandler) GetGroupSplitsHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetGroupSplits(ctx, userId, groupId, pagination)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToSplitListResponseDto(result))
}

func (h *SplitHandler) GetMySplitsHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetMySplits(ctx, userId, pagination)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToSplitListResponseDto(result))
}

func (h *SplitHandler) ReverseSplitHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	splitId, err := Helpers.ParseUUID(c.Params("splitId"))
	if err != nil {
		return err
	}

	result, err := h.service.ReverseSplit(ctx, userId, splitId)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(ToSplitResponseDto(result))
}
