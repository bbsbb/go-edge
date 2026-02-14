package configuration

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New(validator.WithRequiredStructEnabled())

func init() {
	if err := Validate.RegisterValidation("stringenum", validateStringEnum); err != nil {
		panic(fmt.Sprintf("configuration: register stringenum validator: %v", err))
	}
}

func validateStringEnum(fl validator.FieldLevel) bool {
	if v, ok := fl.Field().Interface().(interface{ IsValid() bool }); ok {
		return v.IsValid()
	}
	return false
}
