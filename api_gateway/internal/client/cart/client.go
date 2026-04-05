package cart

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"api_gateway/internal/middleware"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	cartv1 "github.com/stpnv0/protos/gen/go/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api  cartv1.CartServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "cart.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	// Общий таймаут вызова
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
		api:  cartv1.NewCartServiceClient(cc),
		conn: cc,
		log:  log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) AddToCart(ctx context.Context, userID int64, sneakerID int64, quantity int32) error {
	const op = "grpc.AddToCart"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.api.AddToCart(ctx, &cartv1.AddToCartRequest{
		SneakerId: sneakerID,
		Quantity:  quantity,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GetCart(ctx context.Context, userID int64) (*cartv1.Cart, error) {
	const op = "grpc.GetCart"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := c.api.GetCart(ctx, &cartv1.GetCartRequest{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetCart(), nil
}

func (c *Client) UpdateCartItemQuantity(ctx context.Context, userID int64, itemID string, quantity int32) error {
	const op = "grpc.UpdateCartItemQuantity"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.api.UpdateCartItemQuantity(ctx, &cartv1.UpdateQuantityRequest{
		ItemId:   itemID,
		Quantity: quantity,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) RemoveFromCart(ctx context.Context, userID int64, itemID string) error {
	const op = "grpc.RemoveFromCart"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.api.RemoveFromCart(ctx, &cartv1.RemoveFromCartRequest{
		ItemId: itemID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) ClearCart(ctx context.Context, userID int64) error {
	const op = "grpc.ClearCart"

	md := metadata.New(map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.api.ClearCart(ctx, &cartv1.ClearCartRequest{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// deadlineInterceptor добавляет общий таймаут к каждому gRPC-вызову
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

// requestIDInterceptor пробрасывает request_id из контекста в gRPC-метаданные
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
