package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

func NewWriter(addr string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func NewReader(addr string, topic string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{addr},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func (c *Client) Publish(ctx context.Context, key []byte, value []byte) error {
	err := c.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func (c *Client) Close() {
	if c.Writer != nil {
		if err := c.Writer.Close(); err != nil {
			log.Printf("failed to close writer: %v", err)
		}
	}
	if c.Reader != nil {
		if err := c.Reader.Close(); err != nil {
			log.Printf("failed to close reader: %v", err)
		}
	}
}
