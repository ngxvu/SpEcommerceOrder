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
	CreateOutBox(ctx context.Context, tx *gorm.DB, outboxEvent *models.OutboxEvent) error
}

func (o *OutboxRepository) CreateOutBox(ctx context.Context, tx *gorm.DB, outboxEvent *models.OutboxEvent) error {

	if tx == nil {
		var cancel context.CancelFunc
		tx, cancel = o.db.DBWithTimeout(ctx)
		defer cancel()
	}

	outBoxData := map[string]interface{}{
		"event_type":      outboxEvent.EventType,
		"payload":         outboxEvent.Payload,
		"process":         outboxEvent.Process,
		"aggregate_type":  outboxEvent.AggregateType,
		"aggregate_id":    outboxEvent.AggregateID,
		"status":          outboxEvent.Status,
		"attempts":        outboxEvent.Attempts,
		"next_attempt_at": outboxEvent.NextAttemptAt,
		"processed_at":    outboxEvent.ProcessedAt,
	}

	if err := tx.Model(&models.Outbox{}).Create(outBoxData).Error; err != nil {
		return err
	}

	return nil
}
