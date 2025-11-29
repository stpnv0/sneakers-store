package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/segmentio/kafka-go"
)

const (
	MaxRetries     = 3
	RetryCountKey  = "retry_count"
	DLQTopicSuffix = ".dlq"
)

type PaymentProcessedEvent struct {
	EventType  string `json:"event_type"`
	OrderID    int    `json:"order_id"`
	Status     string `json:"status"` // SUCCESS/FAILURE
	PaymentID  string `json:"payment_id"`
	PaymentURL string `json:"payment_url"`
	Timestamp  string `json:"timestamp"`
}

type Consumer struct {
	reader     *kafka.Reader
	dlqWriter  *kafka.Writer
	log        *slog.Logger
	handler    PaymentEventHandler
	maxRetries int
}

type PaymentEventHandler interface {
	HandlePaymentProcessed(ctx context.Context, event PaymentProcessedEvent) error
}

func NewConsumer(brokers []string, topic, groupID string, handler PaymentEventHandler, log *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	// DLQ writer
	dlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic + DLQTopicSuffix,
		Balancer: &kafka.LeastBytes{},
	})

	return &Consumer{
		reader:     reader,
		dlqWriter:  dlqWriter,
		log:        log,
		handler:    handler,
		maxRetries: MaxRetries,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.log.Info("starting kafka consumer with DLQ support")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("stopping kafka consumer")
			c.reader.Close()
			return c.dlqWriter.Close()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.log.Error("failed to read message", slog.String("error", err.Error()))
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	retryCount := c.getRetryCount(msg)

	var event PaymentProcessedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.log.Error("failed to unmarshal event", slog.String("error", err.Error()))
		c.sendToDLQ(ctx, msg, fmt.Sprintf("unmarshal error: %v", err))
		return
	}

	if event.EventType != "PaymentProcessed" {
		return
	}

	err := c.handler.HandlePaymentProcessed(ctx, event)
	if err != nil {
		c.log.Error("failed to handle payment event",
			slog.Int("order_id", event.OrderID),
			slog.Int("retry_count", retryCount),
			slog.String("error", err.Error()))

		// ретраи
		if retryCount < c.maxRetries {
			c.log.Info("retrying message", slog.Int("retry_count", retryCount+1))
			//TODO: можно добавить задержку
		} else {
			//Отправка в DLQ
			c.log.Warn("max retries exceeded, sending to DLQ",
				slog.Int("order_id", event.OrderID),
				slog.Int("retry_count", retryCount))
			c.sendToDLQ(ctx, msg, fmt.Sprintf("max retries exceeded: %v", err))
		}
		return
	}

	c.log.Info("processed payment event successfully",
		slog.Int("order_id", event.OrderID),
		slog.Int("retry_count", retryCount))
}

func (c *Consumer) getRetryCount(msg kafka.Message) int {
	for _, h := range msg.Headers {
		if h.Key == RetryCountKey {
			count, err := strconv.Atoi(string(h.Value))
			if err == nil {
				return count
			}
		}
	}
	return 0
}

func (c *Consumer) sendToDLQ(ctx context.Context, originalMsg kafka.Message, reason string) {
	retryCount := c.getRetryCount(originalMsg)

	// сообщение DLQ с метаданными
	dlqMsg := kafka.Message{
		Key:   originalMsg.Key,
		Value: originalMsg.Value,
		Headers: []kafka.Header{
			{Key: "original_topic", Value: []byte(originalMsg.Topic)},
			{Key: "original_partition", Value: []byte(fmt.Sprintf("%d", originalMsg.Partition))},
			{Key: "original_offset", Value: []byte(fmt.Sprintf("%d", originalMsg.Offset))},
			{Key: "retry_count", Value: []byte(fmt.Sprintf("%d", retryCount))},
			{Key: "failure_reason", Value: []byte(reason)},
			{Key: "timestamp", Value: []byte(originalMsg.Time.Format("2006-01-02T15:04:05Z07:00"))},
		},
	}

	err := c.dlqWriter.WriteMessages(ctx, dlqMsg)
	if err != nil {
		c.log.Error("failed to send message to DLQ",
			slog.String("error", err.Error()),
			slog.String("reason", reason))
	} else {
		c.log.Info("message sent to DLQ", slog.String("reason", reason))
	}
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.dlqWriter.Close()
}
