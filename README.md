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

Konfetty is a Go library that solves the challenge of managing default values in complex configuration hierarchies. Whether you're using Viper, Koanf, or any other configuration solution, Konfetty adds powerful post-processing capabilities while maintaining complete type safety.

Key features:
- 🔍 Recursively applies defaults through nested structures
- 🏗️ Respects type hierarchies, allowing base type defaults to be overridden by more specific types
- 🛡️ Maintains compile-time type safety and eliminates the need for error-prone struct tags
- 🔌 Integrates with existing configuration loading solutions as a post-processing step
- 🧩 Applies to any struct-based hierarchies, not just configurations (e.g., middleware chains, complex domain models)
- 🔧 Supports custom transformations and validations as part of the processing pipeline

Konfetty reduces the boilerplate typically associated with setting default values in complex Go struct hierarchies, allowing developers to focus on their core application logic rather than complex default value management.

> [!NOTE]
> Konfetty is designed for use in single-threaded contexts, typically during application startup for configuration processing. Each `Processor` instance should be used by a single goroutine.

## Installation <a id="installation"></a>

```bash
go get -u github.com/nikoksr/konfetty
```

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
    // Stub configuration, typically pre-populated by your config provider (e.g., Viper or Koanf)
    cfg := &RoomConfig{
        Devices: []any{
            // A light device that's enabled by default
            &LightDevice{BaseDevice: BaseDevice{Enabled: true}},

            // A light device with a custom brightness
            &LightDevice{Brightness: 75},

            // An empty thermostat device
            &ThermostatDevice{},
        },
    }

    cfg, err := konfetty.FromStruct(cfg).
        WithDefaults(
            // Devices are disabled by default
            BaseDevice{Enabled: false},

            // Light devices have a default brightness of 50
            LightDevice{Brightness: 50},

            // Thermostat devices have a default temperature of 20.0 and are enabled by default
            ThermostatDevice{
                // Override the base device default for thermostats
                BaseDevice: BaseDevice{Enabled: true},
                Temperature: 20.0,
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
    // {
    //   "Devices": [
    //     {
    //       // LightDevice
    //       "Enabled": true,     // Kept original value
    //       "Brightness": 50     // Used LightDevice default
    //     },
    //     {
    //       // LightDevice
    //       "Enabled": false,    // Used BaseDevice default
    //       "Brightness": 75     // Kept original value
    //     },
    //     {
    //       // ThermostatDevice
    //       "Enabled": true,     // Used ThermostatDevice default, overriding BaseDevice default
    //       "Temperature": 20.0  // Used ThermostatDevice default
    //     }
    //   ]
    // }

    // Continue using your config struct as usual ...
}
```

In this example, Konfetty automatically applies the `BaseDevice` defaults to all devices, then overlays the specific defaults for `LightDevice` and `ThermostatDevice`. This happens recursively through the entire `RoomConfig` structure while maintaining compile-time type safety.

## How Konfetty Works <a id="how-it-works"></a>

Konfetty's approach to default values sets it apart:

- Define defaults for base types once, and they'll be applied automatically throughout your struct hierarchy, even in nested slices of different types
- Override lower-level defaults with more specific ones for fine-grained control
- Have type safety enforced at compile time, eliminating the need for error-prone struct tags

The processing pipeline follows this order: Recursively apply defaults > apply (optional) transformations > run (optional) validations

## Core Concepts <a id="core-concepts"></a>

### Default Value Resolution <a id="cc-default-value-resolution"></a>

Konfetty applies defaults in a specific order:

1. Base type defaults are applied first
2. More specific type defaults override base defaults
3. Existing non-zero values are always preserved (e.g., values set by your configuration provider)
4. Nested structures are processed recursively

```go
// Entity is our base type
type Entity struct {
    Name       string
    IsFriendly bool
}

// Companion is a more specific type that embeds Entity
type Companion struct {
    Entity            // Base entity properties
    LoyaltyLevel int
}

konfetty.FromStruct(&config).
    WithDefaults(
        // 1. Base type (Entity) defaults are applied first
        Entity{
            Name: "Unknown Entity", // Enforce all entities to have a default name
        },

        // 2. More specific type (Companion) defaults override base defaults
        Companion{
            Entity: Entity{
                Name:       "Dogmeat",  // Overrides (Base-) Entity's name ("Unknown Entity")
                IsFriendly: true,       // Overrides (Base-) Entity's default
            },
            LoyaltyLevel: 10,
        },
    )

// Note: Any existing non-zero values in 'config' would be preserved
// e.g., if config.Name was already set to "Rex", it would not be changed
```

### Type Safety <a id="cc-type-safety"></a>

Unlike solutions that rely on struct tags, Konfetty leverages Go's type system to enforce type safety at compile time. This prevents accidentally setting default values of the wrong type.

```go
type KonfettyDummy struct {
    Money int
}

type StructTagDummy struct {
    Money int `default:"I'm a string"` // This will compile but potentially cause runtime errors
}

konfetty.FromStruct(&KonfettyDummy{}).
    WithDefaults(
        KonfettyDummy{
            Money: "I'm a string", // This will not compile
        },
    )
```

### Recursive Defaults <a id="cc-recursive-defaults"></a>

A common approach to supplying default values is defining a config struct instance with default values. For example:

```go
type Config struct {
    Version string
    Enabled bool
}

// Define an instance of Config with default values that can be overridden by the config provider
var defaultConfig = Config{
    Version: "1.0",
    Enabled: true,
}
```

However, this approach becomes problematic with nested structs or slices of structs. Consider this more complex example from another project:

```go
type BaseProbe struct {
    Name     string
    Interval time.Duration
}

type HTTPProbe struct {
    BaseProbe
    Host string
}

// ... Other probe types

type Config struct {
    HTTPProbes []HTTPProbe
    // ... Other probes and fields
}
```

Using the simple approach, you might try:

```go
var defaultConfig = Config{
    HTTPProbes: []HTTPProbe{
        {
            BaseProbe: BaseProbe{
                Name:     "Default HTTP Probe",
                Interval: 5 * time.Second,
            },
            Host: "http://localhost",
        },
    },
}
```

But what if your config file already contains HTTP Probes, particularly incomplete ones? For example:

```yaml
http_probes:
  - name: "Incomplete Probe #1"
    host: "http://example.com"
  - name: "Incomplete Probe #2"
```

The `defaultConfig.HTTPProbes` would be overwritten by the loaded values, leaving incomplete probes that could cause runtime errors. You'd need to manually merge default values with loaded values. This is where Konfetty shines, automatically applying default values to nested structs and slices at any depth.

```go
konfetty.FromStruct(&config).
    WithDefaults(
        // Set sane defaults for all BaseProbes
        BaseProbe{
            Interval: 5 * time.Second,
        },

        // Fine-tune defaults for HTTPProbes
        HTTPProbe{
            BaseProbe: BaseProbe{
                Interval: 60 * time.Second, // Override BaseProbe default for all HTTPProbes
            },
            Host: "http://localhost",
        },

        // Define default Config structure for empty config files
        Config{
            HTTPProbes: []HTTPProbe{
                {
                    BaseProbe: BaseProbe{
                        // Only need to define the name; Konfetty will apply other defaults
                        Name: "Default HTTP Probe",
                    },
                },
            },
        },
    ).
    Build()
```

Here's how Konfetty handles different config file scenarios:

#### Fully Populated Config File

When all values are set in the config file, Konfetty doesn't need to apply defaults:

```yaml
http_probes:
  - name: "My Probe #1"
    host: "http://example.com"
    interval: "5s"
  - name: "My Probe #2"
    host: "http://localhost"
    interval: "15s"
```

Final config struct after Konfetty processing:

```go
Config{
    HTTPProbes: []HTTPProbe{
        {
            BaseProbe: BaseProbe{
                Name: "My Probe #1",
                Interval: 5 * time.Second,
            },
            Host: "http://example.com",
        },
        {
            BaseProbe: BaseProbe{
                Name: "My Probe #2",
                Interval: 15 * time.Second,
            },
            Host: "http://localhost",
        },
    },
}
```

#### Incomplete Config File

With two incomplete probes (first missing `interval`, second missing `host`):

```yaml
http_probes:
  - name: "Incomplete Probe #1"
    host: "http://example.com"
  - name: "Incomplete Probe #2"
    interval: "10s"
```

Final config struct after Konfetty processing:

```go
Config{
    HTTPProbes: []HTTPProbe{
        {
            BaseProbe: BaseProbe{
                Name: "Incomplete Probe #1",  // Kept from config file
                Interval: 60 * time.Second,   // Applied default
            },
            Host: "http://example.com",     // Kept from config file
        },
        {
            BaseProbe: BaseProbe{
                Name: "Incomplete Probe #2",  // Kept from config file
                Interval: 10 * time.Second,   // Kept from config file
            },
            Host: "http://localhost",       // Applied default
        },
    },
}
```

#### Empty Config File

With an empty config file:

```yaml
{}
```

Final config struct after Konfetty processing:

```go
Config{
    HTTPProbes: []HTTPProbe{
        {
            BaseProbe: BaseProbe{
                Name: "Default HTTP Probe",
                Interval: 60 * time.Second,
            },
            Host: "http://localhost",
        },
    },
}
```

## Integration <a id="integration"></a>

Konfetty complements your current config loading mechanism rather than replacing it. Use it as a post-processing step after loading your config with Viper, Koanf, or any other solution.

### With Viper <a id="integration-viper"></a>

```go
viper.ReadInConfig()
viper.Unmarshal(&config)

config, err := konfetty.FromStruct(&config).
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

config, err := konfetty.FromStruct(&config).
    WithDefaults(defaultConfig).
    WithTransformer(transformer).
    WithValidator(validator).
    Build()
```

## Usage Examples <a id="examples"></a>

- [Simple Example](examples/simple/main.go): A basic example demonstrating Konfetty with a simple configuration structure
- [Complex Example](examples/complex/main.go): A more complex example showcasing Konfetty's hierarchical default system
- [Viper Integration](examples/viper/main.go): A complete example demonstrating Konfetty integration with Viper
- [Koanf Integration](examples/koanf/main.go): A complete example demonstrating Konfetty integration with Koanf

## Contributing <a id="contributing"></a>

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for more details.

## Support <a id="support"></a>

If you find this project useful, consider giving it a ⭐️! Your support helps bring more attention to the project, enabling further improvements.

While you're here, check out my other work:

- [nikoksr/notify](https://github.com/nikoksr/notify) - A dead simple Go library for sending notifications to various messaging services.
