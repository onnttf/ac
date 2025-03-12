package validator

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the go-playground validator to provide custom validation.
type CustomValidator struct {
	validate *validator.Validate
}

// NewCustomValidator creates a new instance of CustomValidator.
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validate: validator.New(),
	}
}

// Validate validates the given struct based on the tags defined.
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validate.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		// return types.ErrInvalidInput
		return err
	}
	return nil
}
