package configuration

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/tests/mocks"
)

type TestConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port int    `yaml:"port" env:"SERVER_PORT,overwrite"`
	Host string `yaml:"host" env:"SERVER_HOST,overwrite"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" env:"DATABASE_HOST,overwrite"`
	Port     int    `yaml:"port" env:"DATABASE_PORT,overwrite"`
	Name     string `yaml:"name" env:"DATABASE_NAME,overwrite"`
	Password string `yaml:"password" env:"DATABASE_PASSWORD,overwrite"`
}

type ConfigurationSuite struct {
	suite.Suite
}

func (s *ConfigurationSuite) TestLoadConfiguration_WithPath() {
	ctx := context.Background()

	cfg, err := LoadConfiguration[*TestConfig](ctx, Testing, WithPath("./testdata"))
	s.Require().NoError(err)

	s.Assert().Equal(8080, cfg.Server.Port)
	s.Assert().Equal("localhost", cfg.Server.Host)
	s.Assert().Equal("localhost", cfg.Database.Host)
	s.Assert().Equal(5432, cfg.Database.Port)
	s.Assert().Equal("testdb", cfg.Database.Name)
}

func (s *ConfigurationSuite) TestLoadConfiguration_WithEnvironmentPrefix() {
	ctx := context.Background()

	s.T().Setenv("TEST_SERVER_PORT", "9090")
	s.T().Setenv("TEST_DATABASE_NAME", "overridden_db")

	cfg, err := LoadConfiguration[*TestConfig](ctx, Testing,
		WithPath("./testdata"),
		WithEnvironmentPrefix("TEST_"),
	)
	s.Require().NoError(err)

	s.Assert().Equal(9090, cfg.Server.Port)
	s.Assert().Equal("overridden_db", cfg.Database.Name)
	s.Assert().Equal("localhost", cfg.Server.Host)
}

func (s *ConfigurationSuite) TestLoadConfiguration_WithSecrets() {
	ctx := context.Background()

	s.T().Setenv("TEST_DATABASE_PASSWORD", "secret://db-password")

	mockService := mocks.NewMockSecretsService(s.T())
	mockService.EXPECT().GetSecretValue("db-password").Return("super-secret-password", nil).Once()

	cfg, err := LoadConfiguration[*TestConfig](ctx, Testing,
		WithPath("./testdata"),
		WithEnvironmentPrefix("TEST_"),
		WithSecrets(mockService),
	)
	s.Require().NoError(err)

	s.Assert().Equal("super-secret-password", cfg.Database.Password)
}

func (s *ConfigurationSuite) TestLoadConfiguration_WithSecretsError() {
	ctx := context.Background()

	s.T().Setenv("TEST_DATABASE_PASSWORD", "secret://db-password")

	mockService := mocks.NewMockSecretsService(s.T())
	mockService.EXPECT().GetSecretValue("db-password").Return("", errors.New("secret not found")).Once()

	_, err := LoadConfiguration[*TestConfig](ctx, Testing,
		WithPath("./testdata"),
		WithEnvironmentPrefix("TEST_"),
		WithSecrets(mockService),
	)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, errSecretResolution)
	s.Assert().Contains(err.Error(), "secret not found")
}

func (s *ConfigurationSuite) TestLoadConfiguration_FileNotFound() {
	ctx := context.Background()

	_, err := LoadConfiguration[*TestConfig](ctx, Testing, WithPath("./nonexistent"))

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "failed to read config file")
}

func (s *ConfigurationSuite) TestLoadConfiguration_InvalidYAML() {
	ctx := context.Background()

	_, err := LoadConfiguration[*TestConfig](ctx, Testing, WithPath("./testdata/invalid"))

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "failed to parse config file")
}

type ValidatedConfig struct {
	Value string `yaml:"value" env:"VALUE"`
}

func (c *ValidatedConfig) Validate() error {
	if c.Value == "" {
		return errors.New("value is required")
	}
	return nil
}

type testEnum string

func (e testEnum) IsValid() bool {
	return e == "valid" || e == "also_valid"
}

type StringEnumConfig struct {
	Status testEnum `validate:"required,stringenum"`
}

func (c *StringEnumConfig) Validate() error {
	return Validate.Struct(c)
}

func (s *ConfigurationSuite) TestLoadConfiguration_ValidationFailure() {
	ctx := context.Background()

	_, err := LoadConfiguration[*ValidatedConfig](ctx, Testing, WithPath("./testdata/empty"))

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "configuration validation failed")
	s.Assert().Contains(err.Error(), "value is required")
}

func (s *ConfigurationSuite) TestValidate_StringEnum_Valid() {
	cfg := &StringEnumConfig{Status: "valid"}
	err := cfg.Validate()
	s.Require().NoError(err)
}

func (s *ConfigurationSuite) TestValidate_StringEnum_Invalid() {
	cfg := &StringEnumConfig{Status: "invalid"}
	err := cfg.Validate()
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "Status")
}

func (s *ConfigurationSuite) TestValidate_StringEnum_Empty() {
	cfg := &StringEnumConfig{Status: ""}
	err := cfg.Validate()
	s.Require().Error(err)
}

func TestConfigurationSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationSuite))
}
