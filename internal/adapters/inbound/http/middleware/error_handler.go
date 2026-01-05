package Middlewares

import (
	Errors "autobill-service/pkg/errors"
	Logger "autobill-service/pkg/logger"
	Validator "autobill-service/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	requestID, _ := c.Locals("requestId").(string)
	Logger.Error().
		Err(err).
		Str("request_id", requestID).
		Str("path", c.Path()).
		Str("method", c.Method()).
		Str("ip", c.IP()).
		Msg("Request error")

	if validationErrs, ok := err.(Validator.ValidationErrors); ok {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Message: "Validation failed",
			Details: validationErrs,
		})
	}

	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(ErrorResponse{
			Message: fiberErr.Message,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Message: Errors.ErrInternal,
	})
}
