package Helpers

import (
	Validator "autobill-service/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

func ValidateRequest(dto any) error {
	validationErrors := Validator.ValidateStruct(dto)
	if validationErrors.HasErrors() {
		return fiber.NewError(fiber.StatusBadRequest, validationErrors.Error())
	}
	return nil
}
