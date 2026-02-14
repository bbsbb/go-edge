package middlewarefx

import (
	"github.com/go-playground/validator/v10"

	"github.com/bbsbb/go-edge/core/configuration"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var _ configuration.WithValidation = (*Configuration)(nil)

// WithMiddleware is implemented by application configurations that provide middleware settings.
type WithMiddleware interface {
	MiddlewareConfiguration() *Configuration
}

type RecoveryConfig struct {
	Enabled bool `yaml:"enabled" env:"ENABLE_RECOVERY,overwrite"`
}

type MaxBytesConfig struct {
	Enabled  bool  `yaml:"enabled" env:"ENABLE_MAX_BYTES,overwrite"`
	MaxBytes int64 `yaml:"max_bytes" env:"MAX_REQUEST_BODY_BYTES,overwrite" validate:"gte=0"`
}

type RequestIDConfig struct {
	Enabled bool `yaml:"enabled" env:"ENABLE_REQUEST_ID,overwrite"`
}

type CorrelationIDConfig struct {
	Enabled bool   `yaml:"enabled" env:"ENABLE_CORRELATION_ID,overwrite"`
	Header  string `yaml:"header" env:"CORRELATION_ID_HEADER,overwrite"`
}

type OTelHTTPConfig struct {
	Enabled bool `yaml:"enabled" env:"ENABLE_OTEL_HTTP,overwrite"`
}

type RequestLogConfig struct {
	Enabled bool `yaml:"enabled" env:"ENABLE_REQUEST_LOGGING,overwrite"`
}

// Configuration controls which middlewares are active in the stack.
// Use DefaultConfiguration() to get a Configuration with all middlewares enabled.
type Configuration struct {
	Recovery      RecoveryConfig      `yaml:"recovery"`
	MaxBytes      MaxBytesConfig      `yaml:"max_bytes"`
	RequestID     RequestIDConfig     `yaml:"request_id"`
	CorrelationID CorrelationIDConfig `yaml:"correlation_id"`
	OTelHTTP      OTelHTTPConfig      `yaml:"otel_http"`
	RequestLog    RequestLogConfig    `yaml:"request_log"`
}

// DefaultConfiguration returns a Configuration with all middlewares enabled.
// Callers can then selectively disable individual middlewares.
func DefaultConfiguration() Configuration {
	return Configuration{
		Recovery:      RecoveryConfig{Enabled: true},
		MaxBytes:      MaxBytesConfig{Enabled: true},
		RequestID:     RequestIDConfig{Enabled: true},
		CorrelationID: CorrelationIDConfig{Enabled: true},
		OTelHTTP:      OTelHTTPConfig{Enabled: true},
		RequestLog:    RequestLogConfig{Enabled: true},
	}
}

func (c *Configuration) Validate() error {
	return validate.Struct(c)
}

func provideConfiguration(cfg WithMiddleware) *Configuration {
	return cfg.MiddlewareConfiguration()
}
