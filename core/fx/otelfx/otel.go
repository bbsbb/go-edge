// Package otelfx provides an fx module for OpenTelemetry tracing and metrics.
package otelfx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/fx"
)

type params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *Configuration
	Logger    *slog.Logger `optional:"true"`
}

func setup(p params) error {
	if !p.Config.Enabled {
		if p.Logger != nil {
			p.Logger.Info("OpenTelemetry disabled")
		}
		return nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(p.Config.ServiceName),
		),
	)
	if err != nil {
		return fmt.Errorf("otelfx: create resource: %w", err)
	}

	tp, err := initTracer(ctx, p.Config, res)
	if err != nil {
		return fmt.Errorf("otelfx: init tracer: %w", err)
	}

	mp, err := initMeter(ctx, p.Config, res)
	if err != nil {
		return fmt.Errorf("otelfx: init meter: %w", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if p.Logger != nil {
				p.Logger.Info("shutting down OpenTelemetry")
			}
			if err := errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx)); err != nil {
				return fmt.Errorf("otelfx: shutdown: %w", err)
			}
			return nil
		},
	})

	if p.Logger != nil {
		p.Logger.Info("OpenTelemetry initialized",
			"endpoint", p.Config.Endpoint,
			"service", p.Config.ServiceName,
			"sample_rate", effectiveSampleRate(p.Config.SampleRate),
		)
	}

	return nil
}

func initTracer(ctx context.Context, cfg *Configuration, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	var opts []otlptracehttp.Option
	if cfg.Endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(effectiveSampleRate(cfg.SampleRate)))

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	), nil
}

func initMeter(ctx context.Context, cfg *Configuration, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	var opts []otlpmetrichttp.Option
	if cfg.Endpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	), nil
}

func provideLoggerProvider(lc fx.Lifecycle, cfg *Configuration) *sdklog.LoggerProvider {
	if !cfg.Enabled {
		return nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil
	}

	lp, err := initLogger(ctx, cfg, res)
	if err != nil {
		return nil
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return lp.Shutdown(ctx)
		},
	})

	return lp
}

func initLogger(ctx context.Context, cfg *Configuration, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	var opts []otlploghttp.Option
	if cfg.Endpoint != "" {
		opts = append(opts, otlploghttp.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create log exporter: %w", err)
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	), nil
}

func effectiveSampleRate(rate float64) float64 {
	if rate == 0 {
		return 1.0
	}
	return rate
}

func provideConfiguration(cfg WithOTel) *Configuration {
	return cfg.OTelConfiguration()
}

var Module = fx.Module(
	"otelfx",
	fx.Provide(provideConfiguration),
	fx.Provide(provideLoggerProvider),
	fx.Invoke(setup),
)
