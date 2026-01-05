package UserAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	UserDtos "autobill-service/internal/adapters/inbound/http/user/dtos"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service HttpPorts.UserUseCase
}

func CreateUserHandler(service HttpPorts.UserUseCase) UserHandler {
	return UserHandler{service: service}
}

func (h *UserHandler) GetUserHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	idStr, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	result, err := h.service.FindUserById(ctx, idStr)
	if err != nil {
		return err
	}

	return c.JSON(ToUserResponseDto(result))
}

func (h *UserHandler) FindUserByEmailHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	reqBody := new(UserDtos.FindUserByEmailRequestDto)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.FindUserByEmail(ctx, reqBody.Email)
	if err != nil {
		return err
	}

	return c.JSON(ToUserResponseDto(result))
}

func (h *UserHandler) UpdateUserHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	idStr, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	reqBody := new(UserDtos.UpdateUserRequestDto)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.UpdateUser(ctx, idStr, ToUpdateUserInput(reqBody))
	if err != nil {
		return err
	}

	return c.JSON(ToUpdateUserResponseDto(result))
}
