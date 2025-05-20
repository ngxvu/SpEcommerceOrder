package service

import (
	"emission/pkg/http/db"
)

// AppConfig presents some basic app configuration
type AppConfig struct {
	Port        int    `env:"PORT" envDefault:"8088"`
	Env         string `env:"ENV" envDefault:"stg"`
	DebugPort   int    `env:"DEBUG_PORT" envDefault:"7070"`
	ReadTimeout int    `env:"READ_TIMEOUT" envDefault:"15"`
	EnableDB    bool   `env:"ENABLE_DB" envDefault:"true"`
	Debug       bool   `env:"DEBUG" envDefault:"false"`
	DB          *db.Config
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		DB: &db.Config{},
	}
}
