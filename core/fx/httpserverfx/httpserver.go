// Package httpserverfx provides an fx module for HTTP server with Chi router.
package httpserverfx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/configuration"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var _ configuration.WithValidation = (*Configuration)(nil)

type CorsConfiguration struct {
	AllowedOrigins   []string `yaml:"allowed_origins" env:"ALLOWED_ORIGINS,overwrite"`
	AllowedMethods   []string `yaml:"allowed_methods" env:"ALLOWED_METHODS,overwrite"`
	AllowedHeaders   []string `yaml:"allowed_headers" env:"ALLOWED_HEADERS,overwrite"`
	AllowCredentials bool     `yaml:"allow_credentials" env:"ALLOW_CREDENTIALS,overwrite"`
	MaxAge           int      `yaml:"max_age" env:"MAX_AGE,overwrite" validate:"gte=0,lte=86400"`
}

func (c *CorsConfiguration) allowedMethodsOrDefault() []string {
	if len(c.AllowedMethods) > 0 {
		return c.AllowedMethods
	}
	return []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
}

func (c *CorsConfiguration) allowedHeadersOrDefault() []string {
	if len(c.AllowedHeaders) > 0 {
		return c.AllowedHeaders
	}
	return []string{"Content-Type", "Authorization"}
}

func (c *CorsConfiguration) maxAgeOrDefault() int {
	if c.MaxAge > 0 {
		return c.MaxAge
	}
	return 300
}

type Configuration struct {
	Port              uint16            `yaml:"port" env:"PORT,overwrite" validate:"required,gte=1,lte=65535"`
	RequestTimeout    uint8             `yaml:"request_timeout" env:"REQUEST_TIMEOUT,overwrite" validate:"required,gte=1,lte=120"`
	ReadHeaderTimeout uint8             `yaml:"read_header_timeout" env:"READ_HEADER_TIMEOUT,overwrite" validate:"lte=120"`
	ReadTimeout       uint8             `yaml:"read_timeout" env:"READ_TIMEOUT,overwrite" validate:"lte=120"`
	IdleTimeout       uint16            `yaml:"idle_timeout" env:"IDLE_TIMEOUT,overwrite" validate:"lte=600"`
	Cors              CorsConfiguration `yaml:"cors" env:"CORS,overwrite"`
}

func (c *Configuration) Validate() error {
	return validate.Struct(c)
}

type WithHTTPServer interface {
	HTTPServerConfiguration() *Configuration
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *Configuration
	Logger    *slog.Logger
}

type Result struct {
	fx.Out
	Server *http.Server
	Mux    *chi.Mux
}

func NewHTTPServer(p Params) (Result, error) {
	mux := chi.NewRouter()

	if len(p.Config.Cors.AllowedOrigins) > 0 {
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins:   p.Config.Cors.AllowedOrigins,
			AllowedMethods:   p.Config.Cors.allowedMethodsOrDefault(),
			AllowedHeaders:   p.Config.Cors.allowedHeadersOrDefault(),
			AllowCredentials: p.Config.Cors.AllowCredentials,
			MaxAge:           p.Config.Cors.maxAgeOrDefault(),
		}))
	}

	mux.Use(middleware.Timeout(time.Duration(p.Config.RequestTimeout) * time.Second))

	readHeaderTimeout := time.Duration(p.Config.ReadHeaderTimeout) * time.Second
	if readHeaderTimeout == 0 {
		readHeaderTimeout = 5 * time.Second
	}
	readTimeout := time.Duration(p.Config.ReadTimeout) * time.Second
	if readTimeout == 0 {
		readTimeout = 10 * time.Second
	}
	idleTimeout := time.Duration(p.Config.IdleTimeout) * time.Second
	if idleTimeout == 0 {
		idleTimeout = 120 * time.Second
	}

	srv := &http.Server{
		Handler:           mux,
		Addr:              fmt.Sprintf(":%d", p.Config.Port),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      time.Duration(p.Config.RequestTimeout+5) * time.Second,
		IdleTimeout:       idleTimeout,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					p.Logger.Error("HTTP server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := srv.Shutdown(ctx); err != nil {
				return fmt.Errorf("httpserverfx: shutdown: %w", err)
			}
			return nil
		},
	})

	return Result{Server: srv, Mux: mux}, nil
}

func provideConfiguration(cfg WithHTTPServer) *Configuration {
	return cfg.HTTPServerConfiguration()
}

var Module = fx.Module(
	"httpserverfx",
	fx.Provide(provideConfiguration, NewHTTPServer),
	// Forces server instantiation so lifecycle hooks register even if nothing else depends on *http.Server.
	fx.Invoke(func(_ *http.Server) {}),
)
