package repo

import (
	"context"
	"gorm.io/gorm"
	models "order/internal/models"
	pgGorm "order/internal/repositories/pg-gorm"
)

type OutboxRepository struct {
	db pgGorm.PGInterface
}

func NewOutboxRepository(newPgRepo pgGorm.PGInterface) *OutboxRepository {
	return &OutboxRepository{db: newPgRepo}
}

type OutboxRepoInterface interface {
	CreateOutbox(ctx context.Context, tx *gorm.DB, outbox *models.Outbox) error
}

func (a *OutboxRepository) CreateOutbox(ctx context.Context, tx *gorm.DB, outbox *models.Outbox) error {
	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = a.db.DBWithTimeout(ctx)
		defer cancel()
	}
	if err := tx.Create(outbox).Error; err != nil {
		return err
	}
	return nil
}
