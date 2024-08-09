// Package konfetty provides zero-dependency, type-safe and powerful post-processing for your data structs,
// mostly focused on applying defaults, transformations, and validations to configuration structures.
package konfetty

import (
	"errors"
	"fmt"
	"reflect"
)

// Provider defines an interface for loading structured data.
type Provider[T any] interface {
	Load() (T, error)
}

// dataSource is an internal type to represent the source of data.
type dataSource[T any] struct {
	data       *T
	loaderFunc func() (T, error)
	provider   Provider[T]
}

// Builder orchestrates the building process. It manages the data source, defaults, transformations, and validations.
type Builder[T any] struct {
	source    dataSource[T]
	defaults  map[reflect.Type][]any
	transform func(*T)
	validate  func(*T) error
}

// Processor exposes methods for further data-structure processing. It wraps a Builder and provides a fluent interface
// for configuration setup.
type Processor[T any] struct {
	builder *Builder[T]
}

// FromStruct initializes a Processor with a pre-populated struct.
//
//	cfg := &MyConfig{...}
//	processor := konfetty.FromStruct(cfg)
func FromStruct[T any](config *T) *Processor[T] {
	return &Processor[T]{
		builder: &Builder[T]{
			source: dataSource[T]{data: config},
		},
	}
}

// FromLoaderFunc initializes a Processor with a function that loads the data-structure.
//
//	loader := func() (MyConfig, error) { ... }
//	processor := konfetty.FromLoaderFunc(loader)
func FromLoaderFunc[T any](loader func() (T, error)) *Processor[T] {
	return &Processor[T]{
		builder: &Builder[T]{
			source: dataSource[T]{loaderFunc: loader},
		},
	}
}

// FromProvider initializes a Processor with a Provider.
//
//	provider := MyConfigProvider{}
//	processor := konfetty.FromProvider(provider)
func FromProvider[T any](provider Provider[T]) *Processor[T] {
	return &Processor[T]{
		builder: &Builder[T]{
			source: dataSource[T]{provider: provider},
		},
	}
}

// WithDefaults adds default values to the processing pipeline. Multiple defaults can be provided and will be applied
// in order.
func (p *Processor[T]) WithDefaults(defaultValues ...any) *Processor[T] {
	if p.builder.defaults == nil {
		p.builder.defaults = make(map[reflect.Type][]any)
	}

	for _, dv := range defaultValues {
		t := reflect.TypeOf(dv)
		p.builder.defaults[t] = append(p.builder.defaults[t], dv)
	}

	return p
}

// WithTransformer sets a custom transformation function to be applied to the data-structure.
func (p *Processor[T]) WithTransformer(fn func(*T)) *Processor[T] {
	p.builder.transform = fn
	return p
}

// WithValidator sets a custom validation function to be applied to the data-structure.
func (p *Processor[T]) WithValidator(fn func(*T) error) *Processor[T] {
	p.builder.validate = fn
	return p
}

// Build processes the data-structure, applying defaults, transformations, and validations. It returns the final
// struct or an error if any step fails.
func (p *Processor[T]) Build() (*T, error) {
	return p.builder.build()
}

func (b *Builder[T]) build() (*T, error) {
	cfg, err := b.load()
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	if err = applyDefaults(&cfg, b.defaults); err != nil {
		return nil, fmt.Errorf("apply defaults: %w", err)
	}

	if b.transform != nil {
		b.transform(&cfg)
	}

	if b.validate != nil {
		if err = b.validate(&cfg); err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}
	}

	return &cfg, nil
}

func (b *Builder[T]) load() (T, error) {
	var cfg T
	var err error

	switch {
	case b.source.data != nil:
		cfg = *b.source.data
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
		return cfg, errors.New("no data source provided")
	}

	return cfg, nil
}
