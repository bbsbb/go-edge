package otelfx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type OtelSuite struct {
	suite.Suite
}

func (s *OtelSuite) TestEffectiveSampleRate() {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"zero defaults to 1.0", 0, 1.0},
		{"half stays 0.5", 0.5, 0.5},
		{"one stays 1.0", 1.0, 1.0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Assert().Equal(tt.want, effectiveSampleRate(tt.input))
		})
	}
}

func (s *OtelSuite) TestSetup_Disabled() {
	otel.SetTracerProvider(nil)

	app := fxtest.New(s.T(),
		fx.Supply(&Configuration{
			Enabled:     false,
			ServiceName: "test-svc",
		}),
		fx.Invoke(setup),
	)
	app.RequireStart()

	_, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	s.Assert().False(ok, "expected no SDK tracer provider when disabled")

	app.RequireStop()
}

func (s *OtelSuite) TestSetup_Enabled() {
	cfg := &Configuration{
		Enabled:     true,
		ServiceName: "test-svc",
		Endpoint:    "localhost:4318",
		Insecure:    true,
		SampleRate:  1.0,
	}

	var hook fx.Hook
	p := params{
		Config: cfg,
		Lifecycle: lifecycleStub{onStop: func(h fx.Hook) {
			hook = h
		}},
	}

	err := setup(p)
	s.Require().NoError(err)

	_, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	s.Assert().True(ok, "expected SDK tracer provider when enabled")

	_, ok = otel.GetMeterProvider().(*sdkmetric.MeterProvider)
	s.Assert().True(ok, "expected SDK meter provider when enabled")

	// Shut down with a canceled context to avoid blocking on unreachable endpoint.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = hook.OnStop(ctx)
}

type lifecycleStub struct {
	fx.Lifecycle
	onStop func(fx.Hook)
}

func (l lifecycleStub) Append(h fx.Hook) {
	if l.onStop != nil {
		l.onStop(h)
	}
}

func TestOtelSuite(t *testing.T) {
	suite.Run(t, new(OtelSuite))
}
