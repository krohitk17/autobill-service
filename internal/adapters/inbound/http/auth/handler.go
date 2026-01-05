package AuthAdapter

import (
	Dtos "autobill-service/internal/adapters/inbound/http/auth/dtos"
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service HttpPorts.AuthUseCase
}

func CreateAuthHandler(service HttpPorts.AuthUseCase) AuthHandler {
	return AuthHandler{service: service}
}

func (h *AuthHandler) RegisterUserHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	reqBody := new(Dtos.RegisterUserRequestDto)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.RegisterUser(ctx, ToRegisterUserInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(ToUserLoginResponseDto(result))
}

func (h *AuthHandler) LoginHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	reqBody := new(Dtos.FindUserRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.AuthenticateUser(ctx, ToLoginInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToUserLoginResponseDto(result))
}

func (h *AuthHandler) UpdatePasswordHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	reqBody := new(Dtos.UpdatePasswordRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	err = h.service.UpdatePassword(ctx, userId, reqBody.OldPassword, reqBody.NewPassword)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) DeactivateUserHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	reqBody := new(Dtos.DeactivateUserRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	err = h.service.DeactivateUser(ctx, userId, reqBody.Password)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) ReactivateUserHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	reqBody := new(Dtos.ReactivateUserRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	err := h.service.ReactivateUser(ctx, reqBody.Email, reqBody.Password)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) RefreshTokenHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	reqBody := new(Dtos.RefreshTokenRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.RefreshToken(ctx, ToRefreshTokenInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToUserLoginResponseDto(result))
}

func (h *AuthHandler) LogoutHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	reqBody := new(Dtos.LogoutRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	err := h.service.Logout(ctx, reqBody.RefreshToken)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) LogoutAllHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	err = h.service.LogoutAll(ctx, userId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
