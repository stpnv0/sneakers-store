package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type OrderCreatedEvent struct {
	EventType   string `json:"event_type"`
	OrderID     int    `json:"order_id"`
	UserID      int    `json:"user_id"`
	TotalAmount int    `json:"total_amount"`
	Timestamp   string `json:"timestamp"`
}

type Consumer struct {
	reader  *kafka.Reader
	log     *slog.Logger
	handler OrderEventHandler
}

type OrderEventHandler interface {
	HandleOrderCreated(ctx context.Context, event OrderCreatedEvent) error
}

func NewConsumer(brokers []string, topic, groupID string, handler OrderEventHandler, log *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:  reader,
		log:     log,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.log.Info("starting kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("stopping kafka consumer")
			return c.reader.Close()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.log.Error("failed to read message", slog.String("error", err.Error()))
				continue
			}

			var event OrderCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.log.Error("failed to unmarshal event", slog.String("error", err.Error()))
				continue
			}

			if event.EventType == "OrderCreated" {
				if err := c.handler.HandleOrderCreated(ctx, event); err != nil {
					c.log.Error("failed to handle order event",
						slog.Int("order_id", event.OrderID),
						slog.String("error", err.Error()))
				} else {
					c.log.Info("processed order event successfully", slog.Int("order_id", event.OrderID))
				}
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
