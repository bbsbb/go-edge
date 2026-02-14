package loggerfx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggerSuite struct {
	suite.Suite
}

func (s *LoggerSuite) TestLogLevel_IsValid() {
	tests := []struct {
		level LogLevel
		valid bool
	}{
		{LogLevelDebug, true},
		{LogLevelInfo, true},
		{LogLevelWarn, true},
		{LogLevelError, true},
		{LogLevel("invalid"), false},
		{LogLevel(""), false},
	}

	for _, tt := range tests {
		s.Run(string(tt.level), func() {
			s.Assert().Equal(tt.valid, tt.level.IsValid())
		})
	}
}

func (s *LoggerSuite) TestLogLevel_SlogLevel() {
	tests := []struct {
		level    LogLevel
		expected slog.Level
	}{
		{LogLevelDebug, slog.LevelDebug},
		{LogLevelInfo, slog.LevelInfo},
		{LogLevelWarn, slog.LevelWarn},
		{LogLevelError, slog.LevelError},
		{LogLevel("invalid"), slog.LevelInfo},
	}

	for _, tt := range tests {
		s.Run(string(tt.level), func() {
			s.Assert().Equal(tt.expected, tt.level.SlogLevel())
		})
	}
}

func (s *LoggerSuite) TestLogFormat_IsValid() {
	tests := []struct {
		format LogFormat
		valid  bool
	}{
		{LogFormatText, true},
		{LogFormatJSON, true},
		{LogFormat("invalid"), false},
		{LogFormat(""), false},
	}

	for _, tt := range tests {
		s.Run(string(tt.format), func() {
			s.Assert().Equal(tt.valid, tt.format.IsValid())
		})
	}
}

func (s *LoggerSuite) TestNewLogger() {
	tests := []struct {
		name    string
		config  Configuration
		wantErr error
	}{
		{
			name:   "text format",
			config: Configuration{Level: LogLevelDebug, Format: LogFormatText},
		},
		{
			name:   "json format",
			config: Configuration{Level: LogLevelError, Format: LogFormatJSON},
		},
		{
			name:    "invalid level",
			config:  Configuration{Level: LogLevel("invalid"), Format: LogFormatText},
			wantErr: ErrInvalidLogLevel,
		},
		{
			name:    "invalid format",
			config:  Configuration{Level: LogLevelInfo, Format: LogFormat("invalid")},
			wantErr: ErrInvalidLogFormat,
		},
		{
			name:    "empty config",
			config:  Configuration{},
			wantErr: ErrInvalidLogLevel,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := NewLogger(Params{Config: &tt.config})
			if tt.wantErr != nil {
				s.Require().Error(err)
				s.Assert().ErrorIs(err, tt.wantErr)
			} else {
				s.Require().NoError(err)
				s.Assert().NotNil(result.Logger)
			}
		})
	}
}

func (s *LoggerSuite) TestNewLogger_TextOutput() {
	var buf bytes.Buffer
	logger, err := newLogger(&Configuration{
		Level:  LogLevelInfo,
		Format: LogFormatText,
	}, &buf)
	s.Require().NoError(err)

	logger.Info("test message", "key", "value")

	output := buf.String()
	s.Assert().Contains(output, "INF")
	s.Assert().Contains(output, "test message")
	s.Assert().Contains(output, "key=")
	s.Assert().Contains(output, "value")
}

func (s *LoggerSuite) TestNewLogger_JSONOutput() {
	var buf bytes.Buffer
	logger, err := newLogger(&Configuration{
		Level:  LogLevelInfo,
		Format: LogFormatJSON,
	}, &buf)
	s.Require().NoError(err)

	logger.Info("test message", "key", "value")

	var logEntry map[string]any
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	s.Require().NoError(err)

	s.Assert().Equal("INFO", logEntry["level"])
	s.Assert().Equal("test message", logEntry["msg"])
	s.Assert().Equal("value", logEntry["key"])
	s.Assert().Contains(logEntry, "time")
}

func (s *LoggerSuite) TestNewLogger_RespectsLogLevel() {
	var buf bytes.Buffer
	logger, err := newLogger(&Configuration{
		Level:  LogLevelWarn,
		Format: LogFormatText,
	}, &buf)
	s.Require().NoError(err)

	logger.Info("info message")
	logger.Warn("warn message")

	output := buf.String()
	s.Assert().NotContains(output, "info message")
	s.Assert().Contains(output, "warn message")
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}
