package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"order_service/internal/models"
)

type OrderServiceImpl struct {
	repo        OrderRepository
	paymentRepo PaymentRepository
	provider    PaymentProvider
	publisher   EventPublisher
	log         *slog.Logger
}

func NewOrderService(
	repo OrderRepository,
	paymentRepo PaymentRepository,
	provider PaymentProvider,
	publisher EventPublisher,
	log *slog.Logger,
) *OrderServiceImpl {
	return &OrderServiceImpl{
		repo:        repo,
		paymentRepo: paymentRepo,
		provider:    provider,
		publisher:   publisher,
		log:         log,
	}
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, userID int, items []models.OrderItem) (*models.OrderWithItems, error) {
	const op = "service.OrderService.CreateOrder"

	var totalAmount int
	for _, item := range items {
		totalAmount += item.PriceAtPurchase * item.Quantity
	}

	order := &models.Order{
		UserID:      userID,
		Status:      models.OrderStatusPendingPayment,
		TotalAmount: totalAmount,
	}

	created, err := s.repo.Create(ctx, order, items)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Синхронно создаём платёж в YooKassa.
	description := fmt.Sprintf("Order #%d", created.ID)
	providerResp, err := s.provider.CreatePayment(ctx, totalAmount, "RUB", description)
	if err != nil {
		s.log.Error("failed to create payment in provider",
			slog.String("op", op),
			slog.Int("order_id", created.ID),
			slog.String("error", err.Error()),
		)
		return created, nil
	}

	// Сохраняем запись о платеже
	payment := &models.Payment{
		OrderID:           created.ID,
		YooKassaPaymentID: providerResp.ID,
		Amount:            totalAmount,
		Currency:          "RUB",
		Status:            models.PaymentStatusPending,
		ConfirmationURL:   providerResp.ConfirmationURL,
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		s.log.Error("failed to save payment record",
			slog.String("op", op),
			slog.Int("order_id", created.ID),
			slog.String("error", err.Error()),
		)
	}

	// Обновляем payment_url в заказе
	created.PaymentURL = providerResp.ConfirmationURL
	if err := s.repo.UpdatePaymentURL(ctx, created.ID, providerResp.ConfirmationURL); err != nil {
		s.log.Error("failed to update payment url",
			slog.String("op", op),
			slog.Int("order_id", created.ID),
			slog.String("error", err.Error()),
		)
	}

	s.publishEvent(ctx, op, models.OrderEvent{
		EventType:   "OrderCreated",
		OrderID:     created.ID,
		UserID:      userID,
		Status:      created.Status,
		TotalAmount: totalAmount,
		PaymentURL:  providerResp.ConfirmationURL,
		Timestamp:   time.Now().Format(time.RFC3339),
	})

	s.log.Info("order created",
		slog.String("op", op),
		slog.Int("order_id", created.ID),
		slog.Int("user_id", userID),
		slog.Int("total_amount", totalAmount),
	)

	return created, nil
}

func (s *OrderServiceImpl) GetOrder(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	const op = "service.OrderService.GetOrder"

	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return order, nil
}

func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error) {
	const op = "service.OrderService.GetUserOrders"

	orders, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return orders, nil
}

func (s *OrderServiceImpl) UpdateOrderStatus(ctx context.Context, orderID int, newStatus string) error {
	const op = "service.OrderService.UpdateOrderStatus"

	if !models.IsValidStatus(newStatus) {
		return fmt.Errorf("%s: invalid order status %q", op, newStatus)
	}

	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("%s: get order: %w", op, err)
	}

	if !models.ValidTransition(order.Status, newStatus) {
		return fmt.Errorf("%s: invalid transition from %q to %q", op, order.Status, newStatus)
	}

	if err := s.repo.UpdateStatus(ctx, orderID, newStatus, order.Status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *OrderServiceImpl) ProcessWebhook(ctx context.Context, yookassaID, status string) error {
	const op = "service.OrderService.ProcessWebhook"

	s.log.Info("processing webhook",
		slog.String("op", op),
		slog.String("payment_id", yookassaID),
		slog.String("status", status),
	)

	var internalPaymentStatus string
	switch status {
	case "succeeded":
		internalPaymentStatus = models.PaymentStatusSucceeded
	case "canceled":
		internalPaymentStatus = models.PaymentStatusCanceled
	default:
		internalPaymentStatus = models.PaymentStatusPending
	}

	// Проверяем валидность перехода.
	existing, err := s.paymentRepo.GetByYooKassaID(ctx, yookassaID)
	if err != nil {
		return fmt.Errorf("%s: get payment: %w", op, err)
	}

	if !models.ValidPaymentTransition(existing.Status, internalPaymentStatus) {
		s.log.Warn("skipping invalid payment transition",
			slog.String("op", op),
			slog.String("payment_id", yookassaID),
			slog.String("from", existing.Status),
			slog.String("to", internalPaymentStatus),
		)
		return nil
	}

	payment, err := s.paymentRepo.UpdateStatusAndGet(ctx, yookassaID, internalPaymentStatus)
	if err != nil {
		return fmt.Errorf("%s: update payment status: %w", op, err)
	}
	if payment == nil {
		s.log.Info("webhook already processed", slog.String("op", op), slog.String("payment_id", yookassaID))
		return nil
	}

	var orderStatus string
	switch internalPaymentStatus {
	case models.PaymentStatusSucceeded:
		orderStatus = models.OrderStatusPaid
	default:
		orderStatus = models.OrderStatusPaymentFailed
	}

	if err := s.UpdateOrderStatus(ctx, payment.OrderID, orderStatus); err != nil {
		s.log.Error("failed to update order status after payment",
			slog.String("op", op),
			slog.Int("order_id", payment.OrderID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%s: update order status: %w", op, err)
	}

	s.publishEvent(ctx, op, models.OrderEvent{
		EventType: "OrderPaymentUpdated",
		OrderID:   payment.OrderID,
		Status:    orderStatus,
		Timestamp: time.Now().Format(time.RFC3339),
	})

	s.log.Info("order status updated after payment",
		slog.String("op", op),
		slog.Int("order_id", payment.OrderID),
		slog.String("new_status", orderStatus),
	)

	return nil
}

func (s *OrderServiceImpl) publishEvent(ctx context.Context, op string, event models.OrderEvent) {
	if err := s.publisher.PublishOrderEvent(ctx, event); err != nil {
		s.log.Error("failed to publish event",
			slog.String("op", op),
			slog.String("event_type", event.EventType),
			slog.String("error", err.Error()),
		)
	}
}
