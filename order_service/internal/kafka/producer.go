package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderCreatedEvent struct {
	EventType   string `json:"event_type"`
	OrderID     int    `json:"order_id"`
	UserID      int    `json:"user_id"`
	TotalAmount int    `json:"total_amount"`
	Timestamp   string `json:"timestamp"`
}

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(brokers []string, topic string, log *slog.Logger) *Producer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
	})

	return &Producer{
		writer: writer,
		log:    log,
	}
}

func (p *Producer) PublishOrderCreated(ctx context.Context, orderID, userID, totalAmount int) error {
	event := OrderCreatedEvent{
		EventType:   "OrderCreated",
		OrderID:     orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.log.Info("published OrderCreated event", slog.Int("order_id", orderID))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
