package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writers map[string]*kafka.Writer
	brokers []string
	logger  *zap.Logger
}

func NewProducer(brokers []string, logger *zap.Logger) *Producer {
	return &Producer{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
		logger:  logger,
	}
}

func (p *Producer) getWriter(topic string) *kafka.Writer {
	if writer, exists := p.writers[topic]; exists {
		return writer
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      p.brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		// Настройки для высокой нагрузки
		Async:        true,
		RequiredAcks: int(kafka.RequireOne),
	})

	p.writers[topic] = writer
	return writer
}

func (p *Producer) SendMessage(ctx context.Context, topic, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	writer := p.getWriter(topic)

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
		Time:  time.Now(),
	})

	if err != nil {
		p.logger.Error("Failed to send message to Kafka",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("key", key))
		return err
	}

	p.logger.Debug("Message sent to Kafka",
		zap.String("topic", topic),
		zap.String("key", key))

	return nil
}

func (p *Producer) Close() {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Failed to close Kafka writer",
				zap.Error(err),
				zap.String("topic", topic))
		}
	}
}
