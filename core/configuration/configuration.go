// Package configuration provides environment-based configuration loading
// with support for YAML files and environment variable overrides.
package configuration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/sethvargo/go-envconfig"

	"github.com/bbsbb/go-edge/core/secretstore"
)

var errSecretResolution = errors.New("secret resolution failed")

type Environment string

var (
	Development = Environment("development")
	Testing     = Environment("testing")
	Staging     = Environment("staging")
	Production  = Environment("production")
)

func (e Environment) IsValid() bool {
	return slices.Contains([]Environment{Development, Testing, Staging, Production}, e)
}

func (e Environment) IsDevelopment() bool {
	return e == Development
}

func (e Environment) IsProduction() bool {
	return e == Production
}

type WithValidation interface {
	Validate() error
}

type Option func(*options) *options

type options struct {
	path              string
	environmentPrefix string
	secretsService    secretstore.Service
}

func WithPath(path string) Option {
	return func(o *options) *options {
		o.path = path
		return o
	}
}

func WithEnvironmentPrefix(prefix string) Option {
	return func(o *options) *options {
		o.environmentPrefix = prefix
		return o
	}
}

func WithSecrets(s secretstore.Service) Option {
	return func(o *options) *options {
		o.secretsService = s
		return o
	}
}

type secretLookuper struct {
	lookuper envconfig.Lookuper
	service  secretstore.Service
	err      error
}

var _ envconfig.Lookuper = (*secretLookuper)(nil)

func (s *secretLookuper) Lookup(key string) (string, bool) {
	value, ok := s.lookuper.Lookup(key)
	if !ok {
		return key, false
	}

	if secretName, found := strings.CutPrefix(value, "secret://"); found {
		secretValue, err := s.service.GetSecretValue(secretName)
		if err != nil {
			s.err = fmt.Errorf("secret %q for key %s: %w", secretName, key, err)
			return "", true
		}
		return secretValue, true
	}

	return value, true
}

func instantiateGeneric[T any]() T {
	typeOfGeneric := reflect.TypeOf((*T)(nil)).Elem()

	if typeOfGeneric.Kind() == reflect.Ptr {
		return reflect.New(typeOfGeneric.Elem()).Interface().(T)
	}

	return reflect.Zero(typeOfGeneric).Interface().(T)
}

func LoadConfiguration[T any](ctx context.Context, environment Environment, opt ...Option) (T, error) {
	cfgInstance := instantiateGeneric[T]()
	lookupFns := []envconfig.Lookuper{}
	opts := &options{}

	for _, funcOpt := range opt {
		opts = funcOpt(opts)
	}

	if opts.path != "" {
		filePath := fmt.Sprintf("%s%s%s.yaml", filepath.Clean(opts.path), string(os.PathSeparator), environment)
		bs, err := os.ReadFile(filePath) //nolint:gosec // path built from trusted config prefix + environment enum
		if err != nil {
			return cfgInstance, fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}

		if err = yaml.Unmarshal(bs, cfgInstance); err != nil {
			return cfgInstance, fmt.Errorf("failed to parse config file %s: %w", filePath, err)
		}
	}

	lookupFns = append(lookupFns, envconfig.OsLookuper())

	var sl *secretLookuper
	for i, fn := range lookupFns {
		newFn := fn

		if opts.secretsService != nil {
			sl = &secretLookuper{
				lookuper: newFn,
				service:  opts.secretsService,
			}
			newFn = sl
		}

		if opts.environmentPrefix != "" {
			newFn = envconfig.PrefixLookuper(opts.environmentPrefix, newFn)
		}

		lookupFns[i] = newFn
	}

	lookuper := envconfig.MultiLookuper(lookupFns...)
	err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   cfgInstance,
		Lookuper: lookuper,
	})
	if err != nil {
		return cfgInstance, fmt.Errorf("failed to process environment variables: %w", err)
	}

	if sl != nil && sl.err != nil {
		return cfgInstance, fmt.Errorf("%w: %w", errSecretResolution, sl.err)
	}

	if cfgWithValidation, ok := any(cfgInstance).(WithValidation); ok {
		if err = cfgWithValidation.Validate(); err != nil {
			return cfgInstance, fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return cfgInstance, nil
}
