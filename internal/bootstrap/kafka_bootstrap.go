package bootstrap

import (
	"context"
	"log"
	"order/pkg/core/configloader"
	"order/pkg/core/kafka"
)

// InitializeKafka returns the app and a stop function to gracefully shut down.
func InitKafka(parent context.Context) (*kafka.App, func()) {
	cfg := configloader.GetConfig()

	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicOrder)
	//consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopicOrder, "order_group")
	//
	//ctx, cancel := context.WithCancel(parent)
	//
	//go consumer.Listen(ctx, func(data []byte) {
	//	log.Printf("Received: %s", string(data))
	//})

	stop := func() {
		//cancel()
		if err := producer.Close(); err != nil {
			log.Printf("producer close error: %v", err)
		}
		//if err := consumer.Reader.Close(); err != nil {
		//	log.Printf("consumer close error: %v", err)
		//}
	}

	return &kafka.App{
		Producer: producer,
		//Consumer: consumer,
	}, stop
}
