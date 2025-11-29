package order

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	orderv1 "github.com/stpnv0/protos/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api orderv1.OrderServiceClient
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "order.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded, codes.Unavailable),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: orderv1.NewOrderServiceClient(cc),
		log: log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

func (c *Client) CreateOrder(ctx context.Context, userID int64, items []*orderv1.OrderItem) (int64, error) {
	const op = "order.CreateOrder"

	req := &orderv1.CreateOrderRequest{
		UserId: int32(userID),
		Items:  items,
	}

	resp, err := c.api.CreateOrder(ctx, req)
	if err != nil {
		c.log.Error("failed to create order", slog.String("error", err.Error()))
		return 0, err
	}

	return int64(resp.Order.Id), nil
}

func (c *Client) GetOrder(ctx context.Context, orderID int64) (*orderv1.Order, error) {
	const op = "order.GetOrder"

	req := &orderv1.GetOrderRequest{
		OrderId: int32(orderID),
	}

	resp, err := c.api.GetOrder(ctx, req)
	if err != nil {
		c.log.Error("failed to get order", slog.String("error", err.Error()))
		return nil, err
	}

	return resp.Order, nil
}

func (c *Client) GetUserOrders(ctx context.Context, userID int64) ([]*orderv1.Order, error) {
	const op = "order.GetUserOrders"

	req := &orderv1.GetUserOrdersRequest{
		UserId: int32(userID),
	}

	resp, err := c.api.GetUserOrders(ctx, req)
	if err != nil {
		c.log.Error("failed to get user orders", slog.String("error", err.Error()))
		return nil, err
	}

	return resp.Orders, nil
}
