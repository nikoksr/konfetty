<div align="center">

&nbsp;
<h1>konfetty 🎉</h1>
<p><i>Zero-dependency, type-safe and powerful post-processing for your existing config solution in Go.</i></p>

&nbsp;

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/nikoksr/konfetty)
[![codecov](https://codecov.io/gh/nikoksr/konfetty/graph/badge.svg?token=lySNULyXHL)](https://codecov.io/gh/nikoksr/konfetty)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikoksr/konfetty)](https://goreportcard.com/report/github.com/nikoksr/konfetty)
[![Maintainability](https://api.codeclimate.com/v1/badges/e87ea581a2617e6afb36/maintainability)](https://codeclimate.com/github/nikoksr/konfetty/maintainability)
</div>

&nbsp;

# About <a id="about"></a>

Konfetty is a Go library that simplifies the management of hierarchical default values in complex struct hierarchies, primarily designed for, but not limited to, configuration management. It addresses the challenge of applying defaults to nested structs, interfaces, and embedded types while maintaining type safety.

Key features:
- 🔍 Recursively applies defaults through nested structures
- 🏗️ Respects type hierarchies, allowing base type defaults to be overridden by more specific types
- 🛡️ Maintains full type safety without relying on struct tags or runtime type assertions
- 🔌 Integrates with existing configuration loading solutions as a post-processing step
- 🧩 Applicable to any struct-based hierarchies, not just configurations (e.g., middleware chains, complex domain models)
- 🔧 Supports custom transformations and validations as part of the processing pipeline

Konfetty aims to reduce the boilerplate typically associated with setting default values in Go struct hierarchies, allowing developers to focus on their core application logic rather than complex default value management.

## Quick Start <a id="quick-start"></a>

```go
package main

import "github.com/nikoksr/konfetty"

type BaseDevice struct {
    Enabled bool
}

type LightDevice struct {
    BaseDevice
    Brightness int
}

type ThermostatDevice struct {
    BaseDevice
    Temperature float64
}

type RoomConfig struct {
    Devices []any
}

func main() {
		// Stubbing a configuration, usually pre-populated by your config provider.
    cfg := &RoomConfig{
        Devices: []any{
            &LightDevice{BaseDevice: BaseDevice{Enabled: true}},  // A light device that's enabled by default
            &LightDevice{Brightness: 75},                         // A light device with a custom brightness
            &ThermostatDevice{},                                  // An empty thermostat device
        },
    }

    cfg, err := konfetty.FromConfig(cfg).
        WithDefaults(
            BaseDevice{Enabled: false},           // Devices are disabled by default, unless overridden, as seen above
            LightDevice{Brightness: 50},          // Light devices have a default brightness of 50, unless overridden
            ThermostatDevice{
            	BaseDevice: BaseDevice{Enabled: true},  // Override the base device default for thermostats
				Temperature: 20.0,                      // Thermostat devices have a default temperature of 20.0, unless overridden
			},
        ).
        WithTransformer(func(cfg *RoomConfig) {
        	// Optional custom transformation logic for more complex processing
        }).
        WithValidator(func(cfg *RoomConfig) error {
            // Optional custom validation logic
            return nil
        }).
        Build()

    // Handle error ...

    // The processed config would look like this:
    //
    // &RoomConfig{
    //     Devices: []any{
    //         &LightDevice{BaseDevice: BaseDevice{Enabled: true}, Brightness: 50},           // The first light device stays enabled and was given a brightness of 50
    //         &LightDevice{BaseDevice: BaseDevice{Enabled: false}, Brightness: 75},          // The second light device was disabled and given a brightness of 75
    //         &ThermostatDevice{BaseDevice: BaseDevice{Enabled: true}, Temperature: 20.0},   // The thermostat device was enabled and given a temperature of 20.0
    //     },
    // }

    // Continue using your config struct as usual ...
}

```

In this example, Konfetty automatically applies the `BaseDevice` defaults to all devices, then overlays the specific defaults for `LightDevice` and `ThermostatDevice`. This happens recursively through the entire `RoomConfig` structure all while maintaining type safety.

## Installation <a id="installation"></a>

```bash
go get -u github.com/nikoksr/konfetty
```

## How Konfetty Works <a id="how-it-works"></a>

Konfetty's approach to default values sets it apart:

- Define defaults for base types once, and have them applied automatically throughout your struct hierarchy, even in nested slices of different types.
- Easily override lower-level defaults with more specific ones, giving you fine-grained control.
- Maintain full type safety throughout the default application process, unlike solutions using struct tags or reflection-based approaches.

The processing pipeline: Recursively apply defaults > apply (optional) transformations > run (optional) validations

## Integration <a id="integration"></a>

Konfetty doesn't replace your current config loading mechanism — it enhances it. Use it as a powerful post-processing step after loading your config with Viper, Koanf, or any other solution.

### With Viper <a id="integration-viper"></a>

```go
viper.ReadInConfig()
viper.Unmarshal(&config)

config, err := konfetty.FromConfig(&config).
    WithDefaults(defaultConfig).
    WithTransformer(transformer).
    WithValidator(validator).
    Build()
```

### With Koanf <a id="integration-koanf"></a>

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.Unmarshal("", &config)

config, err := konfetty.FromConfig(&config).
    WithDefaults(defaultConfig).
    Build()
```

## Usage Examples <a id="examples"></a>

- [Simple Example](examples/simple/main.go): A basic example demonstrating how to use Konfetty with a simple configuration structure.
- [Complex Example](examples/complex/main.go): A more complex example showcasing the power of Konfetty's hierarchical default system.
- [Viper Integration](examples/viper/main.go): A full example demonstrating how to integrate Konfetty with Viper.
- [Koanf Integration](examples/koanf/main.go): A full example demonstrating how to integrate Konfetty with Koanf.

## Contributing <a id="contributing"></a>

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for more details.

## Support <a id="support"></a>

If you find this project useful, consider giving it a ⭐️! Your support helps bring more attention to the project, allowing us to enhance it even further.

While you're here, feel free to check out my other work:

- [nikoksr/notify](https://github.com/nikoksr/notify) - A dead simple Go library for sending notifications to various messaging services.
