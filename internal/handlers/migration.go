package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	model "kimistore/internal/models"
	repo "kimistore/internal/repo/pg-gorm"
	"kimistore/pkg/http/logger"
	"net/http"
)

type MigrationHandler struct {
	newRepo repo.PGInterface
}

func NewMigrationHandler(newRepo repo.PGInterface) *MigrationHandler {
	return &MigrationHandler{newRepo: newRepo}
}

func (m *MigrationHandler) Migrate(ctx *gin.Context) {
	m.MigrateCmdPublic(ctx)
}

func (m *MigrationHandler) BaseMigratePublic(ctx *gin.Context, tx *gorm.DB) error {
	log := logger.WithCtx(ctx, "BaseMigratePublic")

	sqlCommands := []string{
		`CREATE SCHEMA IF NOT EXISTS public`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
	}

	for _, sql := range sqlCommands {
		if err := tx.Exec(sql).Error; err != nil {
			log.Errorf(err.Error())
			tx.Rollback()
			return err
		}
	}

	models := []interface{}{
		&model.User{},
		&model.Post{},
		&model.Product{},
	}

	tx.Config.NamingStrategy = schema.NamingStrategy{
		TablePrefix: "public.",
	}

	if err := tx.AutoMigrate(models...); err != nil {
		log.Errorf(err.Error())
		tx.Rollback()
		return err
	}

	return nil
}

func (m *MigrationHandler) MigrateCmdPublic(ctx *gin.Context) {

	tx := m.newRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := m.newRepo.DBWithTimeout(ctx)
	defer cancel()

	migrate := gormigrate.New(tx, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20220523172948",
			Migrate: func(tx *gorm.DB) error {
				return m.BaseMigratePublic(ctx, tx)
			},
		},
	})

	if err := migrate.Migrate(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	ctx.JSON(http.StatusOK, gin.H{"message": "Migration completed successfully"})
}
