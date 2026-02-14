package rlsfx

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigurationSuite struct {
	suite.Suite
}

func (s *ConfigurationSuite) TestValidate() {
	tests := []struct {
		name    string
		config  Configuration
		wantErr bool
	}{
		{
			name:    "valid",
			config:  Configuration{Schema: "app", Field: "current_organization"},
			wantErr: false,
		},
		{
			name:    "missing schema",
			config:  Configuration{Field: "current_organization"},
			wantErr: true,
		},
		{
			name:    "missing field",
			config:  Configuration{Schema: "app"},
			wantErr: true,
		},
		{
			name:    "empty",
			config:  Configuration{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.config.Validate()
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

func (s *ConfigurationSuite) TestNewRLS() {
	cfg := &Configuration{Schema: "app", Field: "current_organization"}
	db, err := NewRLS(params{Config: cfg})

	s.Require().NoError(err)
	s.Assert().NotNil(db)
	s.Assert().Equal("app", db.schema)
	s.Assert().Equal("current_organization", db.field)
}

func (s *ConfigurationSuite) TestNewRLS_EmptySchema() {
	cfg := &Configuration{Schema: "", Field: "current_organization"}
	_, err := NewRLS(params{Config: cfg})

	s.Require().Error(err)
	s.Assert().ErrorIs(err, ErrMissingSchema)
}

func TestConfigurationSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationSuite))
}
