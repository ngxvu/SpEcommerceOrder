package workers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	model "order/internal/models"
	"order/internal/repositories/pg-gorm"
	pbPayment "order/pkg/proto/paymentpb"

	"gorm.io/gorm"
)

type OutboxWorker struct {
	pg       pgGorm.PGInterface
	payment  paymentclient.PaymentClient
	interval time.Duration
	limit    int
}

func NewOutboxWorker(pg pgGorm.PGInterface, pay paymentclient.PaymentClient) *OutboxWorker {
	return &OutboxWorker{
		pg:       pg,
		payment:  pay,
		interval: 5 * time.Second,
		limit:    10,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				log.Printf("outbox worker error: %v", err)
			}
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) error {
	db := w.pg.GetRepo()
	now := time.Now()

	var outs []model.Outbox
	// select pending or retry rows that are due
	if err := db.Where("status IN ? AND next_attempt_at <= ?", []model.OutboxStatus{model.OutboxStatusPending, model.OutboxStatusRetry}, now).
		Order("next_attempt_at").
		Limit(w.limit).
		Find(&outs).Error; err != nil {
		return err
	}

	for _, o := range outs {
		// attempt delivery in transaction to handle concurrent workers safely
		if err := db.Transaction(func(tx *gorm.DB) error {
			// reload row FOR UPDATE
			var row model.Outbox
			if err := tx.Clauses(gorm.Locking{Strength: "UPDATE"}).Where("id = ?", o.ID).First(&row).Error; err != nil {
				return err
			}
			// skip if already done by another worker
			if row.Status == model.OutboxStatusDone {
				return nil
			}

			// unmarshal payload into PayRequest
			var payReq pbPayment.PayRequest
			if err := json.Unmarshal([]byte(row.Payload), &payReq); err != nil {
				// mark failed permanently if payload invalid
				row.Status = model.OutboxStatusFailed
				now := time.Now()
				row.ProcessedAt = &now
				return tx.Save(&row).Error
			}

			// call payment service
			_, err := w.payment.Pay(ctx, &payReq)
			if err == nil {
				row.Status = model.OutboxStatusDone
				now := time.Now()
				row.ProcessedAt = &now
				return tx.Save(&row).Error
			}

			// on failure, increment attempts and schedule retry with backoff
			row.Attempts++
			row.Status = model.OutboxStatusRetry
			backoff := time.Duration(row.Attempts*row.Attempts) * time.Second // simple quadratic backoff
			row.NextAttemptAt = time.Now().Add(backoff)
			return tx.Save(&row).Error
		}); err != nil {
			// log and continue with other rows
			log.Printf("failed to process outbox %s: %v", o.ID, err)
		}
	}

	return nil
}
