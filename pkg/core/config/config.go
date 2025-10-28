package config

import (
	"basesource/conf"
	repo "basesource/internal/repositories/pg-gorm"
	"basesource/pkg/http/db"
	"fmt"
)

type App struct {
	Config *conf.Config
	PGRepo repo.PGInterface
}

// initializeApp initializes all application dependencies
func InitializeApp() (*App, error) {
	config := conf.GetConfig()

	// Initialize database
	dbBackend, err := db.InitDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	pgRepo := repo.NewPGRepo(dbBackend)

	return &App{
		PGRepo: pgRepo,
		Config: config,
	}, nil
}
