package konfetty

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type FileFormat int

const (
	YAML FileFormat = iota
	JSON
	TOML

	defaultStructTag = "konfetty"
)

type LoaderConfig struct {
	KoanfDelimiter string
	EnvPrefix      string
	FileFormat     FileFormat
	StructTag      string
}

// Loader is the main interface for loading and validating configurations.
type Loader[T any] interface {
	Load(paths ...string) (*T, error)
	Validate(cfg *T) error
	Transform(cfg *T) error
}

type loader[T any] struct {
	k            *koanf.Koanf
	config       LoaderConfig
	validateFn   func(*T) error
	transformFn  func(*T) error
	unmarshalFn  func(interface{}) (*T, error)
	koanfSetupFn func(*koanf.Koanf) error
}

type Option[T any] func(*loader[T])

func NewLoader[T any](options ...Option[T]) Loader[T] {
	l := &loader[T]{
		config: LoaderConfig{
			KoanfDelimiter: ".",
			StructTag:      defaultStructTag,
			FileFormat:     YAML,
		},
	}

	for _, option := range options {
		option(l)
	}

	l.k = koanf.New(l.config.KoanfDelimiter)

	return l
}

func WithValidator[T any](fn func(*T) error) Option[T] {
	return func(l *loader[T]) {
		l.validateFn = fn
	}
}

func WithTransformer[T any](fn func(*T) error) Option[T] {
	return func(l *loader[T]) {
		l.transformFn = fn
	}
}

func WithUnmarshal[T any](fn func(interface{}) (*T, error)) Option[T] {
	return func(l *loader[T]) {
		l.unmarshalFn = fn
	}
}

func WithKoanfSetup[T any](fn func(*koanf.Koanf) error) Option[T] {
	return func(l *loader[T]) {
		l.koanfSetupFn = fn
	}
}

func WithEnvPrefix[T any](prefix string) Option[T] {
	return func(l *loader[T]) {
		l.config.EnvPrefix = prefix
	}
}

func WithFileFormat[T any](format FileFormat) Option[T] {
	return func(l *loader[T]) {
		l.config.FileFormat = format
	}
}

func WithStructTag[T any](tag string) Option[T] {
	return func(l *loader[T]) {
		l.config.StructTag = tag
	}
}

func WithKoanfDelimiter[T any](delimiter string) Option[T] {
	return func(l *loader[T]) {
		l.config.KoanfDelimiter = delimiter
	}
}

// Load implements the ConfigLoader interface.
func (l *loader[T]) Load(paths ...string) (*T, error) {
	// Apply custom koanf setup if provided
	if l.koanfSetupFn != nil {
		if err := l.koanfSetupFn(l.k); err != nil {
			return nil, fmt.Errorf("koanf setup: %w", err)
		}
	} else {
		// Default setup
		if err := l.defaultKoanfSetup(paths); err != nil {
			return nil, fmt.Errorf("default koanf setup: %w", err)
		}
	}

	// Unmarshal
	var cfg *T
	var err error
	if l.unmarshalFn != nil {
		cfg, err = l.unmarshalFn(l.k.Raw())
	} else {
		cfg = new(T)
		decodeHook := mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(), // Convert strings to time.Duration
			mapstructure.StringToSliceHookFunc(","),     // Convert comma-separated strings to slices
		)

		err = l.k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
			Tag: l.config.StructTag,
			DecoderConfig: &mapstructure.DecoderConfig{
				Result:           &cfg,
				WeaklyTypedInput: true,
				Squash:           true,
				TagName:          l.config.StructTag,
				DecodeHook:       decodeHook,
			},
		})
	}
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// Transform
	if l.transformFn != nil {
		if err := l.transformFn(cfg); err != nil {
			return nil, fmt.Errorf("transform: %w", err)
		}
	}

	// Validate
	if l.validateFn != nil {
		if err := l.validateFn(cfg); err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}
	}

	return cfg, nil
}

func (l *loader[T]) defaultKoanfSetup(paths []string) error {
	var parser koanf.Parser
	switch l.config.FileFormat {
	case YAML:
		parser = yaml.Parser()
	case JSON:
		parser = json.Parser()
	case TOML:
		parser = toml.Parser()
	default:
		return errors.New("unsupported file format")
	}

	for _, path := range paths {
		if err := l.k.Load(file.Provider(path), parser); err != nil {
			return fmt.Errorf("load file %s: %w", path, err)
		}
	}

	if l.config.EnvPrefix != "" {
		if err := l.k.Load(env.Provider(l.config.EnvPrefix, l.config.KoanfDelimiter, func(s string) string {
			return strings.Replace(strings.ToLower(strings.TrimPrefix(s, l.config.EnvPrefix)), "_", l.config.KoanfDelimiter, -1)
		}), nil); err != nil {
			return fmt.Errorf("load env vars: %w", err)
		}
	}

	return nil
}

// Transform implements the ConfigLoader interface.
func (l *loader[T]) Transform(cfg *T) error {
	if l.transformFn != nil {
		return l.transformFn(cfg)
	}
	return nil
}

// Validate implements the ConfigLoader interface.
func (l *loader[T]) Validate(cfg *T) error {
	if l.validateFn != nil {
		return l.validateFn(cfg)
	}

	return nil
}
