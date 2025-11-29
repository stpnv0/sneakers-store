package order

import (
	"context"
	"log/slog"
	"order_service/internal/models"
	"order_service/internal/service"

	pb "github.com/stpnv0/protos/gen/go/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedOrderServiceServer
	service *service.OrderService
	log     *slog.Logger
}

func NewHandler(service *service.OrderService, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items cannot be empty")
	}

	// proto items to domain items
	items := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		if item.SneakerId <= 0 || item.Quantity <= 0 || item.PriceAtPurchase <= 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid item data")
		}
		items[i] = models.OrderItem{
			SneakerID:       int(item.SneakerId),
			Quantity:        int(item.Quantity),
			PriceAtPurchase: int(item.PriceAtPurchase),
		}
	}

	order, err := h.service.CreateOrder(ctx, int(req.UserId), items)
	if err != nil {
		h.log.Error("failed to create order", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	return &pb.CreateOrderResponse{
		Order: convertToProto(order),
	}, nil
}

func (h *Handler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req.OrderId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	order, err := h.service.GetOrder(ctx, int(req.OrderId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &pb.GetOrderResponse{
		Order: convertToProto(order),
	}, nil
}

func (h *Handler) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	orders, err := h.service.GetUserOrders(ctx, int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get orders")
	}

	protoOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		protoOrders[i] = convertToProto(order)
	}

	return &pb.GetUserOrdersResponse{
		Orders: protoOrders,
	}, nil
}

func (h *Handler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	if req.OrderId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	err := h.service.UpdateOrderStatus(ctx, int(req.OrderId), req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update order status")
	}

	return &pb.UpdateOrderStatusResponse{
		Success: true,
	}, nil
}

func convertToProto(order *models.OrderWithItems) *pb.Order {
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &pb.OrderItem{
			SneakerId:       int32(item.SneakerID),
			Quantity:        int32(item.Quantity),
			PriceAtPurchase: int32(item.PriceAtPurchase),
		}
	}

	return &pb.Order{
		Id:          int32(order.ID),
		UserId:      int32(order.UserID),
		Status:      order.Status,
		TotalAmount: int32(order.TotalAmount),
		Items:       items,
		CreatedAt:   order.CreatedAt.Unix(),
		UpdatedAt:   order.UpdatedAt.Unix(),
		PaymentUrl:  order.PaymentURL,
	}
}
