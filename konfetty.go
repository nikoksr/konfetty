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

type FileFormat uint8

const (
	FileFormatYAML FileFormat = iota
	FileFormatJSON
	FileFormatTOML
)

const defaultStructTag = "konfetty"

// Loader is the main interface for loading and validating configurations.
type Loader[T any] interface {
	Load(paths ...string) (*T, error)
	Validate(cfg *T) error
}

// Option is a function type for configuring the loader.
type Option[T any] func(*loader[T])

// loader is the internal implementation of ConfigLoader.
type loader[T any] struct {
	k          *koanf.Koanf
	envPrefix  string
	fileFormat FileFormat
	structTag  string
	validateFn func(*T) error
}

// NewLoader creates a new configuration loader.
func NewLoader[T any](options ...Option[T]) Loader[T] {
	l := &loader[T]{
		k: koanf.NewWithConf(koanf.Conf{
			Delim:       ".",
			StrictMerge: true,
		}),
		fileFormat: FileFormatYAML, // default to YAML
		structTag:  defaultStructTag,
	}

	for _, option := range options {
		option(l)
	}

	return l
}

// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix[T any](prefix string) Option[T] {
	return func(l *loader[T]) {
		l.envPrefix = prefix
	}
}

// WithFileFormat sets the format for configuration files.
func WithFileFormat[T any](format FileFormat) Option[T] {
	return func(l *loader[T]) {
		l.fileFormat = format
	}
}

// WithStructTag sets the struct tag for configuration fields.
func WithStructTag[T any](tag string) Option[T] {
	return func(l *loader[T]) {
		l.structTag = tag
	}
}

// WithValidator sets a custom validation function.
func WithValidator[T any](fn func(*T) error) Option[T] {
	return func(l *loader[T]) {
		l.validateFn = fn
	}
}

// Load implements the ConfigLoader interface.
func (l *loader[T]) Load(paths ...string) (*T, error) {
	var cfg T

	// Load from files
	for _, path := range paths {
		if err := l.loadFile(path); err != nil {
			return nil, fmt.Errorf("read from file: %w", err)
		}
	}

	// Load from environment variables
	if err := l.loadEnv(); err != nil {
		return nil, fmt.Errorf("read from environment: %w", err)
	}

	// Unmarshal into the config struct
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(), // Convert strings to time.Duration
		mapstructure.StringToSliceHookFunc(","),     // Convert comma-separated strings to slices
	)

	if err := l.k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		Tag: l.structTag,
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           &cfg,
			WeaklyTypedInput: true,
			Squash:           true,
			TagName:          l.structTag,
			DecodeHook:       decodeHook,
		},
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Apply defaults
	if err := fillDefaults(&cfg); err != nil {
		return nil, fmt.Errorf("fill defaults: %w", err)
	}

	// Validate
	if err := l.Validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &cfg, nil
}

// loadFile loads configuration from a file.
func (l *loader[T]) loadFile(path string) error {
	var parser koanf.Parser

	switch l.fileFormat {
	case FileFormatYAML:
		parser = yaml.Parser()
	case FileFormatJSON:
		parser = json.Parser()
	case FileFormatTOML:
		parser = toml.Parser()
	default:
		return errors.New("unsupported file format")
	}

	if err := l.k.Load(file.Provider(path), parser); err != nil {
		return err
	}

	return nil
}

// loadEnv loads configuration from environment variables.
func (l *loader[T]) loadEnv() error {
	return l.k.Load(env.Provider(l.envPrefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, l.envPrefix)), "_", ".")
	}), nil)
}

// Validate implements the ConfigLoader interface.
func (l *loader[T]) Validate(cfg *T) error {
	if l.validateFn != nil {
		return l.validateFn(cfg)
	}

	return nil
}

// MustLoad is a helper function that panics on error.
func MustLoad[T any](loader Loader[T], paths ...string) *T {
	cfg, err := loader.Load(paths...)
	if err != nil {
		panic(err)
	}

	return cfg
}
