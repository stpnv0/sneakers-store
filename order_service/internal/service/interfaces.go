package service

import (
	"context"

	"order_service/internal/models"
)

//go:generate mockery --name=OrderRepository --output=mocks --outpkg=mocks --filename=mock_order_repository.go
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.OrderWithItems, error)
	GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error)
	GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error)
	UpdateStatus(ctx context.Context, orderID int, newStatus, expectedCurrentStatus string) error
	UpdatePaymentURL(ctx context.Context, orderID int, paymentURL string) error
}

//go:generate mockery --name=PaymentRepository --output=mocks --outpkg=mocks --filename=mock_payment_repository.go
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	UpdateStatusAndGet(ctx context.Context, yookassaID, newStatus string) (*models.Payment, error)
	GetByYooKassaID(ctx context.Context, yookassaID string) (*models.Payment, error)
}

//go:generate mockery --name=PaymentProvider --output=mocks --outpkg=mocks --filename=mock_payment_provider.go
type PaymentProvider interface {
	CreatePayment(ctx context.Context, amount int, currency, description string) (*models.PaymentProviderResponse, error)
}

//go:generate mockery --name=EventPublisher --output=mocks --outpkg=mocks --filename=mock_event_publisher.go
type EventPublisher interface {
	PublishOrderEvent(ctx context.Context, event models.OrderEvent) error
	Close() error
}
