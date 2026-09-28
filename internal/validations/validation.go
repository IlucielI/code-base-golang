package validations

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate executes struct validation rules using ozzo-validation.
func Validate(v validation.Validatable) error {
	if v == nil {
		return nil
	}
	return v.Validate()
}
