package cart

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	cartv1 "github.com/stpnv0/protos/gen/go/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api cartv1.CartServiceClient
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "cart.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
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
		api: cartv1.NewCartServiceClient(cc),
		log: log,
	}, nil
}

func (c *Client) AddToCart(ctx context.Context, userID int, sneakerID, quantity int32) error {
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

func (c *Client) GetCart(ctx context.Context, userID int) (*cartv1.Cart, error) {
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

func (c *Client) UpdateCartItemQuantity(ctx context.Context, userID int, itemID string, quantity int32) error {
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

func (c *Client) RemoveFromCart(ctx context.Context, userID int, itemID string) error {
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

func (c *Client) ClearCart(ctx context.Context, userID int) error {
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

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}
