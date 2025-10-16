package conf

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"sync"
)

type Config struct {
	// Database configs
	PgUser     string `env:"POSTGRES_USER"`
	PgPassword string `env:"POSTGRES_PASSWORD"`
	PgHost     string `env:"POSTGRES_HOST"`
	PgPort     string `env:"POSTGRES_PORT" envDefault:"5432"`
	PgDatabase string `env:"POSTGRES_DATABASE"`

	// Server configs
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`

	// JWT Security configs
	JWTAccessSecure     string `env:"JWT_ACCESS_SECURE"`
	JWTRefreshSecure    string `env:"JWT_REFRESH_SECURE"`
	JWTAccessTimeMinute string `env:"JWT_ACCESS_TIME_MINUTE" envDefault:"15"`
	JWTRefreshTimeHour  string `env:"JWT_REFRESH_TIME_HOUR" envDefault:"168"`
}

var (
	configOnce sync.Once
	config     Config
)

func GetConfig() *Config {
	configOnce.Do(func() {
		err := godotenv.Load("./.env")
		if err != nil {
			fmt.Println("Error loading .env file:", err)
		}

		if err := env.Parse(&config); err != nil {
			fmt.Printf("Failed to parse environment variables: %v\n", err)
		}
	})
	return &config
}
