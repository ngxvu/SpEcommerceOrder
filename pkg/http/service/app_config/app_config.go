package app_config

import (
	"fmt"
	"kimistore/conf"
	repo "kimistore/internal/repo/pg-gorm"
	"kimistore/pkg/http/db"
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
