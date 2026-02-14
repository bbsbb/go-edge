package secretstore

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EnvServiceSuite struct {
	suite.Suite
}

func (s *EnvServiceSuite) TestGetSecretValue_Found() {
	s.T().Setenv("DB_PASSWORD", "hunter2")
	svc := NewEnvService("")

	val, err := svc.GetSecretValue("db-password")
	s.Require().NoError(err)
	s.Assert().Equal("hunter2", val)
}

func (s *EnvServiceSuite) TestGetSecretValue_WithPrefix() {
	s.T().Setenv("SECRET_JWT_SIGNING_KEY", "my-key")
	svc := NewEnvService("SECRET")

	val, err := svc.GetSecretValue("jwt-signing-key")
	s.Require().NoError(err)
	s.Assert().Equal("my-key", val)
}

func (s *EnvServiceSuite) TestGetSecretValue_NotFound() {
	svc := NewEnvService("")

	_, err := svc.GetSecretValue("nonexistent")
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "NONEXISTENT")
	s.Assert().Contains(err.Error(), "not set")
}

func TestEnvServiceSuite(t *testing.T) {
	suite.Run(t, new(EnvServiceSuite))
}
