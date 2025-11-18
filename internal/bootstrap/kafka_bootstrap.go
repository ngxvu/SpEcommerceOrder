package bootstrap

import (
	"context"
	"log"
	"order/internal/services"
	"order/internal/workers"
	"order/pkg/core/configloader"
	"order/pkg/core/kafka"
)

// InitializeKafka returns the app and a stop function to gracefully shut down.
func InitKafka(parent context.Context, orderService services.OrderService) (*kafka.App, func()) {
	cfg := configloader.GetConfig()

	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaPaymentAuthorizedTopic)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaPaymentAuthorizedTopic, "payment_group")

	ctx, cancel := context.WithCancel(parent)

	paymentWorker := workers.NewPaymentEventWorker(&orderService)

	go consumer.Listen(ctx, func(data []byte) {
		paymentWorker.Handle(ctx, data)
	})

	stop := func() {
		cancel()
		if err := producer.Close(); err != nil {
			log.Printf("producer close error: %v", err)
		}
		if err := consumer.Reader.Close(); err != nil {
			log.Printf("consumer close error: %v", err)
		}
	}

	return &kafka.App{
		Producer: producer,
		Consumer: consumer,
	}, stop
}
