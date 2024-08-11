package main

import (
	"errors"
	"log"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/nikoksr/konfetty"
	"github.com/nikoksr/konfetty/examples"
)

type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type ServerConfig struct {
	Name string `koanf:"name"`
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type AppConfig struct {
	Database DatabaseConfig `koanf:"database"`
	Server   ServerConfig   `koanf:"server"`
	LogLevel string         `koanf:"log_level"`
}

func main() {
	cfg := new(AppConfig)
	var err error

	//
	// Your (existing) koanf setup.
	//

	k := koanf.New(".")
	if err := k.Load(file.Provider("./app.config.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := k.Unmarshal("", cfg); err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	//
	// Simply pass the config to konfetty and let it handle the rest.
	//

	cfg, err = konfetty.FromStruct(cfg).
		WithDefaults(
			AppConfig{
				Database: DatabaseConfig{
					Host: "localhost",
					Port: 5432,
				},
				Server: ServerConfig{
					Port: 8080,
				},
				LogLevel: "info",
			},
		).
		WithTransformer(func(c *AppConfig) {
			if c.Server.Host == "localhost" {
				c.Server.Name = "Procrastination Station"
			}
		}).
		WithValidator(func(c *AppConfig) error {
			if c.Database.Username == "" || c.Database.Password == "" {
				return errors.New("database credentials are required")
			}
			return nil
		}).
		Build()
	if err != nil {
		log.Fatalf("Error building config: %v", err)
	}

	// Use the final config as needed...

	examples.PrettyPrint(os.Stdout, cfg)

	// The final config would look like this:
	//
	// {
	//   "Database": {
	//     "Host": "super-secret-database.myapp.com",  // Kept original value, loaded by Koanf
	//     "Port": 5555, 							   // Kept original value, loaded by Koanf
	//     "Username": "myuser",                       // Kept original value, loaded by Koanf
	//     "Password": "mypassword"                    // Kept original value, loaded by Koanf
	//   },
	//   "Server": {
	//     "Name": "Procrastination Station",          // Set by transformer function
	//     "Host": "localhost",                        // Kept original value, loaded by Koanf
	//     "Port": 8080                                // Applied from defaults
	//   },
	//   "LogLevel": "info"                            // Applied from defaults
	// }
}
