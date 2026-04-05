package favourites

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"api_gateway/internal/middleware"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	favv1 "github.com/stpnv0/protos/gen/go/favourites"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api  favv1.FavouritesServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "favourites.grpc.New"

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
		api:  favv1.NewFavouritesServiceClient(cc),
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

func (c *Client) AddToFavourites(ctx context.Context, userID, sneakerID int64) error {
	const op = "favourites.grpc.AddToFavourites"

	ctx = attachUserMD(ctx, userID)

	_, err := c.api.AddToFavourites(ctx, &favv1.AddToFavouritesRequest{
		SneakerId: sneakerID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) RemoveFromFavourites(ctx context.Context, userID, sneakerID int64) error {
	const op = "favourites.grpc.RemoveFromFavourites"

	ctx = attachUserMD(ctx, userID)

	_, err := c.api.RemoveFromFavourites(ctx, &favv1.RemoveFromFavouritesRequest{
		SneakerId: sneakerID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GetFavourites(ctx context.Context, userID int64) ([]*favv1.FavouriteItem, error) {
	const op = "favourites.grpc.GetFavourites"

	ctx = attachUserMD(ctx, userID)

	resp, err := c.api.GetFavourites(ctx, &favv1.GetFavouritesRequest{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp.Items, nil
}

func (c *Client) IsFavourite(ctx context.Context, userID, sneakerID int64) (bool, error) {
	const op = "favourites.grpc.IsFavourite"

	ctx = attachUserMD(ctx, userID)

	resp, err := c.api.IsFavourite(ctx, &favv1.IsFavouriteRequest{
		SneakerId: sneakerID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.IsFavourite, nil
}
