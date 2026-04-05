package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"order_service/internal/models"
)

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(brokers []string, topic string, log *slog.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &Producer{
		writer: writer,
		log:    log,
	}
}

func (p *Producer) PublishOrderEvent(ctx context.Context, event models.OrderEvent) error {
	const op = "kafka.Producer.PublishOrderEvent"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: marshal event: %w", op, err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", event.OrderID)),
		Value: data,
	}); err != nil {
		return fmt.Errorf("%s: write message: %w", op, err)
	}

	p.log.Info("published order event",
		slog.String("op", op),
		slog.String("event_type", event.EventType),
		slog.Int("order_id", event.OrderID),
	)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
