// Package bootfx provides application bootstrapping utilities for fx.
package bootfx

import (
	"github.com/bbsbb/go-edge/core/fx/loggerfx"

	"go.uber.org/fx"
)

type WithFx interface {
	AsFx() fx.Option
}

func BootFx(cfg WithFx, extra ...fx.Option) fx.Option {
	return fx.Options(
		cfg.AsFx(),
		loggerfx.Module,
		fx.Options(extra...),
		fx.WithLogger(loggerfx.NewFxSlogLogger),
	)
}
