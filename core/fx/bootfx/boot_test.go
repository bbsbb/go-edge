package bootfx

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type BootSuite struct {
	suite.Suite
}

type stubConfig struct{}

func (s *stubConfig) AsFx() fx.Option {
	return fx.Options()
}

func (s *BootSuite) TestBootFxComposesWithoutError() {
	opt := BootFx(&stubConfig{})
	app := fx.New(opt, fx.NopLogger)
	s.Require().NoError(app.Err())
}

func (s *BootSuite) TestBootFxAcceptsExtraOptions() {
	type marker struct{}
	opt := BootFx(&stubConfig{}, fx.Supply(&marker{}))
	app := fx.New(opt, fx.NopLogger)
	s.Require().NoError(app.Err())
}

func TestBootSuite(t *testing.T) {
	suite.Run(t, new(BootSuite))
}
