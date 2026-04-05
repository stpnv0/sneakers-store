package order

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"api_gateway/internal/middleware"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	orderv1 "github.com/stpnv0/protos/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api  orderv1.OrderServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "order.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded, codes.Unavailable),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	// Общий таймаут вызова: (кол-во попыток * таймаут на попытку) + запас.
	callTimeout := timeout*time.Duration(retriesCount) + 2*time.Second

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			requestIDInterceptor(),
			deadlineInterceptor(callTimeout),
			grpclog.UnaryClientInterceptor(InterceptorLogger(log)),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:  orderv1.NewOrderServiceClient(cc),
		conn: cc,
		log:  log,
	}, nil
}

// Close закрывает базовое gRPC-соединение.
func (c *Client) Close() error {
	return c.conn.Close()
}

// deadlineInterceptor добавляет общий таймаут к каждому gRPC-вызову.
func deadlineInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

// requestIDInterceptor пробрасывает request_id из контекста в gRPC-метаданные.
func requestIDInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if rid := middleware.RequestIDFromContext(ctx); rid != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if ok {
				md = md.Copy()
			} else {
				md = metadata.New(nil)
			}
			md.Set("request_id", rid)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// attachUserMD добавляет user_id в gRPC-метаданные.
func attachUserMD(ctx context.Context, userID int64) context.Context {
	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) CreateOrder(ctx context.Context, userID int64, items []*orderv1.OrderItem) (*orderv1.Order, error) {
	const op = "order.CreateOrder"

	ctx = attachUserMD(ctx, userID)

	req := &orderv1.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	}

	resp, err := c.api.CreateOrder(ctx, req)
	if err != nil {
		c.log.Error("failed to create order", slog.String("error", err.Error()))
		return nil, err
	}

	return resp.GetOrder(), nil
}

func (c *Client) GetOrder(ctx context.Context, orderID int64) (*orderv1.Order, error) {
	const op = "order.GetOrder"

	req := &orderv1.GetOrderRequest{
		OrderId: orderID,
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

	ctx = attachUserMD(ctx, userID)

	req := &orderv1.GetUserOrdersRequest{
		UserId: userID,
	}

	resp, err := c.api.GetUserOrders(ctx, req)
	if err != nil {
		c.log.Error("failed to get user orders", slog.String("error", err.Error()))
		return nil, err
	}

	return resp.Orders, nil
}
