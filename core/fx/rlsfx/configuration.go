package rlsfx

import (
	"github.com/go-playground/validator/v10"

	"github.com/bbsbb/go-edge/core/configuration"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var _ configuration.WithValidation = (*Configuration)(nil)

// WithRLS is implemented by application configurations that provide RLS settings.
type WithRLS interface {
	RLSConfiguration() *Configuration
}

// Configuration defines the PostgreSQL session variable used for row-level security.
type Configuration struct {
	Schema string `yaml:"schema" env:"SCHEMA,overwrite" validate:"required"`
	Field  string `yaml:"field" env:"FIELD,overwrite" validate:"required"`
}

func (c *Configuration) Validate() error {
	return validate.Struct(c)
}
