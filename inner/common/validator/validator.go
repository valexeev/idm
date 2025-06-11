package validator

import (
	"errors"

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

// Метод валидации
func (v Validator) Validate(request any) error {
	err := v.validate.Struct(request)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			return validateErrs
		}
	}
	return err
}
