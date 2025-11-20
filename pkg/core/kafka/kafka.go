package kafka

import (
	"github.com/segmentio/kafka-go"
	"strings"
)

type App struct {
	// Producers keyed by topic name (or logical name)
	Producers map[string]*Producer
	// Consumers keyed by topic name (or logical name)
	Consumers map[string]*Consumer
}

func NewWriter(brokers, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(brokers, ",")...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func Close(writer *kafka.Writer, reader *kafka.Reader) {
	if writer != nil {
		writer.Close()
	}
	if reader != nil {
		reader.Close()
	}
}
