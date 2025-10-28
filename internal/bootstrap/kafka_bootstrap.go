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

func InitializeKafka() *KafkaApp {
	cfg := configloader.GetConfig()

	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicOrder)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopicOrder, "order_group")

	// Ví dụ chạy consumer ở goroutine riêng
	go consumer.Listen(context.Background(), func(data []byte) {
		log.Printf("Received: %s", string(data))
	})

	return &KafkaApp{
		Producer: producer,
		Consumer: consumer,
	}
}
