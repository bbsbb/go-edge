package otelfx

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigurationSuite struct {
	suite.Suite
}

func (s *ConfigurationSuite) TestValidate_Valid() {
	cfg := &Configuration{
		ServiceName: "my-service",
		SampleRate:  0.5,
	}
	s.Assert().NoError(cfg.Validate())
}

func (s *ConfigurationSuite) TestValidate_MissingServiceName() {
	cfg := &Configuration{
		SampleRate: 0.5,
	}
	err := cfg.Validate()
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "ServiceName")
}

func (s *ConfigurationSuite) TestValidate_SampleRateTooHigh() {
	cfg := &Configuration{
		ServiceName: "my-service",
		SampleRate:  1.5,
	}
	err := cfg.Validate()
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "SampleRate")
}

func (s *ConfigurationSuite) TestValidate_SampleRateNegative() {
	cfg := &Configuration{
		ServiceName: "my-service",
		SampleRate:  -0.1,
	}
	err := cfg.Validate()
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "SampleRate")
}

func (s *ConfigurationSuite) TestValidate_EnabledWithoutEndpoint() {
	cfg := &Configuration{
		Enabled:     true,
		ServiceName: "my-service",
		SampleRate:  0.5,
	}
	err := cfg.Validate()
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "Endpoint")
}

func (s *ConfigurationSuite) TestValidate_EnabledWithEndpoint() {
	cfg := &Configuration{
		Enabled:     true,
		Endpoint:    "localhost:4317",
		ServiceName: "my-service",
		SampleRate:  0.5,
	}
	s.Assert().NoError(cfg.Validate())
}

func (s *ConfigurationSuite) TestValidate_DisabledWithoutEndpoint() {
	cfg := &Configuration{
		Enabled:     false,
		ServiceName: "my-service",
		SampleRate:  0.5,
	}
	s.Assert().NoError(cfg.Validate())
}

func (s *ConfigurationSuite) TestValidate_SampleRateZero() {
	cfg := &Configuration{
		ServiceName: "my-service",
		SampleRate:  0,
	}
	s.Assert().NoError(cfg.Validate())
}

func TestConfigurationSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationSuite))
}
