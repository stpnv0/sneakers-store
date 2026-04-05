package grpc

import (
	"context"
	"fav_service/internal/grpc/favourites"
	"fmt"
	"log/slog"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	favService favourites.FavouritesService,
	port int,
) *App {
	// Настройки interceptor'а восстановления после паник
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("recovered from panic", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal server error")
		}),
	}

	// Создаём gRPC-сервер с interceptor'ами
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
				loggingInterceptor(log),
				userContextInterceptor(log),
			),
		),
	)

	// Регистрируем сервис избранного
	favourites.Register(gRPCServer, favService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}


func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

// loggingInterceptor логирует все gRPC-запросы
func loggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		rid := requestIDFromMD(ctx)

		log.Info("gRPC request",
			slog.String("method", info.FullMethod),
			slog.String("request_id", rid),
		)

		resp, err := handler(ctx, req)

		if err != nil {
			log.Error("gRPC request failed",
				slog.String("method", info.FullMethod),
				slog.String("request_id", rid),
				slog.String("error", err.Error()),
			)
		}

		return resp, err
	}
}

// requestIDFromMD извлекает request_id из gRPC-метаданных.
func requestIDFromMD(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("request_id")
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// userContextInterceptor извлекает user_id из метаданных и добавляет в контекст
func userContextInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		userIDs := md.Get("user_id")
		if len(userIDs) == 0 {
			return nil, status.Error(codes.Unauthenticated, "user_id not found in metadata")
		}

		// Добавляем user_id в контекст для использования в хендлерах
		ctx = context.WithValue(ctx, favourites.UserIDKey, userIDs[0])

		return handler(ctx, req)
	}
}
