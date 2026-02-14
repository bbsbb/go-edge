package psqlfx

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ConfigurationSuite struct {
	suite.Suite
}

func (s *ConfigurationSuite) TestCredentials_Validate() {
	tests := []struct {
		name    string
		creds   Credentials
		wantErr bool
	}{
		{
			name:    "valid credentials",
			creds:   Credentials{Username: "user", Password: "pass"},
			wantErr: false,
		},
		{
			name:    "missing username",
			creds:   Credentials{Password: "pass"},
			wantErr: true,
		},
		{
			name:    "missing password",
			creds:   Credentials{Username: "user"},
			wantErr: true,
		},
		{
			name:    "empty credentials",
			creds:   Credentials{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.creds.Validate()
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ConfigurationSuite) TestCredentials_EnvDecode() {
	tests := []struct {
		name     string
		input    string
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{
			name:     "valid JSON",
			input:    `{"username":"testuser","password":"testpass"}`,
			wantUser: "testuser",
			wantPass: "testpass",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			wantUser: "",
			wantPass: "",
			wantErr:  false,
		},
		{
			name:    "invalid JSON",
			input:   "invalid json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var creds Credentials
			err := creds.EnvDecode(tt.input)
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Assert().Equal(tt.wantUser, creds.Username)
				s.Assert().Equal(tt.wantPass, creds.Password)
			}
		})
	}
}

func (s *ConfigurationSuite) TestConfiguration_Validate() {
	validCreds := &Credentials{Username: "user", Password: "pass"}

	tests := []struct {
		name       string
		config     Configuration
		wantErr    bool
		wantFields []string
	}{
		{
			name: "valid config",
			config: Configuration{
				Host:        "localhost",
				Port:        5432,
				Database:    "testdb",
				Credentials: validCreds,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: Configuration{
				Host:        "",
				Port:        5432,
				Database:    "testdb",
				Credentials: validCreds,
			},
			wantErr:    true,
			wantFields: []string{"Host"},
		},
		{
			name: "invalid port zero",
			config: Configuration{
				Host:        "localhost",
				Port:        0,
				Database:    "testdb",
				Credentials: validCreds,
			},
			wantErr:    true,
			wantFields: []string{"Port"},
		},
		{
			name: "empty database",
			config: Configuration{
				Host:        "localhost",
				Port:        5432,
				Database:    "",
				Credentials: validCreds,
			},
			wantErr:    true,
			wantFields: []string{"Database"},
		},
		{
			name: "nil credentials",
			config: Configuration{
				Host:        "localhost",
				Port:        5432,
				Database:    "testdb",
				Credentials: nil,
			},
			wantErr:    true,
			wantFields: []string{"Credentials"},
		},
		{
			name: "missing credential username",
			config: Configuration{
				Host:        "localhost",
				Port:        5432,
				Database:    "testdb",
				Credentials: &Credentials{Password: "pass"},
			},
			wantErr:    true,
			wantFields: []string{"Username"},
		},
		{
			name: "multiple errors",
			config: Configuration{
				Host:        "",
				Port:        0,
				Database:    "",
				Credentials: nil,
			},
			wantErr:    true,
			wantFields: []string{"Host", "Port", "Database", "Credentials"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.config.Validate()
			if tt.wantErr {
				s.Require().Error(err)
				var validationErrs validator.ValidationErrors
				s.Require().ErrorAs(err, &validationErrs)

				fields := make([]string, len(validationErrs))
				for i, fe := range validationErrs {
					fields[i] = fe.Field()
				}
				s.Assert().ElementsMatch(tt.wantFields, fields)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

func (s *ConfigurationSuite) TestConfiguration_DSN() {
	tests := []struct {
		name     string
		config   Configuration
		expected string
	}{
		{
			name: "with SSL",
			config: Configuration{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Credentials: &Credentials{
					Username: "user",
					Password: "pass",
				},
				DisableSSL: false,
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=verify-full",
		},
		{
			name: "without SSL",
			config: Configuration{
				Host:     "db.example.com",
				Port:     5433,
				Database: "proddb",
				Credentials: &Credentials{
					Username: "admin",
					Password: "secret",
				},
				DisableSSL: true,
			},
			expected: "host=db.example.com port=5433 user=admin password=secret dbname=proddb sslmode=disable",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Assert().Equal(tt.expected, tt.config.DSN())
		})
	}
}

func (s *ConfigurationSuite) TestDefaultPoolConfiguration() {
	pool := DefaultPoolConfiguration()

	s.Assert().Equal(int32(10), pool.MaxIdleConns)
	s.Assert().Equal(int32(100), pool.MaxOpenConns)
	s.Assert().Equal(time.Hour, pool.ConnMaxLifetime)
}

func (s *ConfigurationSuite) TestPoolConfiguration_Validate() {
	tests := []struct {
		name    string
		pool    PoolConfiguration
		wantErr bool
	}{
		{
			name:    "valid defaults",
			pool:    PoolConfiguration{},
			wantErr: false,
		},
		{
			name: "valid with values",
			pool: PoolConfiguration{
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "invalid negative MaxIdleConns",
			pool: PoolConfiguration{
				MaxIdleConns: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid negative MaxOpenConns",
			pool: PoolConfiguration{
				MaxOpenConns: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid negative ConnMaxLifetime",
			pool: PoolConfiguration{
				ConnMaxLifetime: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.pool.Validate()
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ConfigurationSuite) TestConnectionDefaults_DSN() {
	tests := []struct {
		name     string
		defaults ConnectionDefaults
		expected string
	}{
		{
			name:     "empty defaults",
			defaults: ConnectionDefaults{},
			expected: "",
		},
		{
			name: "application name only",
			defaults: ConnectionDefaults{
				ApplicationName: "myapp",
			},
			expected: "application_name=myapp",
		},
		{
			name: "statement timeout only",
			defaults: ConnectionDefaults{
				StatementTimeout: 5000,
			},
			expected: "statement_timeout=5000",
		},
		{
			name: "all params",
			defaults: ConnectionDefaults{
				ApplicationName:                 "myapp",
				IdleInTransactionSessionTimeout: 10000,
				StatementTimeout:                5000,
			},
			expected: "application_name=myapp idle_in_transaction_session_timeout=10000 statement_timeout=5000",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			dsn, err := tt.defaults.DSN()
			s.Require().NoError(err)
			s.Assert().Equal(tt.expected, dsn)
		})
	}
}

func TestConfigurationSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationSuite))
}
