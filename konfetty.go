// Package konfetty provides a flexible, type-safe configuration processing system.
package konfetty

import (
	"errors"
	"fmt"
	"reflect"
)

// ConfigProvider defines an interface for loading configurations.
type ConfigProvider[T any] interface {
	Load() (T, error)
}

// configSource is an internal type to represent the source of configuration.
type configSource[T any] struct {
	config     *T
	loaderFunc func() (T, error)
	provider   ConfigProvider[T]
}

// ConfigBuilder orchestrates the configuration building process. It manages the configuration source, defaults,
// transformations, and validations.
type ConfigBuilder[T any] struct {
	source    configSource[T]
	defaults  map[reflect.Type][]any
	transform func(*T)
	validate  func(*T) error
}

// ConfigProcessor exposes methods for further configuration processing. It wraps a ConfigBuilder and provides a fluent
// interface for configuration setup.
type ConfigProcessor[T any] struct {
	builder *ConfigBuilder[T]
}

// FromConfig initializes a ConfigProcessor with a pre-loaded configuration. Use this when you already have a populated
// config struct:
//
//	cfg := &MyConfig{...}
//	processor := konfetty.FromConfig(cfg)
func FromConfig[T any](config *T) *ConfigProcessor[T] {
	return &ConfigProcessor[T]{
		builder: &ConfigBuilder[T]{
			source: configSource[T]{config: config},
		},
	}
}

// FromLoaderFunc initializes a ConfigProcessor with a function that loads the configuration. Use this when you have a
// custom loading function:
//
//	loader := func() (MyConfig, error) { ... }
//	processor := konfetty.FromLoaderFunc(loader)
func FromLoaderFunc[T any](loader func() (T, error)) *ConfigProcessor[T] {
	return &ConfigProcessor[T]{
		builder: &ConfigBuilder[T]{
			source: configSource[T]{loaderFunc: loader},
		},
	}
}

// FromProvider initializes a ConfigProcessor with a ConfigProvider. Use this when you have an implementation of the
// ConfigProvider interface:
//
//	provider := MyConfigProvider{}
//	processor := konfetty.FromProvider(provider)
func FromProvider[T any](provider ConfigProvider[T]) *ConfigProcessor[T] {
	return &ConfigProcessor[T]{
		builder: &ConfigBuilder[T]{
			source: configSource[T]{provider: provider},
		},
	}
}

// WithDefaults adds default values to the configuration processing pipeline. Multiple defaults can be provided and will
// be applied in order.
func (p *ConfigProcessor[T]) WithDefaults(defaultValues ...any) *ConfigProcessor[T] {
	if p.builder.defaults == nil {
		p.builder.defaults = make(map[reflect.Type][]any)
	}

	for _, dv := range defaultValues {
		t := reflect.TypeOf(dv)
		p.builder.defaults[t] = append(p.builder.defaults[t], dv)
	}

	return p
}

// WithTransformer sets a custom transformation function to be applied to the configuration.
func (p *ConfigProcessor[T]) WithTransformer(fn func(*T)) *ConfigProcessor[T] {
	p.builder.transform = fn
	return p
}

// WithValidator sets a custom validation function to be applied to the configuration.
func (p *ConfigProcessor[T]) WithValidator(fn func(*T) error) *ConfigProcessor[T] {
	p.builder.validate = fn
	return p
}

// Build processes the configuration, applying defaults, transformations, and validations. It returns the final
// configuration or an error if any step fails.
func (p *ConfigProcessor[T]) Build() (*T, error) {
	return p.builder.build()
}

func (b *ConfigBuilder[T]) build() (*T, error) {
	// Load
	cfg, err := b.load()
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	// Apply defaults
	if err = applyDefaults(&cfg, b.defaults); err != nil {
		return nil, fmt.Errorf("apply defaults: %w", err)
	}

	// Transform
	if b.transform != nil {
		b.transform(&cfg)
	}

	// Validate
	if b.validate != nil {
		if err = b.validate(&cfg); err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}
	}

	return &cfg, nil
}

func (b *ConfigBuilder[T]) load() (T, error) {
	var cfg T
	var err error

	switch {
	case b.source.config != nil:
		cfg = *b.source.config
	case b.source.loaderFunc != nil:
		cfg, err = b.source.loaderFunc()
		if err != nil {
			return cfg, fmt.Errorf("from loader func: %w", err)
		}
	case b.source.provider != nil:
		cfg, err = b.source.provider.Load()
		if err != nil {
			return cfg, fmt.Errorf("from provider: %w", err)
		}
	default:
		return cfg, errors.New("no configuration source provided")
	}

	return cfg, nil
}
