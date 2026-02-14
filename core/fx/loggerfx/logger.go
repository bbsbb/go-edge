// Package loggerfx provides an fx module for structured logging using slog.
package loggerfx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/lmittmann/tint"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/configuration"
)

var (
	ErrInvalidLogLevel  = errors.New("loggerfx: invalid log level")
	ErrInvalidLogFormat = errors.New("loggerfx: invalid log format")
)

var _ configuration.WithValidation = (*Configuration)(nil)

type LogLevel string

var (
	LogLevelDebug = LogLevel("debug")
	LogLevelInfo  = LogLevel("info")
	LogLevelWarn  = LogLevel("warn")
	LogLevelError = LogLevel("error")
)

func (l LogLevel) IsValid() bool {
	return slices.Contains([]LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}, l)
}

func (l LogLevel) SlogLevel() slog.Level {
	switch l {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type LogFormat string

var (
	LogFormatText = LogFormat("text")
	LogFormatJSON = LogFormat("json")
)

func (f LogFormat) IsValid() bool {
	return slices.Contains([]LogFormat{LogFormatText, LogFormatJSON}, f)
}

type Configuration struct {
	Level      LogLevel  `yaml:"level" env:"LOG_LEVEL,overwrite" validate:"required,stringenum"`
	Format     LogFormat `yaml:"format" env:"LOG_FORMAT,overwrite" validate:"required,stringenum"`
	OTelBridge bool      `yaml:"otel_bridge" env:"OTEL_BRIDGE,overwrite"`
}

func (c *Configuration) Validate() error {
	return configuration.Validate.Struct(c)
}

type WithLogging interface {
	LoggingConfiguration() *Configuration
}

type Params struct {
	fx.In
	Config         *Configuration
	LoggerProvider *sdklog.LoggerProvider `optional:"true"`
}

type Result struct {
	fx.Out
	Logger *slog.Logger
}

func newHandler(cfg *Configuration, w io.Writer) (slog.Handler, error) {
	if !cfg.Level.IsValid() {
		return nil, ErrInvalidLogLevel
	}
	if !cfg.Format.IsValid() {
		return nil, ErrInvalidLogFormat
	}

	if cfg.Format == LogFormatJSON {
		return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: cfg.Level.SlogLevel()}), nil
	}

	return tint.NewHandler(w, &tint.Options{
		Level:      cfg.Level.SlogLevel(),
		TimeFormat: time.DateTime,
	}), nil
}

func newLogger(cfg *Configuration, w io.Writer) (*slog.Logger, error) {
	handler, err := newHandler(cfg, w)
	if err != nil {
		return nil, err
	}
	return slog.New(handler), nil
}

func NewLogger(p Params) (Result, error) {
	handler, err := newHandler(p.Config, os.Stdout)
	if err != nil {
		return Result{}, fmt.Errorf("loggerfx: create logger: %w", err)
	}

	if p.Config.OTelBridge && p.LoggerProvider != nil {
		bridge := otelslog.NewHandler("",
			otelslog.WithLoggerProvider(p.LoggerProvider),
		)
		handler = &fanoutHandler{handlers: []slog.Handler{handler, bridge}}
	}

	return Result{Logger: slog.New(handler)}, nil
}

type fanoutHandler struct {
	handlers []slog.Handler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			errs = append(errs, handler.Handle(ctx, record))
		}
	}
	return errors.Join(errs...)
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: handlers}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: handlers}
}

func provideConfiguration(cfg WithLogging) *Configuration {
	return cfg.LoggingConfiguration()
}

var Module = fx.Module(
	"loggerfx",
	fx.Provide(provideConfiguration, NewLogger),
)
