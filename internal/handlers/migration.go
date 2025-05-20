package handlers

import (
	"emission/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type MigrationHandler struct {
	db *gorm.DB
}

func NewMigrationHandler(db *gorm.DB) *MigrationHandler {
	return &MigrationHandler{db: db}
}

func (h *MigrationHandler) Migrate(ctx *gin.Context) {
	migrate := gormigrate.New(h.db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20220523172948",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.Exec(`
						CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
				`).Error; err != nil {
					return err
				}
				if err := tx.AutoMigrate(
					&models.Factory{},
					&models.EmissionFactor{},
					&models.Emission{},
				); err != nil {
					return err
				}
				return nil
			},
			Rollback: func(db *gorm.DB) error {
				if err := db.Exec(`
					DROP TABLE IF EXISTS emissions;
					DROP TABLE IF EXISTS emission_factors;
					DROP TABLE IF EXISTS factories;
				`).Error; err != nil {
					return err
				}
				return nil
			},
		},
	})
	err := migrate.Migrate()
	if err != nil {
		panic(err)
	}
}
