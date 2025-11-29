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
	// Recovery interceptor options
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("recovered from panic", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal server error")
		}),
	}

	// Create gRPC server with interceptors
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
				loggingInterceptor(log),
				userContextInterceptor(log),
			),
		),
	)

	// Register favourites service
	favourites.Register(gRPCServer, favService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
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

// loggingInterceptor logs all gRPC requests
func loggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Info("gRPC request",
			slog.String("method", info.FullMethod),
		)

		resp, err := handler(ctx, req)

		if err != nil {
			log.Error("gRPC request failed",
				slog.String("method", info.FullMethod),
				slog.String("error", err.Error()),
			)
		}

		return resp, err
	}
}

// userContextInterceptor extracts user_id from metadata and adds to context
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

		// Add user_id to context for use in handlers
		ctx = context.WithValue(ctx, "user_id", userIDs[0])

		return handler(ctx, req)
	}
}
