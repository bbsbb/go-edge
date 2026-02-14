package otelfx

import (
	"github.com/go-playground/validator/v10"

	"github.com/bbsbb/go-edge/core/configuration"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var _ configuration.WithValidation = (*Configuration)(nil)

// WithOTel is implemented by application configurations that provide OpenTelemetry settings.
type WithOTel interface {
	OTelConfiguration() *Configuration
}

// Configuration defines OpenTelemetry SDK settings.
// When Enabled is false, no TracerProvider or MeterProvider is initialized
// and the OTel globals remain no-op (zero overhead).
type Configuration struct {
	Enabled     bool    `yaml:"enabled" env:"ENABLED,overwrite"`
	Endpoint    string  `yaml:"endpoint" env:"ENDPOINT,overwrite" validate:"required_if=Enabled true"`
	ServiceName string  `yaml:"service_name" env:"SERVICE_NAME,overwrite" validate:"required"`
	SampleRate  float64 `yaml:"sample_rate" env:"SAMPLE_RATE,overwrite" validate:"gte=0,lte=1"`
	Insecure    bool    `yaml:"insecure" env:"INSECURE,overwrite"`
}

func (c *Configuration) Validate() error {
	return validate.Struct(c)
}
