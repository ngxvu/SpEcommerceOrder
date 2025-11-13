package bootstrap

import (
	"fmt"
	repo "order/internal/repositories/pg-gorm"
	"order/pkg/core/configloader"
	"order/pkg/core/db"
)

type App struct {
	Config *configloader.Config
	PGRepo repo.PGInterface
}

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
