package kafka

import (
	"context"
	"log"
)

type Consumer struct {
	Reader *ReaderWrapper
}

type ReaderWrapper interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	Close() error
}

func NewConsumer(brokers, topic, groupID string) *Consumer {
	reader := NewReader(brokers, topic, groupID)
	return &Consumer{Reader: reader}
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
