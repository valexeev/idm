package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

// Конструктор
func New() *Validator {
	validate := validator.New()
	return &Validator{validate: validate}
}

// Метод валидации - теперь возвращает оригинальные ValidationErrors для совместимости с тестами
func (v Validator) Validate(request any) error {
	err := v.validate.Struct(request)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			// Возвращаем оригинальные ValidationErrors для совместимости с тестами
			return validateErrs
		}
	}
	return err
}

// ValidateWithCustomMessages - новый метод для получения кастомных сообщений
func (v Validator) ValidateWithCustomMessages(request any) error {
	err := v.validate.Struct(request)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			return v.customValidationError(validateErrs)
		}
	}
	return err
}

// Создает кастомное сообщение об ошибке валидации
func (v Validator) customValidationError(errs validator.ValidationErrors) error {
	var messages []string

	for _, err := range errs {
		switch err.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s cannot be empty", strings.ToLower(err.Field())))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters long", strings.ToLower(err.Field()), err.Param()))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must be at most %s characters long", strings.ToLower(err.Field()), err.Param()))
		case "gt":
			messages = append(messages, fmt.Sprintf("%s must be greater than %s", strings.ToLower(err.Field()), err.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s validation failed on %s", strings.ToLower(err.Field()), err.Tag()))
		}
	}

	return errors.New(strings.Join(messages, ", "))
}
