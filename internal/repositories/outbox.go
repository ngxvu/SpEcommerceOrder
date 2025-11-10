package repo

import (
	"context"
	"gorm.io/gorm"
	models "order_service/internal/models"
	pgGorm "order_service/internal/repositories/pg-gorm"
)

type OutboxRepository struct {
	db pgGorm.PGInterface
}

func NewOutboxRepository(newPgRepo pgGorm.PGInterface) *OutboxRepository {
	return &OutboxRepository{db: newPgRepo}
}

type OutboxRepoInterface interface {
	CreateOutBox(ctx context.Context, tx *gorm.DB, outboxEvent *models.OutboxEvent) error
}

func (o *OutboxRepository) CreateOutBox(ctx context.Context, tx *gorm.DB, outboxEvent *models.OutboxEvent) error {

	if tx == nil {
		var cancel context.CancelFunc
		tx, cancel = o.db.DBWithTimeout(ctx)
		defer cancel()
	}

	outBoxData := map[string]interface{}{
		"event_type": outboxEvent.EventType,
		"payload":    outboxEvent.Payload,
		"process":    outboxEvent.Process,
	}

	if err := tx.Model(&models.Outbox{}).Create(outBoxData).Error; err != nil {
		return err
	}

	return nil
}
