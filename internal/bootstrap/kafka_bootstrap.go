package bootstrap

import (
	"context"
	"log"
	"order/pkg/core/configloader"
	"order/pkg/core/kafka"
)

type KafkaApp struct {
	Producer *kafka.Producer
	Consumer *kafka.Consumer
}

// InitializeKafka returns the app and a stop function to gracefully shut down.
func InitializeKafka(parent context.Context) (*KafkaApp, func()) {
	cfg := configloader.GetConfig()

	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicOrder)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopicOrder, "order_group")

	ctx, cancel := context.WithCancel(parent)

	go consumer.Listen(ctx, func(data []byte) {
		log.Printf("Received: %s", string(data))
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

	return &KafkaApp{
		Producer: producer,
		Consumer: consumer,
	}, stop
}
