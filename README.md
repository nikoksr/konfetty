<div align="center">

&nbsp;
<h1>konfetty</h1>
<p><i>Zero-dependency, type-safe and powerful post-processing for your existing config solution.</i></p>

&nbsp;

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/nikoksr/konfetty)
[![codecov](https://codecov.io/gh/nikoksr/konfetty/graph/badge.svg?token=lySNULyXHL)](https://codecov.io/gh/nikoksr/konfetty)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikoksr/konfetty)](https://goreportcard.com/report/github.com/nikoksr/konfetty)
[![Maintainability](https://api.codeclimate.com/v1/badges/e87ea581a2617e6afb36/maintainability)](https://codeclimate.com/github/nikoksr/konfetty/maintainability)
</div>

&nbsp;

```go
type BaseDevice struct {
    Enabled bool
    Type    string
}

type LightDevice struct {
    BaseDevice
    Brightness int
}

// Other device types...

type SmartHomeConfig struct {
    Devices []any
}

// Use your preferred way of loading the config, e.g., Viper, Koanf, etc.

// Let Konfetty handle the complexity of recursively setting deeply nested defaults.
cfg, err := konfetty.FromConfig(&SmartHomeConfig{}).
	WithDefaults(
		// Defaults for all BaseDevice instances.
		BaseDevice{Enabled: false, Type: "unknown"},
		// Defaults for all LightDevice instances, atomically overriding defaults for BaseDevice.
		LightDevice{BaseDevice: BaseDevice{Type: "light"}, Brightness: 50},
	).
	WithTransformer(func(cfg *SmartHomeConfig) {
		// Apply type-safe, custom transformations to the config.
	}).
	WithValidator(func(cfg *SmartHomeConfig) error {
		// Apply type-safe, custom validations to the config.
	}).
	Build()

// Result: All devices are disabled by default, with the type set to "unknown". Light devices have a brightness of 50 and a type of "light".
```

Konfetty is a zero-dependency, type-safe, and powerful post-processing library for your existing configuration solution. It enhances your current setup by providing a seamless way to apply hierarchical defaults, custom transformations, and validations to your configuration.

> Take a moment and think about how you'd ensure that every device in the `SmartHomeConfig` structure has the correct default values. You'd need to write recursive functions, type switches, and careful merging logic. Konfetty handles all of this for you, seamlessly.

## Installation

```bash
go get -u github.com/nikoksr/konfetty
```

## Key Features

1. **Intelligent Hierarchical Default System**:
   Define defaults once for base types and let konfetty recursively apply them throughout your entire config structure. Easily override these base defaults for more specific types, all while maintaining full type safety and maximum simplicity.

   This feature is the main motivation for me personally behind creating Konfetty. I've worked on several projects where managing defaults for complex configurations was a nightmare. Konfetty solves this problem elegantly and efficiently. Features like the transformation and validation steps were added to round up the library and provide a complete solution for post-processing configurations.

2. **Type-Safe Configuration**: Utilize Go's generics for fully typed access to your config structs at every stage of processing. Compared to struct-tag based solutions, Konfetty provides a safer, more robust config management experience that catches errors at compile-time rather than runtime.

3. **Seamless Integration**: Works with your current config loading mechanism (Viper, Koanf, etc.). Simply plug your config in directly or use one of the other supported config-provision methods. Konfetty enhances your existing setup by providing a powerful post-processing step for your configuration.

4. **Streamline transformation and validation**: Pipeline your configuration through custom transformation and validation functions. Konfetty provides a clean, type-safe way to apply transformations and validations to your configuration after filling the gaps with defaults.

## Why Konfetty?

### Powerful Hierarchical Default System

Konfetty's approach to default values sets it apart:

- Define defaults for base types once, and have them applied automatically throughout your config structure, even in nested slices of different types.
- Easily override lower-level defaults with more specific ones, giving you fine-grained control over your configuration.
- Maintain full type safety throughout the default application process, unlike solutions using struct tags or reflection-based approaches.

This system shines when dealing with complex configurations:

```go
type Device interface {
    IsEnabled() bool
}

type BaseDevice struct {
    Enabled bool
    Type    string
}

type LightDevice struct {
    BaseDevice
    Brightness int
}

type ThermostatDevice struct {
    BaseDevice
    TargetTemp float64
}

type RoomConfig struct {
    Devices []Device
}

type HomeConfig struct {
    Rooms []RoomConfig
}

konfetty.FromConfig(&HomeConfig{}).
    WithDefaults(
        BaseDevice{Enabled: true, Type: "unknown"},
        LightDevice{BaseDevice: BaseDevice{Type: "light"}, Brightness: 50},
        ThermostatDevice{TargetTemp: 22.0},
    )
```

In this example, Konfetty automatically applies the `BaseDevice` defaults to all devices, then overlays the specific defaults for `LightDevice` and `ThermostatDevice`. This happens recursively through the entire `HomeConfig` structure, including within slices, all while maintaining type safety.

### Seamless Integration with Existing Solutions

Konfetty doesn't replace your current config loading mechanism—it enhances it:

- Use Konfetty as a powerful post-processing step after loading your config with Viper, Koanf, or any other solution.
- No need to rewrite your initial config loading logic; Konfetty works with your existing setup.

```go
viper.ReadInConfig()
viper.Unmarshal(&config)

processedConfig, err := konfetty.FromConfig(&config).
    WithDefaults(defaultConfig).
    WithTransformer(transformer).
    WithValidator(validator).
    Build()
```

### Type-Safe Configuration Handling

Konfetty leverages Go's type system to provide a safer, more robust config management experience:

- Fully typed access to your config structs at every stage of processing.
- Catch configuration errors at compile-time rather than runtime.
- No need for type assertions when working with complex, nested structures.

## Usage

Take a look at our examples to see different ways to use Konfetty:

- [Simple Example](examples/simple/main.go): A basic example demonstrating how to use Konfetty with a simple configuration structure.

- [Complex Example](examples/complex/main.go): A more complex example showcasing the power of Konfetty's hierarchical default system.

- [Viper Integration](examples/viper/main.go): An example demonstrating how to integrate Konfetty with Viper.

- [Koanf Integration](examples/koanf/main.go): An example demonstrating how to integrate Konfetty with Koanf.

## Contributing

Contributions are welcome. Please see our [Contributing Guide](CONTRIBUTING.md) for more details.
