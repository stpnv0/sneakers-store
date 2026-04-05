package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"order_service/internal/models"
	"order_service/internal/service"
	"order_service/internal/service/mocks"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestService() (
	*service.OrderServiceImpl,
	*mocks.MockOrderRepository,
	*mocks.MockPaymentRepository,
	*mocks.MockPaymentProvider,
	*mocks.MockEventPublisher,
) {
	repo := new(mocks.MockOrderRepository)
	paymentRepo := new(mocks.MockPaymentRepository)
	provider := new(mocks.MockPaymentProvider)
	pub := new(mocks.MockEventPublisher)
	svc := service.NewOrderService(repo, paymentRepo, provider, pub, newTestLogger())
	return svc, repo, paymentRepo, provider, pub
}

// ---------------------------------------------------------------------------
// CreateOrder
// ---------------------------------------------------------------------------

func TestCreateOrder_Success(t *testing.T) {
	svc, repo, paymentRepo, provider, pub := newTestService()

	items := []models.OrderItem{
		{SneakerID: 1, Quantity: 2, PriceAtPurchase: 100},
		{SneakerID: 2, Quantity: 1, PriceAtPurchase: 200},
	}

	expectedOrder := &models.OrderWithItems{
		Order: models.Order{
			ID: 1, UserID: 42, Status: models.OrderStatusPendingPayment,
			TotalAmount: 400, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		Items: items,
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Order"), items).
		Return(expectedOrder, nil)
	provider.On("CreatePayment", mock.Anything, 400, "RUB", "Order #1").
		Return(&models.PaymentProviderResponse{
			ID: "yoo-123", Status: "pending", ConfirmationURL: "https://pay.example.com/123",
		}, nil)
	paymentRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Payment")).Return(nil)
	repo.On("UpdatePaymentURL", mock.Anything, 1, "https://pay.example.com/123").Return(nil)
	pub.On("PublishOrderEvent", mock.Anything, mock.AnythingOfType("models.OrderEvent")).Return(nil)

	result, err := svc.CreateOrder(context.Background(), 42, items)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "https://pay.example.com/123", result.PaymentURL)
	repo.AssertExpectations(t)
	provider.AssertExpectations(t)
	paymentRepo.AssertExpectations(t)
}

func TestCreateOrder_RepositoryError(t *testing.T) {
	svc, repo, _, _, _ := newTestService()

	items := []models.OrderItem{{SneakerID: 1, Quantity: 1, PriceAtPurchase: 100}}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Order"), items).
		Return(nil, errors.New("db connection lost"))

	result, err := svc.CreateOrder(context.Background(), 42, items)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db connection lost")
}

func TestCreateOrder_PaymentProviderError_StillSucceeds(t *testing.T) {
	svc, repo, _, provider, _ := newTestService()

	items := []models.OrderItem{{SneakerID: 1, Quantity: 1, PriceAtPurchase: 100}}

	created := &models.OrderWithItems{
		Order: models.Order{
			ID: 5, UserID: 42, Status: models.OrderStatusPendingPayment,
			TotalAmount: 100, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		Items: items,
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Order"), items).
		Return(created, nil)
	provider.On("CreatePayment", mock.Anything, 100, "RUB", "Order #5").
		Return(nil, errors.New("yookassa unavailable"))

	result, err := svc.CreateOrder(context.Background(), 42, items)

	require.NoError(t, err, "order should succeed even when payment provider fails")
	assert.Equal(t, 5, result.ID)
	assert.Empty(t, result.PaymentURL)
}

// ---------------------------------------------------------------------------
// GetOrder / GetUserOrders
// ---------------------------------------------------------------------------

func TestGetOrder_Success(t *testing.T) {
	svc, repo, _, _, _ := newTestService()

	expected := &models.OrderWithItems{
		Order: models.Order{ID: 10, UserID: 1, Status: models.OrderStatusPaid},
	}
	repo.On("GetByID", mock.Anything, 10).Return(expected, nil)

	result, err := svc.GetOrder(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 10, result.ID)
}

func TestGetOrder_NotFound(t *testing.T) {
	svc, repo, _, _, _ := newTestService()

	repo.On("GetByID", mock.Anything, 999).Return(nil, errors.New("not found"))

	result, err := svc.GetOrder(context.Background(), 999)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetUserOrders_Success(t *testing.T) {
	svc, repo, _, _, _ := newTestService()

	orders := []*models.OrderWithItems{
		{Order: models.Order{ID: 1, UserID: 42}},
		{Order: models.Order{ID: 2, UserID: 42}},
	}
	repo.On("GetUserOrders", mock.Anything, 42).Return(orders, nil)

	result, err := svc.GetUserOrders(context.Background(), 42)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// ---------------------------------------------------------------------------
// UpdateOrderStatus
// ---------------------------------------------------------------------------

func TestUpdateOrderStatus_Success(t *testing.T) {
	svc, repo, _, _, _ := newTestService()

	existing := &models.OrderWithItems{
		Order: models.Order{ID: 1, Status: models.OrderStatusPendingPayment},
	}
	repo.On("GetByID", mock.Anything, 1).Return(existing, nil)
	repo.On("UpdateStatus", mock.Anything, 1, models.OrderStatusPaid, models.OrderStatusPendingPayment).Return(nil)

	err := svc.UpdateOrderStatus(context.Background(), 1, models.OrderStatusPaid)
	require.NoError(t, err)
}

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	svc, _, _, _, _ := newTestService()

	err := svc.UpdateOrderStatus(context.Background(), 1, "BOGUS")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")
}

// ---------------------------------------------------------------------------
// ProcessWebhook
// ---------------------------------------------------------------------------

func TestProcessWebhook_Succeeded(t *testing.T) {
	svc, repo, paymentRepo, _, pub := newTestService()

	existing := &models.Payment{
		ID: 1, OrderID: 10, YooKassaPaymentID: "yoo-abc", Status: models.PaymentStatusPending,
	}
	updated := &models.Payment{
		ID: 1, OrderID: 10, YooKassaPaymentID: "yoo-abc", Status: models.PaymentStatusSucceeded,
	}

	paymentRepo.On("GetByYooKassaID", mock.Anything, "yoo-abc").Return(existing, nil)
	paymentRepo.On("UpdateStatusAndGet", mock.Anything, "yoo-abc", models.PaymentStatusSucceeded).Return(updated, nil)

	orderWithItems := &models.OrderWithItems{
		Order: models.Order{ID: 10, Status: models.OrderStatusPendingPayment},
	}
	repo.On("GetByID", mock.Anything, 10).Return(orderWithItems, nil)
	repo.On("UpdateStatus", mock.Anything, 10, models.OrderStatusPaid, models.OrderStatusPendingPayment).Return(nil)
	pub.On("PublishOrderEvent", mock.Anything, mock.AnythingOfType("models.OrderEvent")).Return(nil)

	err := svc.ProcessWebhook(context.Background(), "yoo-abc", "succeeded")
	require.NoError(t, err)
	repo.AssertExpectations(t)
	paymentRepo.AssertExpectations(t)
}

func TestProcessWebhook_Canceled(t *testing.T) {
	svc, repo, paymentRepo, _, pub := newTestService()

	existing := &models.Payment{
		ID: 2, OrderID: 20, Status: models.PaymentStatusPending,
	}
	updated := &models.Payment{
		ID: 2, OrderID: 20, Status: models.PaymentStatusCanceled,
	}

	paymentRepo.On("GetByYooKassaID", mock.Anything, "yoo-xyz").Return(existing, nil)
	paymentRepo.On("UpdateStatusAndGet", mock.Anything, "yoo-xyz", models.PaymentStatusCanceled).Return(updated, nil)

	orderWithItems := &models.OrderWithItems{
		Order: models.Order{ID: 20, Status: models.OrderStatusPendingPayment},
	}
	repo.On("GetByID", mock.Anything, 20).Return(orderWithItems, nil)
	repo.On("UpdateStatus", mock.Anything, 20, models.OrderStatusPaymentFailed, models.OrderStatusPendingPayment).Return(nil)
	pub.On("PublishOrderEvent", mock.Anything, mock.AnythingOfType("models.OrderEvent")).Return(nil)

	err := svc.ProcessWebhook(context.Background(), "yoo-xyz", "canceled")
	require.NoError(t, err)
}

func TestProcessWebhook_AlreadyProcessed(t *testing.T) {
	svc, _, paymentRepo, _, _ := newTestService()

	existing := &models.Payment{
		ID: 3, OrderID: 30, Status: models.PaymentStatusPending,
	}

	paymentRepo.On("GetByYooKassaID", mock.Anything, "yoo-dup").Return(existing, nil)
	paymentRepo.On("UpdateStatusAndGet", mock.Anything, "yoo-dup", models.PaymentStatusSucceeded).Return(nil, nil)

	err := svc.ProcessWebhook(context.Background(), "yoo-dup", "succeeded")
	require.NoError(t, err)
}

func TestProcessWebhook_InvalidTransition(t *testing.T) {
	svc, _, paymentRepo, _, _ := newTestService()

	existing := &models.Payment{
		ID: 4, OrderID: 40, Status: models.PaymentStatusSucceeded,
	}

	paymentRepo.On("GetByYooKassaID", mock.Anything, "yoo-done").Return(existing, nil)

	err := svc.ProcessWebhook(context.Background(), "yoo-done", "canceled")
	require.NoError(t, err)
	paymentRepo.AssertNotCalled(t, "UpdateStatusAndGet")
}
