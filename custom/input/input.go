package input

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// BindAndValidate binds and validates the input from the echo context.
func BindAndValidate(c echo.Context, input any) error {
	if err := c.Bind(input); err != nil {
		return fmt.Errorf("failed to bind input, err: %w", err)
	}
	if err := c.Validate(input); err != nil {
		return fmt.Errorf("failed to validate input, err: %w", err)
	}
	return nil
}
