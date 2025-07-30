package product

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/stpnv0/protos/gen/go/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	domain "product_service/internal/domain/model"
	"product_service/internal/repository"
)

type App interface {
	GetSneakerByID(ctx context.Context, id int64) (*domain.Sneaker, error)
	AddSneaker(ctx context.Context, sneaker *domain.Sneaker) (int64, error)
	GetAllSneakers(ctx context.Context, limit, offset uint64) ([]*domain.Sneaker, error)
	GetSneakersByIDs(ctx context.Context, ids []int64) ([]*domain.Sneaker, error) // Новый в интерфейсе
	DeleteSneaker(ctx context.Context, id int64) error
	GenerateUploadURL(ctx context.Context, originalFilename string, contentType string) (uploadURL string, fileKey string, err error)
	UpdateProductImage(ctx context.Context, productID int64, imageKey string) error
}

type serverAPI struct {
	pb.UnimplementedProductServer
	app App
	log *slog.Logger
}

func Register(gRPCServer *grpc.Server, app App, log *slog.Logger) {
	pb.RegisterProductServer(gRPCServer, &serverAPI{app: app, log: log})
}

func (s *serverAPI) GetSneakerByID(ctx context.Context, req *pb.GetSneakerByIDRequest) (*pb.Sneaker, error) {
	sneaker, err := s.app.GetSneakerByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "sneaker not found")
		}
		s.log.Error("failed to get sneaker by id", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return toProtoSneaker(sneaker), nil
}

func (s *serverAPI) AddSneaker(ctx context.Context, req *pb.AddSneakerRequest) (*pb.Sneaker, error) {
	// Валидация
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}

	id, err := s.app.AddSneaker(ctx, &domain.Sneaker{
		Title: req.GetTitle(),
		Price: req.GetPrice(),
	})
	if err != nil {
		s.log.Error("failed to add sneaker", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.Sneaker{Id: id, Title: req.GetTitle(), Price: req.GetPrice(), ImageKey: ""}, nil
}

func (s *serverAPI) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
	uploadURL, fileKey, err := s.app.GenerateUploadURL(ctx, req.GetOriginalFilename(), req.GetContentType())
	if err != nil {
		s.log.Error("failed to generate upload url", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.GenerateUploadURLResponse{
		UploadUrl: uploadURL,
		FileKey:   fileKey,
	}, nil
}

func (s *serverAPI) UpdateProductImage(ctx context.Context, req *pb.UpdateProductImageRequest) (*emptypb.Empty, error) {
	err := s.app.UpdateProductImage(ctx, req.GetProductId(), req.GetImageKey())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "product not found to update image")
		}
		s.log.Error("failed to update product image", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &emptypb.Empty{}, nil
}

func (s *serverAPI) GetAllSneakers(ctx context.Context, req *pb.GetAllSneakersRequest) (*pb.GetAllSneakersResponse, error) {
	limit := req.GetLimit()
	if limit == 0 || limit > 100 {
		limit = 20
	}
	sneakers, err := s.app.GetAllSneakers(ctx, limit, req.GetOffset())
	if err != nil {
		s.log.Error("failed to get all sneakers", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}

	protoSneakers := make([]*pb.Sneaker, len(sneakers))
	for i, sn := range sneakers {
		protoSneakers[i] = toProtoSneaker(sn)
	}

	return &pb.GetAllSneakersResponse{Sneakers: protoSneakers}, nil
}

func (s *serverAPI) GetSneakersByIDs(ctx context.Context, req *pb.GetSneakersByIDsRequest) (*pb.GetSneakersByIDsResponse, error) {
	sneakers, err := s.app.GetSneakersByIDs(ctx, req.GetIds())
	if err != nil {
		s.log.Error("failed to get sneakers by ids", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}

	protoSneakers := make([]*pb.Sneaker, len(sneakers))
	for i, sn := range sneakers {
		protoSneakers[i] = toProtoSneaker(sn)
	}

	return &pb.GetSneakersByIDsResponse{Sneakers: protoSneakers}, nil
}

func (s *serverAPI) DeleteSneaker(ctx context.Context, req *pb.DeleteSneakerRequest) (*emptypb.Empty, error) {
	err := s.app.DeleteSneaker(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "sneaker not found")
		}
		s.log.Error("failed to delete sneaker", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &emptypb.Empty{}, nil
}

func toProtoSneaker(sneaker *domain.Sneaker) *pb.Sneaker {
	return &pb.Sneaker{
		Id:       sneaker.Id,
		Title:    sneaker.Title,
		Price:    sneaker.Price,
		ImageKey: sneaker.ImageKey,
	}
}
