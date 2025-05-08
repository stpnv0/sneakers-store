package services

import (
	"context"
	"errors"
	"payment/internal/domain/models"
	"payment/internal/infrastructure/kafka"
	"payment/internal/infrastructure/yookassa"
	"payment/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInvalidCurrency = errors.New("invalid currency")
)

type PaymentService struct {
	repo           repository.PaymentRepository
	yookassaClient *yookassa.Client
	kafkaProducer  *kafka.Producer
	logger         *zap.Logger
}

func NewPaymentService(
	repo repository.PaymentRepository,
	yookassaClient *yookassa.Client,
	kafkaProducer *kafka.Producer,
	logger *zap.Logger,
) *PaymentService {
	return &PaymentService{
		repo:           repo,
		yookassaClient: yookassaClient,
		kafkaProducer:  kafkaProducer,
		logger:         logger,
	}
}

// InitiatePayment создает новый платеж
func (s *PaymentService) InitiatePayment(ctx context.Context, orderID, userID string, amount float64,
	currency, description, returnURL string, metadata map[string]string, idempotencyKey string) (*models.Payment, error) {

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if currency == "" {
		return nil, ErrInvalidCurrency
	}

	if idempotencyKey != "" {
		existingPayment, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil && existingPayment != nil {
			return existingPayment, nil
		}
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata["order_id"] = orderID
	metadata["user_id"] = userID

	payment := &models.Payment{
		ID:          uuid.New().String(),
		OrderID:     orderID,
		UserID:      userID,
		Amount:      amount,
		Currency:    currency,
		Status:      models.PaymentStatusPending,
		Description: description,
		ReturnURL:   returnURL,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, payment, idempotencyKey); err != nil {
		s.logger.Error("Failed to create payment in DB",
			zap.Error(err),
			zap.String("order_id", orderID))
		return nil, err
	}

	yookassaReq := &yookassa.CreatePaymentRequest{
		Amount: yookassa.Amount{
			Value:    amount,
			Currency: currency,
		},
		Confirmation: yookassa.Confirmation{
			Type:      "redirect",
			ReturnURL: returnURL,
		},
		Description: description,
		Metadata:    metadata,
	}

	yookassaResp, err := s.yookassaClient.CreatePayment(ctx, yookassaReq, idempotencyKey)
	if err != nil {
		s.logger.Error("Failed to create payment in YooKassa",
			zap.Error(err),
			zap.String("payment_id", payment.ID))

		// Обновляем статус на failed
		payment.Status = models.PaymentStatusFailed
		payment.UpdatedAt = time.Now()
		_ = s.repo.Update(ctx, payment)

		return nil, err
	}

	payment.ExternalID = yookassaResp.ID
	payment.PaymentURL = yookassaResp.Confirmation.ConfirmationURL
	payment.Status = models.PaymentStatusWaiting
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		s.logger.Error("Failed to update payment in DB",
			zap.Error(err),
			zap.String("payment_id", payment.ID))
		return nil, err
	}

	event := map[string]interface{}{
		"event_type":  "payment_created",
		"payment_id":  payment.ID,
		"order_id":    payment.OrderID,
		"user_id":     payment.UserID,
		"amount":      payment.Amount,
		"currency":    payment.Currency,
		"status":      payment.Status,
		"external_id": payment.ExternalID,
		"created_at":  payment.CreatedAt,
		"metadata":    payment.Metadata,
	}

	if err := s.kafkaProducer.SendMessage(ctx, "payment_events", payment.OrderID, event); err != nil {
		s.logger.Warn("Failed to send Kafka message",
			zap.Error(err),
			zap.String("payment_id", payment.ID))
	}

	return payment, nil
}

// UpdatePaymentStatusByExternalID обновляет статус платежа по внешнему ID
func (s *PaymentService) UpdatePaymentStatusByExternalID(ctx context.Context, externalID string, status models.PaymentStatus) error {
	payment, err := s.repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return err
	}

	if payment == nil {
		s.logger.Warn("Payment not found by external ID",
			zap.String("external_id", externalID))
		return ErrPaymentNotFound
	}

	// Если статус не изменился, ничего не делаем
	if payment.Status == status {
		return nil
	}

	oldStatus := payment.Status
	payment.Status = status
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return err
	}

	eventType := "payment_status_changed"
	if status == models.PaymentStatusSucceeded {
		eventType = "payment_succeeded"
	} else if status == models.PaymentStatusCanceled || status == models.PaymentStatusFailed {
		eventType = "payment_failed"
	}

	event := map[string]interface{}{
		"event_type":  eventType,
		"payment_id":  payment.ID,
		"order_id":    payment.OrderID,
		"user_id":     payment.UserID,
		"amount":      payment.Amount,
		"currency":    payment.Currency,
		"status":      payment.Status,
		"old_status":  oldStatus,
		"external_id": payment.ExternalID,
		"updated_at":  payment.UpdatedAt,
		"metadata":    payment.Metadata, // Включаем метаданные
	}

	if err := s.kafkaProducer.SendMessage(ctx, "payment_events", payment.OrderID, event); err != nil {
		s.logger.Warn("Failed to send Kafka message",
			zap.Error(err),
			zap.String("payment_id", payment.ID))
	}

	return nil
}

// GetPayment возвращает платеж по ID
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

// GetPaymentByOrderID возвращает платеж по ID заказа
func (s *PaymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	payment, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

// CapturePayment подтверждает платеж (для двухстадийных платежей)
func (s *PaymentService) CapturePayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	if payment.Status != models.PaymentStatusWaiting {
		s.logger.Warn("Cannot capture payment with status other than waiting_for_capture",
			zap.String("payment_id", paymentID),
			zap.String("status", string(payment.Status)))
		return payment, nil
	}

	// Подтверждаем платеж в ЮКассе
	amount := &yookassa.Amount{
		Value:    payment.Amount,
		Currency: payment.Currency,
	}

	_, err = s.yookassaClient.CapturePayment(ctx, payment.ExternalID, amount, payment.ID)
	if err != nil {
		s.logger.Error("Failed to capture payment in YooKassa",
			zap.Error(err),
			zap.String("payment_id", paymentID))
		return nil, err
	}

	payment.Status = models.PaymentStatusSucceeded
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	// Отправляем событие в Kafka
	event := map[string]interface{}{
		"event_type":  "payment_succeeded",
		"payment_id":  payment.ID,
		"order_id":    payment.OrderID,
		"user_id":     payment.UserID,
		"amount":      payment.Amount,
		"currency":    payment.Currency,
		"status":      payment.Status,
		"external_id": payment.ExternalID,
		"updated_at":  payment.UpdatedAt,
		"metadata":    payment.Metadata,
	}

	if err := s.kafkaProducer.SendMessage(ctx, "payment_events", payment.OrderID, event); err != nil {
		s.logger.Warn("Failed to send Kafka message",
			zap.Error(err),
			zap.String("payment_id", payment.ID))
	}

	return payment, nil
}

// CancelPayment отменяет платеж
func (s *PaymentService) CancelPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	if payment.Status == models.PaymentStatusSucceeded || payment.Status == models.PaymentStatusCanceled {
		s.logger.Warn("Cannot cancel payment with status succeeded or canceled",
			zap.String("payment_id", paymentID),
			zap.String("status", string(payment.Status)))
		return payment, nil
	}

	// Отменяем платеж в ЮКассе
	_, err = s.yookassaClient.CancelPayment(ctx, payment.ExternalID, payment.ID)
	if err != nil {
		s.logger.Error("Failed to cancel payment in YooKassa",
			zap.Error(err),
			zap.String("payment_id", paymentID))
		return nil, err
	}

	payment.Status = models.PaymentStatusCanceled
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	event := map[string]interface{}{
		"event_type":  "payment_canceled",
		"payment_id":  payment.ID,
		"order_id":    payment.OrderID,
		"user_id":     payment.UserID,
		"amount":      payment.Amount,
		"currency":    payment.Currency,
		"status":      payment.Status,
		"external_id": payment.ExternalID,
		"updated_at":  payment.UpdatedAt,
		"metadata":    payment.Metadata,
	}

	if err := s.kafkaProducer.SendMessage(ctx, "payment_events", payment.OrderID, event); err != nil {
		s.logger.Warn("Failed to send Kafka message",
			zap.Error(err),
			zap.String("payment_id", payment.ID))
	}

	return payment, nil
}
