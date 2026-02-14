package psqlfx

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"

	"github.com/bbsbb/go-edge/core/configuration"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var (
	_ configuration.WithValidation = (*Credentials)(nil)
	_ configuration.WithValidation = (*PoolConfiguration)(nil)
	_ configuration.WithValidation = (*Configuration)(nil)
)

type Credentials struct {
	Username string `json:"username" yaml:"username" env:"USERNAME" validate:"required"`
	Password string `json:"password" yaml:"password" env:"PASSWORD" validate:"required"`
}

func (c *Credentials) Validate() error {
	return validate.Struct(c)
}

// EnvDecode implements envconfig.Decoder to parse JSON credentials from an environment variable.
func (c *Credentials) EnvDecode(value string) error {
	if value == "" {
		return nil
	}
	return json.Unmarshal([]byte(value), c)
}

type PoolConfiguration struct {
	MaxIdleConns    int32         `yaml:"max_idle_conns" env:"MAX_IDLE_CONNS,overwrite" validate:"gte=0"`
	MaxOpenConns    int32         `yaml:"max_open_conns" env:"MAX_OPEN_CONNS,overwrite" validate:"gte=0"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"CONN_MAX_LIFETIME,overwrite" validate:"gte=0"`
}

func (p *PoolConfiguration) Validate() error {
	return validate.Struct(p)
}

func DefaultPoolConfiguration() *PoolConfiguration {
	return &PoolConfiguration{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}
}

type Configuration struct {
	Host        string             `yaml:"host" env:"HOST,overwrite" validate:"required,hostname"`
	Port        uint16             `yaml:"port" env:"PORT,overwrite" validate:"required,gte=1,lte=65535"`
	Database    string             `yaml:"database" env:"DATABASE,overwrite" validate:"required"`
	Credentials *Credentials       `yaml:"credentials" env:"CREDENTIALS,overwrite,noinit" validate:"required"`
	DisableSSL  bool               `yaml:"disable_ssl" env:"DISABLE_SSL,overwrite"`
	Pool        *PoolConfiguration `yaml:"pool" env:"POOL,overwrite"`
}

func (c *Configuration) Validate() error {
	return validate.Struct(c)
}

func (c *Configuration) DSN() string {
	sslMode := "verify-full"
	if c.DisableSSL {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Credentials.Username, c.Credentials.Password, c.Database, sslMode,
	)
}

// ConnectionDefaults configures postgres settings at the connection level via DSN.
// Timeouts are specified in milliseconds.
// See https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT.
type ConnectionDefaults struct {
	ApplicationName                 string `yaml:"application_name" mapstructure:"application_name,omitempty"`
	IdleInTransactionSessionTimeout uint32 `yaml:"idle_in_transaction_session_timeout" mapstructure:"idle_in_transaction_session_timeout,omitempty"`
	StatementTimeout                uint32 `yaml:"statement_timeout" mapstructure:"statement_timeout,omitempty"`
}

func (d *ConnectionDefaults) DSN() (string, error) {
	var serialized map[string]any
	if err := mapstructure.Decode(d, &serialized); err != nil {
		return "", fmt.Errorf("encode connection defaults: %w", err)
	}

	keys := make([]string, 0, len(serialized))
	for k := range serialized {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, serialized[k]))
	}
	return strings.Join(parts, " "), nil
}
