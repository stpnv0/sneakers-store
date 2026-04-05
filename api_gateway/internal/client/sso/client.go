package sso

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"api_gateway/internal/middleware"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	ssov1 "github.com/stpnv0/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api  ssov1.AuthClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "sso.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
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
		api:  ssov1.NewAuthClient(cc),
		conn: cc,
		log:  log,
	}, nil
}

// Close закрывает базовое gRPC-соединение.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Register(ctx context.Context, email string, password string) (int64, error) {
	const op = "grpc.Register"
	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetUserId(), nil
}

func (c *Client) Login(ctx context.Context, email string, password string, app_id int32) (string, error) {
	const op = "grpc.Login"

	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    app_id,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.Token, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "grpc.IsAdmin"

	resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.IsAdmin, nil
}

// GetAppSecret удалён — секрет подписи JWT передаётся через APP_SECRET.

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
