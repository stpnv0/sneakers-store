package order

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/go-playground/validator/v10"
	pb "github.com/stpnv0/protos/gen/go/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcserver "order_service/internal/grpc"
	"order_service/internal/models"
)

//go:generate mockery --name=Service --output=mocks --outpkg=mocks --filename=mock_service.go
type Service interface {
	CreateOrder(ctx context.Context, userID int, items []models.OrderItem) (*models.OrderWithItems, error)
	GetOrder(ctx context.Context, orderID int) (*models.OrderWithItems, error)
	GetUserOrders(ctx context.Context, userID int) ([]*models.OrderWithItems, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
	ProcessWebhook(ctx context.Context, yookassaID, status string) error
}

type createOrderInput struct {
	UserID int              `validate:"required,gt=0"`
	Items  []orderItemInput `validate:"required,min=1,dive"`
}

type orderItemInput struct {
	SneakerID       int `validate:"required,gt=0"`
	Quantity        int `validate:"required,gt=0,lte=100"`
	PriceAtPurchase int `validate:"required,gt=0"`
}

type Handler struct {
	pb.UnimplementedOrderServiceServer
	svc      Service
	log      *slog.Logger
	validate *validator.Validate
}

func NewHandler(svc Service, log *slog.Logger) *Handler {
	return &Handler{
		svc:      svc,
		log:      log,
		validate: validator.New(),
	}
}

func (h *Handler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	input := createOrderInput{
		UserID: userID,
		Items:  make([]orderItemInput, len(req.GetItems())),
	}
	for i, item := range req.GetItems() {
		input.Items[i] = orderItemInput{
			SneakerID:       int(item.GetSneakerId()),
			Quantity:        int(item.GetQuantity()),
			PriceAtPurchase: int(item.GetPriceAtPurchaseKopecks()),
		}
	}

	if err := h.validate.Struct(input); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %s", err.Error())
	}

	items := make([]models.OrderItem, len(input.Items))
	for i, it := range input.Items {
		items[i] = models.OrderItem{
			SneakerID:       it.SneakerID,
			Quantity:        it.Quantity,
			PriceAtPurchase: it.PriceAtPurchase,
		}
	}

	order, err := h.svc.CreateOrder(ctx, input.UserID, items)
	if err != nil {
		h.log.Error("create order failed", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	return &pb.CreateOrderResponse{Order: orderToProto(order)}, nil
}

func (h *Handler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	order, err := h.svc.GetOrder(ctx, int(req.GetOrderId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	if order.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	return &pb.GetOrderResponse{Order: orderToProto(order)}, nil
}

func (h *Handler) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	orders, err := h.svc.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get orders")
	}

	out := make([]*pb.Order, len(orders))
	for i, o := range orders {
		out[i] = orderToProto(o)
	}

	return &pb.GetUserOrdersResponse{Orders: out}, nil
}

func (h *Handler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	if req.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}
	if req.GetStatus() == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	if err := h.svc.UpdateOrderStatus(ctx, int(req.GetOrderId()), req.GetStatus()); err != nil {
		return nil, status.Error(codes.Internal, "failed to update order status")
	}

	return &pb.UpdateOrderStatusResponse{Success: true}, nil
}

// getUserIDFromContext извлекает user_id, установленный interceptor'ом из gRPC-метаданных.
func getUserIDFromContext(ctx context.Context) (int, error) {
	userIDStr := grpcserver.UserIDFromContext(ctx)
	if userIDStr == "" {
		return 0, fmt.Errorf("user_id не найден в контексте")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("невалидный формат user_id: %w", err)
	}

	return userID, nil
}

func orderToProto(o *models.OrderWithItems) *pb.Order {
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			SneakerId:              int64(item.SneakerID),
			Quantity:               int32(item.Quantity),
			PriceAtPurchaseKopecks: int64(item.PriceAtPurchase),
		}
	}

	return &pb.Order{
		Id:                 int64(o.ID),
		UserId:             int64(o.UserID),
		Status:             o.Status,
		TotalAmountKopecks: int64(o.TotalAmount),
		Items:              items,
		CreatedAt:          o.CreatedAt.Unix(),
		UpdatedAt:          o.UpdatedAt.Unix(),
		PaymentUrl:         o.PaymentURL,
	}
}
