package Validator

import (
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
	})
	return validate
}

func ValidateStruct(s interface{}) ValidationErrors {
	err := GetValidator().Struct(s)
	if err == nil {
		return nil
	}

	var errors ValidationErrors
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Value:   err.Value(),
			Message: formatValidationError(err),
		})
	}
	return errors
}

type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var messages []string
	for _, e := range ve {
		messages = append(messages, e.Message)
	}
	return strings.Join(messages, "; ")
}

func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "uuid":
		return fe.Field() + " must be a valid UUID"
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	case "gt":
		return fe.Field() + " must be greater than " + fe.Param()
	case "gte":
		return fe.Field() + " must be greater than or equal to " + fe.Param()
	case "lt":
		return fe.Field() + " must be less than " + fe.Param()
	case "lte":
		return fe.Field() + " must be less than or equal to " + fe.Param()
	default:
		return fe.Field() + " failed validation: " + fe.Tag()
	}
}
