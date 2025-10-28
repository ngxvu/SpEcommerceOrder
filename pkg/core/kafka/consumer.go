package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
)

type ReaderWrapper interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Consumer struct {
	Reader ReaderWrapper
}

func NewConsumer(brokers, topic, groupID string) *Consumer {
	reader := NewReader(brokers, topic, groupID)
	return &Consumer{Reader: reader}
}

// Adapter implement ReaderWrapper
type KafkaReader struct {
	reader *kafka.Reader
}

func (r *KafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return r.reader.FetchMessage(ctx)
}

func (r *KafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return r.reader.CommitMessages(ctx, msgs...)
}

func (r *KafkaReader) Close() error {
	return r.reader.Close()
}

func NewReader(brokers, topic, groupID string) ReaderWrapper {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return &KafkaReader{reader: r}
}

func (c *Consumer) Listen(ctx context.Context, handler func([]byte)) {
	for {
		msg, err := c.Reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		handler(msg.Value)

		if err := c.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Failed to commit message: %v", err)
		}
	}
}
