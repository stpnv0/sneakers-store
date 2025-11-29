package grpc

import (
	"context"
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

type Server struct {
	grpcServer *grpc.Server
	log        *slog.Logger
}

func NewServer(log *slog.Logger) *Server {
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Error("panic recovered", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal server error")
		}),
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
			loggingInterceptor(log),
			userContextInterceptor,
		)),
	)

	return &Server{
		grpcServer: grpcServer,
		log:        log,
	}
}

func (s *Server) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}

func (s *Server) Run(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.log.Info("grpc server started", slog.Int("port", port))
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
	s.log.Info("grpc server stopped")
}

func loggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Info("gRPC request", slog.String("method", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			log.Error("gRPC error", slog.String("method", info.FullMethod), slog.String("error", err.Error()))
		}
		return resp, err
	}
}

func userContextInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if userIDs := md.Get("user_id"); len(userIDs) > 0 {
			ctx = context.WithValue(ctx, "user_id", userIDs[0])
		}
	}
	return handler(ctx, req)
}
