package workers

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	paymentclient "order/internal/clients/payment"
	model "order/internal/models"
	repo "order/internal/repositories/pg-gorm"
	pbPayment "order/pkg/proto/paymentpb"
	"time"
)

type OutboxWorker struct {
	pg       repo.PGInterface
	payment  paymentclient.PaymentClient
	interval time.Duration
	limit    int
}

func NewOutboxWorker(pg repo.PGInterface, pay paymentclient.PaymentClient) *OutboxWorker {
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

	tracer := otel.Tracer("order/outbox-worker")
	ctx, span := tracer.Start(ctx, "OutboxWorker.processBatch",
		trace.WithAttributes(attribute.Int("limit", w.limit)))
	defer span.End()

	tx, cancel := w.pg.DBWithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	now := time.Now()

	var outs []model.Outbox
	// select PENDING or RETRY rows that are due
	if err := tx.Where("status IN ? AND next_attempt_at <= ? AND attempts <= ?",
		[]model.OutboxStatus{model.OutboxStatusPending, model.OutboxStatusRetry},
		now,
		model.NumberOfAttempts).
		Order("next_attempt_at").
		Limit(w.limit).
		Find(&outs).Error; err != nil {
		return err
	}

	for _, o := range outs {

		msgCtx, msgSpan := tracer.Start(ctx, "OutboxWorker.processMessage",
			trace.WithAttributes(attribute.String("event_id", o.EventID.String()), attribute.String("event_type", o.EventType)))

		// attempt delivery in transaction to handle concurrent workers safely
		if err := tx.Transaction(func(tx *gorm.DB) error {
			// reload row FOR UPDATE
			var row model.Outbox
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", o.ID).First(&row).Error; err != nil {
				return err
			}
			// skip if already done by another worker
			if row.Status == model.OutboxStatusDone {
				return nil
			}

			// unmarshal payload into PayRequest
			var payReq pbPayment.PayRequest
			if err := json.Unmarshal([]byte(row.Payload), &payReq); err != nil {

				msgSpan.RecordError(err)
				msgSpan.SetStatus(codes.Error, "invalid payload")
				msgSpan.End()

				// mark failed permanently if payload invalid
				row.Status = model.OutboxStatusFailed
				now := time.Now()
				row.ProcessedAt = &now
				return tx.Save(&row).Error

			}

			// set EventID in request for idempotency
			payReq.EventId = row.EventID.String()

			headers := map[string]string{}
			otel.GetTextMapPropagator().Inject(msgCtx, propagation.MapCarrier(headers))
			md := metadata.New(headers)
			msgCtx = metadata.NewOutgoingContext(msgCtx, md)

			// call payment service
			_, err := w.payment.Pay(ctx, &payReq)
			if err == nil {
				msgSpan.RecordError(err)
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
		msgSpan.End()
	}
	return nil
}
