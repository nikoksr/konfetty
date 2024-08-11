package main

import (
	"log"
	"os"
	"time"

	"github.com/nikoksr/konfetty"
	"github.com/nikoksr/konfetty/examples"
)

type SmartHomeConfig struct {
	General  GeneralConfig
	Rooms    []RoomConfig
	Routines []RoutineConfig
}

type GeneralConfig struct {
	HomeName            string
	OwnerName           string
	TimeZone            string
	EnergyMode          string
	MaintenanceInterval time.Duration
}

type BaseDevice struct {
	Name     string
	Type     string
	Enabled  bool
	Location string
}

type LightDevice struct {
	BaseDevice
	Brightness int
	ColorTemp  int
}

type ThermostatDevice struct {
	BaseDevice
	TargetTemp float64
	Mode       string
}

type RoomConfig struct {
	Name    string
	Devices []any // Can be LightDevice or ThermostatDevice
}

type RoutineConfig struct {
	Name      string
	StartTime string
	Days      []string
	Actions   []Action
}

type Action struct {
	DeviceName string
	Command    string
	Value      any
}

func main() {
	// Initial config (simulating loaded from file)
	cfg := &SmartHomeConfig{
		General: GeneralConfig{
			HomeName:  "My Smart Home",
			OwnerName: "John Doe",
		},
		Rooms: []RoomConfig{
			{
				Name: "Living Room",
				Devices: []any{
					LightDevice{
						BaseDevice: BaseDevice{Name: "Main Light", Type: "light"},
						Brightness: 80,
					},
					ThermostatDevice{
						BaseDevice: BaseDevice{Name: "AC", Type: "thermostat"},
					},
				},
			},
		},
		Routines: []RoutineConfig{
			{
				Name:      "Morning Routine",
				StartTime: "07:00",
				Days:      []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
				Actions: []Action{
					{DeviceName: "Main Light", Command: "setBrightness", Value: 100},
				},
			},
		},
	}

	// Use konfetty to process the config
	cfg, err := konfetty.FromStruct(cfg).
		WithDefaults(
			// Default for all BaseDevice instances
			BaseDevice{
				Enabled:  false,
				Location: "Unknown",
			},
			// Specific defaults for LightDevice instances
			LightDevice{
				BaseDevice: BaseDevice{
					Type: "light", // Overwrite the default for BaseDevice.Type for LightDevice instances
				},
				Brightness: 50,
				ColorTemp:  3000,
			},
			// Specific defaults for ThermostatDevice instances
			ThermostatDevice{
				BaseDevice: BaseDevice{
					Enabled: true,         // All ThermostatDevice instances are enabled by default
					Type:    "thermostat", // Overwrite the default for BaseDevice.Type for ThermostatDevice instances
				},
				TargetTemp: 22.0,
				Mode:       "auto",
			},
			// General config defaults
			GeneralConfig{
				TimeZone:            "UTC",
				EnergyMode:          "balanced",
				MaintenanceInterval: 30 * 24 * time.Hour, // 30 days
			},
		).
		WithTransformer(func(c *SmartHomeConfig) {
			// Example transformation: Set all lights to 20% brightness if EnergyMode is "saving"

			if c.General.EnergyMode != "saving" {
				return
			}

			for i, room := range c.Rooms {
				for j, device := range room.Devices {
					if light, ok := device.(LightDevice); ok {
						light.Brightness = 20
						c.Rooms[i].Devices[j] = light
					}
				}
			}
		}).
		Build()
	if err != nil {
		log.Fatalf("Error building config: %v", err)
	}

	// Use the final config as needed...

	examples.PrettyPrint(os.Stdout, cfg)

	// The processed config would look like this:
	//
	// {
	//   "General": {
	//     "HomeName": "My Smart Home",           // Kept original value
	//     "OwnerName": "John Doe",               // Kept original value
	//     "TimeZone": "UTC",                     // Applied from defaults
	//     "EnergyMode": "balanced",              // Applied from defaults
	//     "MaintenanceInterval": 2592000000000000 // Applied from defaults (30 days in nanoseconds)
	//   },
	//   "Rooms": [
	//     {
	//       "Name": "Living Room",               // Kept original value
	//       "Devices": [
	//         {
	//           "Name": "Main Light",            // Kept original value
	//           "Type": "light",                 // Applied from LightDevice default
	//           "Enabled": false,                // Applied from BaseDevice default
	//           "Location": "Unknown",           // Applied from BaseDevice default
	//           "Brightness": 80,                // Kept original value
	//           "ColorTemp": 3000                // Applied from LightDevice default
	//         },
	//         {
	//           "Name": "AC",                    // Kept original value
	//           "Type": "thermostat",            // Applied from ThermostatDevice default
	//           "Enabled": true,                 // Applied from ThermostatDevice default
	//           "Location": "Unknown",           // Applied from BaseDevice default
	//           "TargetTemp": 22,                // Applied from ThermostatDevice default
	//           "Mode": "auto"                   // Applied from ThermostatDevice default
	//         }
	//       ]
	//     }
	//   ],
	//   "Routines": [
	//     {
	//       "Name": "Morning Routine",           // Kept original value
	//       "StartTime": "07:00",                // Kept original value
	//       "Days": [
	//         "Monday",
	//         "Tuesday",
	//         "Wednesday",
	//         "Thursday",
	//         "Friday"
	//       ],                                   // Kept original values
	//       "Actions": [
	//         {
	//           "DeviceName": "Main Light",      // Kept original value
	//           "Command": "setBrightness",      // Kept original value
	//           "Value": 100                     // Kept original value
	//         }
	//       ]
	//     }
	//   ]
	// }
}
