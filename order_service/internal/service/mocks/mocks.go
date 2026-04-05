package mocks

import (
	"context"

	"order_service/internal/models"

	"github.com/stretchr/testify/mock"
)

// --- MockOrderRepository ---

type MockOrderRepository struct{ mock.Mock }

func (m *MockOrderRepository) Create(ctx context.Context, order *models.Order, items []models.OrderItem) (*models.OrderWithItems, error) {
	args := m.Called(ctx, order, items)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrderWithItems), args.Error(1)
}
func (m *MockOrderRepository) GetByID(ctx context.Context, orderID int) (*models.OrderWithItems, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrderWithItems), args.Error(1)
}
func (m *MockOrderRepository) GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.OrderWithItems), args.Error(1)
}
func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID int, newStatus, expectedCurrentStatus string) error {
	return m.Called(ctx, orderID, newStatus, expectedCurrentStatus).Error(0)
}
func (m *MockOrderRepository) UpdatePaymentURL(ctx context.Context, orderID int, paymentURL string) error {
	return m.Called(ctx, orderID, paymentURL).Error(0)
}

// --- MockPaymentRepository ---

type MockPaymentRepository struct{ mock.Mock }

func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return m.Called(ctx, payment).Error(0)
}
func (m *MockPaymentRepository) UpdateStatusAndGet(ctx context.Context, yookassaID, newStatus string) (*models.Payment, error) {
	args := m.Called(ctx, yookassaID, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}
func (m *MockPaymentRepository) GetByYooKassaID(ctx context.Context, yookassaID string) (*models.Payment, error) {
	args := m.Called(ctx, yookassaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

// --- MockPaymentProvider ---

type MockPaymentProvider struct{ mock.Mock }

func (m *MockPaymentProvider) CreatePayment(ctx context.Context, amount int, currency, description string) (*models.PaymentProviderResponse, error) {
	args := m.Called(ctx, amount, currency, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentProviderResponse), args.Error(1)
}

// --- MockEventPublisher ---

type MockEventPublisher struct{ mock.Mock }

func (m *MockEventPublisher) PublishOrderEvent(ctx context.Context, event models.OrderEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *MockEventPublisher) Close() error {
	return m.Called().Error(0)
}
