package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentProcessedEvent struct {
	EventType  string `json:"event_type"`
	OrderID    int    `json:"order_id"`
	Status     string `json:"status"` // SUCCESS or FAILURE
	PaymentID  string `json:"payment_id"`
	PaymentURL string `json:"payment_url,omitempty"`
	Timestamp  string `json:"timestamp"`
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

func (p *Producer) PublishPaymentProcessed(ctx context.Context, orderID int, status, paymentID, paymentURL string) error {
	event := PaymentProcessedEvent{
		EventType:  "PaymentProcessed",
		OrderID:    orderID,
		Status:     status,
		PaymentID:  paymentID,
		PaymentURL: paymentURL,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("payment-%d", orderID)),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.log.Info("published PaymentProcessed event",
		slog.Int("order_id", orderID),
		slog.String("status", status))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
