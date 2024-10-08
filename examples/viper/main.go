package main

import (
	"errors"
	"log"
	"os"

	"github.com/spf13/viper"

	"github.com/nikoksr/konfetty"
	"github.com/nikoksr/konfetty/examples"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type AppConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	LogLevel string         `mapstructure:"log_level"`
}

func main() {
	cfg := new(AppConfig)
	var err error

	// Viper setup
	v := viper.New()
	v.SetConfigName("app.config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	// Use konfetty to handle the rest
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
	//     "Host": "super-secret-database.myapp.com",  // Kept original value, loaded by Viper
	//     "Port": 5555,                               // Kept original value, loaded by Viper
	//     "Username": "myuser",                       // Kept original value, loaded by Viper
	//     "Password": "mypassword"                    // Kept original value, loaded by Viper
	//   },
	//   "Server": {
	//     "Name": "Procrastination Station",          // Set by transformer function
	//     "Host": "localhost",                        // Kept original value, loaded by Viper
	//     "Port": 8080                                // Applied from defaults
	//   },
	//   "LogLevel": "info"                            // Applied from defaults
	// }
}
