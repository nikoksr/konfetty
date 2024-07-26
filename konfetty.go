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

// FileFormat represents the supported configuration file formats.
type FileFormat int

const (
	YAML FileFormat = iota
	JSON
	TOML
)

// LoaderConfig holds the configuration options for the Loader.
type LoaderConfig struct {
	KoanfDelimiter string
	EnvPrefix      string
	FileFormat     FileFormat
	StructTag      string
	MaxDepth       int
}

// Loader is the main interface for loading and validating configurations.
// It provides methods to load configuration from various sources, validate,
// and transform the loaded configuration.
type Loader[T any] interface {
	Load(paths ...string) (*T, error)
	Validate(cfg *T) error
	Transform(cfg *T) error
}

// loader is the internal implementation of the Loader interface.
type loader[T any] struct {
	k            *koanf.Koanf
	config       LoaderConfig
	validateFn   func(*T) error
	transformFn  func(*T) error
	unmarshalFn  func(interface{}) (*T, error)
	koanfSetupFn func(*koanf.Koanf) error
}

// Option is a function type used to configure the Loader.
type Option[T any] func(*loader[T])

// NewLoader creates a new Loader instance with the provided options.
// It returns a Loader interface that can be used to load, validate, and transform configurations.
func NewLoader[T any](options ...Option[T]) Loader[T] {
	l := &loader[T]{
		config: LoaderConfig{
			KoanfDelimiter: ".",
			StructTag:      "konfetty",
			FileFormat:     YAML,
			MaxDepth:       100,
		},
	}

	for _, option := range options {
		option(l)
	}

	l.k = koanf.New(l.config.KoanfDelimiter)

	return l
}

// WithValidator sets a custom validation function for the Loader.
func WithValidator[T any](fn func(*T) error) Option[T] {
	return func(l *loader[T]) {
		l.validateFn = fn
	}
}

// WithTransformer sets a custom transformation function for the Loader.
func WithTransformer[T any](fn func(*T) error) Option[T] {
	return func(l *loader[T]) {
		l.transformFn = fn
	}
}

// WithUnmarshal sets a custom unmarshal function for the Loader.
func WithUnmarshal[T any](fn func(interface{}) (*T, error)) Option[T] {
	return func(l *loader[T]) {
		l.unmarshalFn = fn
	}
}

// WithKoanfSetup sets a custom Koanf setup function for the Loader.
func WithKoanfSetup[T any](fn func(*koanf.Koanf) error) Option[T] {
	return func(l *loader[T]) {
		l.koanfSetupFn = fn
	}
}

// WithKoanfDelimiter sets the delimiter used by Koanf for nested keys.
func WithKoanfDelimiter[T any](delimiter string) Option[T] {
	return func(l *loader[T]) {
		l.config.KoanfDelimiter = delimiter
	}
}

// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix[T any](prefix string) Option[T] {
	return func(l *loader[T]) {
		l.config.EnvPrefix = prefix
	}
}

// WithFileFormat sets the file format for configuration files.
func WithFileFormat[T any](format FileFormat) Option[T] {
	return func(l *loader[T]) {
		l.config.FileFormat = format
	}
}

// WithStructTag sets the struct tag used for unmarshaling.
func WithStructTag[T any](tag string) Option[T] {
	return func(l *loader[T]) {
		l.config.StructTag = tag
	}
}

// WithMaxDepth sets the maximum recursion depth for filling defaults.
func WithMaxDepth[T any](depth int) Option[T] {
	return func(l *loader[T]) {
		l.config.MaxDepth = depth
	}
}

// Load implements the ConfigLoader interface.
// It loads configuration from the provided file paths, applies defaults,
// transforms, and validates the configuration.
func (l *loader[T]) Load(paths ...string) (*T, error) {
	// Apply custom koanf setup if provided, otherwise use default setup
	if l.koanfSetupFn != nil {
		if err := l.koanfSetupFn(l.k); err != nil {
			return nil, fmt.Errorf("koanf setup: %w", err)
		}
	} else {
		if err := l.defaultKoanfSetup(paths); err != nil {
			return nil, fmt.Errorf("default koanf setup: %w", err)
		}
	}

	// Unmarshal configuration
	var cfg *T
	var err error
	if l.unmarshalFn != nil {
		cfg, err = l.unmarshalFn(l.k.Raw())
	} else {
		cfg = new(T)
		decodeHook := mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
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

	// Apply defaults
	// if err := fillDefaults(cfg, l.config.MaxDepth); err != nil {
	// 	return nil, fmt.Errorf("fill defaults: %w", err)
	// }

	fillDefaults(cfg)

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

// defaultKoanfSetup sets up the default Koanf configuration.
// It loads configuration from files and environment variables.
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

	// Load configuration from files
	for _, path := range paths {
		if err := l.k.Load(file.Provider(path), parser); err != nil {
			return fmt.Errorf("load file %s: %w", path, err)
		}
	}

	// Load configuration from environment variables
	if l.config.EnvPrefix != "" {
		if err := l.k.Load(env.Provider(l.config.EnvPrefix, l.config.KoanfDelimiter, func(s string) string {
			return strings.Replace(strings.ToLower(strings.TrimPrefix(s, l.config.EnvPrefix)), "_", l.config.KoanfDelimiter, -1)
		}), nil); err != nil {
			return fmt.Errorf("load env vars: %w", err)
		}
	}

	return nil
}

// Transform applies the custom transformation function to the configuration.
func (l *loader[T]) Transform(cfg *T) error {
	if l.transformFn != nil {
		return l.transformFn(cfg)
	}
	return nil
}

// Validate applies the custom validation function to the configuration.
func (l *loader[T]) Validate(cfg *T) error {
	if l.validateFn != nil {
		return l.validateFn(cfg)
	}
	return nil
}
