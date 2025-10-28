package kafka

import (
	"github.com/segmentio/kafka-go"
	"strings"
)

func NewWriter(brokers, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(brokers, ",")...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func NewReader(brokers, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(brokers, ","),
		Topic:   topic,
		GroupID: groupID,
	})
}

func Close(writer *kafka.Writer, reader *kafka.Reader) {
	if writer != nil {
		writer.Close()
	}
	if reader != nil {
		reader.Close()
	}
}
