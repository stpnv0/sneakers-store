package order_test

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

	pb "github.com/stpnv0/protos/gen/go/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcserver "order_service/internal/grpc"
	handler "order_service/internal/grpc/order"
	handlerMocks "order_service/internal/grpc/order/mocks"
	"order_service/internal/models"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ctxWithUserID creates a context with user_id set via the grpc package helper.
func ctxWithUserID(userID string) context.Context {
	return grpcserver.ContextWithUserID(context.Background(), userID)
}

// ---------------------------------------------------------------------------
// CreateOrder
// ---------------------------------------------------------------------------

func TestCreateOrder_Success(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	now := time.Now()
	created := &models.OrderWithItems{
		Order: models.Order{
			ID:          1,
			UserID:      42,
			Status:      models.OrderStatusPendingPayment,
			TotalAmount: 300,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Items: []models.OrderItem{
			{SneakerID: 10, Quantity: 3, PriceAtPurchase: 100},
		},
	}

	svc.On("CreateOrder", mock.Anything, 42, mock.AnythingOfType("[]models.OrderItem")).
		Return(created, nil)

	resp, err := h.CreateOrder(ctxWithUserID("42"), &pb.CreateOrderRequest{
		UserId: 42,
		Items: []*pb.OrderItem{
			{SneakerId: 10, Quantity: 3, PriceAtPurchaseKopecks: 100},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.GetOrder().GetId())
	assert.Equal(t, int64(42), resp.GetOrder().GetUserId())
	svc.AssertExpectations(t)
}

func TestCreateOrder_NoAuth(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	_, err := h.CreateOrder(context.Background(), &pb.CreateOrderRequest{
		UserId: 1,
		Items: []*pb.OrderItem{
			{SneakerId: 1, Quantity: 1, PriceAtPurchaseKopecks: 100},
		},
	})

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestCreateOrder_InvalidItemData(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	_, err := h.CreateOrder(ctxWithUserID("1"), &pb.CreateOrderRequest{
		UserId: 1,
		Items: []*pb.OrderItem{
			{SneakerId: 0, Quantity: 1, PriceAtPurchaseKopecks: 100},
		},
	})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	_, err := h.CreateOrder(ctxWithUserID("1"), &pb.CreateOrderRequest{
		UserId: 1,
		Items:  []*pb.OrderItem{},
	})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateOrder_ServiceError(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	svc.On("CreateOrder", mock.Anything, 1, mock.Anything).
		Return(nil, errors.New("boom"))

	_, err := h.CreateOrder(ctxWithUserID("1"), &pb.CreateOrderRequest{
		UserId: 1,
		Items:  []*pb.OrderItem{{SneakerId: 1, Quantity: 1, PriceAtPurchaseKopecks: 100}},
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// ---------------------------------------------------------------------------
// GetOrder
// ---------------------------------------------------------------------------

func TestGetOrder_Success(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	now := time.Now()
	svc.On("GetOrder", mock.Anything, 5).Return(&models.OrderWithItems{
		Order: models.Order{ID: 5, UserID: 1, CreatedAt: now, UpdatedAt: now},
	}, nil)

	resp, err := h.GetOrder(ctxWithUserID("1"), &pb.GetOrderRequest{OrderId: 5})
	require.NoError(t, err)
	assert.Equal(t, int64(5), resp.GetOrder().GetId())
}

func TestGetOrder_InvalidID(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	_, err := h.GetOrder(ctxWithUserID("1"), &pb.GetOrderRequest{OrderId: -1})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetOrder_PermissionDenied(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	now := time.Now()
	svc.On("GetOrder", mock.Anything, 5).Return(&models.OrderWithItems{
		Order: models.Order{ID: 5, UserID: 99, CreatedAt: now, UpdatedAt: now},
	}, nil)

	_, err := h.GetOrder(ctxWithUserID("1"), &pb.GetOrderRequest{OrderId: 5})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

// ---------------------------------------------------------------------------
// GetUserOrders
// ---------------------------------------------------------------------------

func TestGetUserOrders_Success(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	now := time.Now()
	svc.On("GetUserOrders", mock.Anything, 42).Return([]*models.OrderWithItems{
		{Order: models.Order{ID: 1, UserID: 42, CreatedAt: now, UpdatedAt: now}},
		{Order: models.Order{ID: 2, UserID: 42, CreatedAt: now, UpdatedAt: now}},
	}, nil)

	resp, err := h.GetUserOrders(ctxWithUserID("42"), &pb.GetUserOrdersRequest{UserId: 42})
	require.NoError(t, err)
	assert.Len(t, resp.GetOrders(), 2)
}

// ---------------------------------------------------------------------------
// UpdateOrderStatus
// ---------------------------------------------------------------------------

func TestUpdateOrderStatus_Success(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	svc.On("UpdateOrderStatus", mock.Anything, 1, "PAID").Return(nil)

	resp, err := h.UpdateOrderStatus(context.Background(), &pb.UpdateOrderStatusRequest{
		OrderId: 1,
		Status:  "PAID",
	})

	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
}

func TestUpdateOrderStatus_InvalidID(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	_, err := h.UpdateOrderStatus(context.Background(), &pb.UpdateOrderStatusRequest{
		OrderId: 0,
		Status:  "PAID",
	})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestUpdateOrderStatus_EmptyStatus(t *testing.T) {
	svc := new(handlerMocks.MockService)
	h := handler.NewHandler(svc, newTestLogger())

	_, err := h.UpdateOrderStatus(context.Background(), &pb.UpdateOrderStatusRequest{
		OrderId: 1,
		Status:  "",
	})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
