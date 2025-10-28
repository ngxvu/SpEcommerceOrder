package bootstrap

import (
	repo "basesource/internal/repositories/pg-gorm"
	"basesource/pkg/core/configloader"
	"basesource/pkg/core/db"
	"fmt"
)

type App struct {
	Config *configloader.Config
	PGRepo repo.PGInterface
}

// initializeApp initializes all application dependencies
func InitializeApp() (*App, error) {
	config := configloader.GetConfig()

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
