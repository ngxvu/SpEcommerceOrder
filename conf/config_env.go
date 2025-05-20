package conf

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// AppConfig presents app conf
type AppConfig struct {
	Port     string `env:"PORT"`
	DBHost   string `env:"DB_HOST"`
	DBPort   string `env:"DB_PORT"`
	DBUser   string `env:"DB_USER"`
	DBPass   string `env:"DB_PASS"`
	DBName   string `env:"DB_NAME"`
	EnableDB bool   `env:"ENABLE_DB" envDefault:"true"`

	JWTSecret string `env:"JWT_SECRET"`
	JWTExpire int    `env:"JWT_EXPIRE" envDefault:"24"`
}

var config AppConfig

func SetEnvFromFile(path string) {
	err := godotenv.Load(path)
	if err != nil {
		fmt.Println(err)
	}
	_ = env.Parse(&config)
}

func SetEnv() {
	err := godotenv.Load("./.env")
	if err != nil {
		fmt.Println(err)
	}
	_ = env.Parse(&config)
}

func LoadEnv() AppConfig {
	return config
}
