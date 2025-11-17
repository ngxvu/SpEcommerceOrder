package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Writer *kafka.Writer
}

func NewProducer(brokers, topic string) *Producer {
	return &Producer{
		Writer: NewWriter(brokers, topic),
	}
}

func (p *Producer) SendMessage(ctx context.Context, key, value string) error {
	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	}
	return p.Writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	if p.Writer == nil {
		return nil
	}
	return p.Writer.Close()
}
