package bootstrap

import (
	"context"
	"log"
	"order/internal/events"
	"order/internal/services"
	"order/internal/workers"
	"order/pkg/core/configloader"
	"order/pkg/core/kafka"
	"order/pkg/http/utils/errors"
)

// InitKafka initializes Kafka producers/consumers based on config and wires the payment worker.
// It returns the kafka.App instance and a stop function to gracefully shut down.
func InitKafka(
	parent context.Context,
	orderService services.OrderServiceInterface,
	promotionService services.PromotionServiceInterface) (
	*kafka.App, func() error, error) {
	cfg := configloader.GetConfig()

	ctx, cancel := context.WithCancel(parent)

	app := &kafka.App{
		Producers: make(map[string]*kafka.Producer),
		Consumers: make(map[string]*kafka.Consumer),
	}

	// Set up producers for all configured topics
	for _, topic := range cfg.KafkaTopics {
		prod := kafka.NewProducer(cfg.KafkaBrokers, topic)

		// store in app producers map
		app.Producers[topic] = prod
	}

	// For now, set up consumer only for the payment_authorized topic
	var paymentTopic string
	for _, topic := range cfg.KafkaTopics {
		if paymentTopic != "" {
			err := errors.Error(errors.StatusInternalServerError, errors.StatusValidationError)
			return nil, nil, err
			break
		}
		switch topic {
		case string(events.PaymentAuthorizationTopic):
			c := kafka.NewConsumer(cfg.KafkaBrokers, topic, "payment_group")
			app.Consumers[topic] = c
			w := workers.NewPaymentEventWorker(orderService, app.Producers[string(events.PromotionRewardTopic)])
			go c.Listen(ctx, func(data []byte) { w.Handle(ctx, data) })

		case string(events.PromotionRewardTopic):
			c := kafka.NewConsumer(cfg.KafkaBrokers, topic, "promotion_group")
			app.Consumers[topic] = c
			w := workers.NewPromotionRewardWorker(promotionService)
			go c.Listen(ctx, func(data []byte) { w.Handle(ctx, data) })
		}
	}

	stop := func() error {
		cancel()
		// close all producers
		for topic, p := range app.Producers {
			if err := p.Close(); err != nil {
				log.Printf("producer[%s] close error: %v", topic, err)
				return err
			}
		}
		// close all consumers
		for topic, c := range app.Consumers {
			if err := c.Reader.Close(); err != nil {
				log.Printf("consumer[%s] close error: %v", topic, err)
				return err
			}
		}
		return nil
	}

	return app, stop, nil
}
