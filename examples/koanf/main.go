package main

import (
	"errors"
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/nikoksr/konfetty"
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
	if err := k.Load(file.Provider("../app.config.yaml"), yaml.Parser()); err != nil {
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

	// The final config would look like this:
	//
	// 	{
	//   "Database": {
	//     "Host": "super-secret-database.myapp.com",
	//     "Port": 5555,
	//     "Username": "myuser",
	//     "Password": "mypassword"
	//   },
	//   "Server": {
	//     "Name": "Procrastination Station",
	//     "Host": "localhost",
	//     "Port": 8080
	//   },
	//   "LogLevel": "info"
	// }
}
