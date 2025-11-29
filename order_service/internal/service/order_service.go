package service

import (
	"context"
	"fmt"
	"log/slog"
	"order_service/internal/kafka"
	"order_service/internal/models"
	"order_service/internal/repository"
)

type OrderService struct {
	repo     *repository.OrderRepository
	producer *kafka.Producer
	log      *slog.Logger
}

func NewOrderService(repo *repository.OrderRepository, producer *kafka.Producer, log *slog.Logger) *OrderService {
	return &OrderService{
		repo:     repo,
		producer: producer,
		log:      log,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int, items []models.OrderItem) (*models.OrderWithItems, error) {
	var totalAmount int
	for _, item := range items {
		totalAmount += item.PriceAtPurchase * item.Quantity
	}

	order := &models.Order{
		UserID:      userID,
		Status:      models.OrderStatusPendingPayment,
		TotalAmount: totalAmount,
	}

	createdOrder, err := s.repo.Create(ctx, order, items)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if err := s.producer.PublishOrderCreated(ctx, createdOrder.ID, userID, totalAmount); err != nil {
		s.log.Error("failed to publish OrderCreated event", slog.String("error", err.Error()))
		// Note: заказ уже создан в БД, поэтому не нужно откатывать транзакцию
	}

	s.log.Info("order created", slog.Int("order_id", createdOrder.ID), slog.Int("user_id", userID))
	return createdOrder, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	return s.repo.GetByID(ctx, orderID)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error) {
	return s.repo.GetUserOrders(ctx, userID)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	return s.repo.UpdateStatus(ctx, orderID, status)
}

func (s *OrderService) HandlePaymentProcessed(ctx context.Context, event kafka.PaymentProcessedEvent) error {
	var newStatus string
	if event.Status == "SUCCESS" {
		newStatus = models.OrderStatusPaid
	} else if event.Status == "PENDING" {
		newStatus = models.OrderStatusPendingPayment
	} else {
		newStatus = models.OrderStatusPaymentFailed
	}

	err := s.repo.UpdateStatus(ctx, event.OrderID, newStatus)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if event.PaymentURL != "" {
		if err := s.repo.UpdatePaymentURL(ctx, event.OrderID, event.PaymentURL); err != nil {
			s.log.Error("failed to update payment url", slog.String("error", err.Error()))
		}
	}

	s.log.Info("order status updated",
		slog.Int("order_id", event.OrderID),
		slog.String("status", newStatus))
	return nil
}
