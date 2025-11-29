package service

import (
	"context"
	"fmt"
	"log/slog"
	"payment_service/internal/kafka"
	"payment_service/internal/models"
	"payment_service/internal/provider"
	"payment_service/internal/repository"
)

type PaymentService struct {
	repo     *repository.PaymentRepository
	provider *provider.YooKassaProvider
	producer *kafka.Producer
	log      *slog.Logger
}

func NewPaymentService(repo *repository.PaymentRepository, provider *provider.YooKassaProvider, producer *kafka.Producer, log *slog.Logger) *PaymentService {
	return &PaymentService{
		repo:     repo,
		provider: provider,
		producer: producer,
		log:      log,
	}
}

// HandleOrderCreated implements kafka.OrderEventHandler
func (s *PaymentService) HandleOrderCreated(ctx context.Context, event kafka.OrderCreatedEvent) error {
	s.log.Info("handling order created event", slog.Int("order_id", event.OrderID))

	// Create payment in YooKassa
	description := fmt.Sprintf("Order #%d", event.OrderID)
	yooPayment, err := s.provider.CreatePayment(ctx, event.TotalAmount, "RUB", description)
	if err != nil {
		return fmt.Errorf("failed to create payment in provider: %w", err)
	}

	// Save to DB
	payment := &models.Payment{
		OrderID:           event.OrderID,
		YooKassaPaymentID: yooPayment.ID,
		Amount:            event.TotalAmount,
		Currency:          "RUB",
		Status:            models.PaymentStatusPending,
		ConfirmationURL:   yooPayment.Confirmation.ConfirmationURL,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	s.log.Info("payment created",
		slog.Int("order_id", event.OrderID),
		slog.String("payment_id", yooPayment.ID),
		slog.String("url", yooPayment.Confirmation.ConfirmationURL))

	// Publish event to Kafka with PENDING status and URL so order service can update the URL
	if err := s.producer.PublishPaymentProcessed(ctx, event.OrderID, "PENDING", yooPayment.ID, yooPayment.Confirmation.ConfirmationURL); err != nil {
		s.log.Error("failed to publish payment created event", slog.String("error", err.Error()))
		// Don't fail the whole process as payment is created
	}

	return nil
}

// ProcessWebhook handles YooKassa webhook
func (s *PaymentService) ProcessWebhook(ctx context.Context, yookassaID, status string) error {
	s.log.Info("processing webhook", slog.String("payment_id", yookassaID), slog.String("status", status))

	// Update status in DB
	var internalStatus string
	switch status {
	case "succeeded":
		internalStatus = models.PaymentStatusSucceeded
	case "canceled":
		internalStatus = models.PaymentStatusCanceled
	default:
		internalStatus = models.PaymentStatusPending
	}

	if err := s.repo.UpdateStatus(ctx, yookassaID, internalStatus); err != nil {
		return fmt.Errorf("failed to update status in db: %w", err)
	}

	// Get payment to find order_id
	payment, err := s.repo.GetByYooKassaID(ctx, yookassaID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Publish event to Kafka
	kafkaStatus := "FAILURE"
	if internalStatus == models.PaymentStatusSucceeded {
		kafkaStatus = "SUCCESS"
	}

	if err := s.producer.PublishPaymentProcessed(ctx, payment.OrderID, kafkaStatus, yookassaID, payment.ConfirmationURL); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
