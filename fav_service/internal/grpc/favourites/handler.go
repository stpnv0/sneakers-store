package favourites

import (
	"context"
	"fav_service/internal/models"
	"fmt"
	"log/slog"
	"strconv"

	favv1 "github.com/stpnv0/protos/gen/go/favourites"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FavouritesService interface for business logic
type FavouritesService interface {
	AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error
	RemoveFromFavourite(ctx context.Context, userSSOID, sneakerID int) error
	GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error)
	IsFavourite(ctx context.Context, userSSOID, sneakerID int) (bool, error)
	GetByIDs(ctx context.Context, ids []int) ([]models.Favourite, error)
	ParseIDsString(idsParam string) ([]int, error)
}

// serverAPI implements the gRPC FavouritesServiceServer interface
type serverAPI struct {
	favv1.UnimplementedFavouritesServiceServer
	favService FavouritesService
	log        *slog.Logger
}

// Register registers the favourites service with the gRPC server
func Register(gRPC *grpc.Server, favService FavouritesService) {
	favv1.RegisterFavouritesServiceServer(gRPC, &serverAPI{
		favService: favService,
		log:        slog.Default(),
	})
}

// AddToFavourites implements FavouritesServiceServer.AddToFavourites
func (s *serverAPI) AddToFavourites(
	ctx context.Context,
	req *favv1.AddToFavouritesRequest,
) (*favv1.AddToFavouritesResponse, error) {
	const op = "favourites.AddToFavourites"

	// Validate request
	if req.GetSneakerId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sneaker_id must be positive")
	}

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Call business logic
	err = s.favService.AddToFavourite(ctx, userID, int(req.GetSneakerId()))
	if err != nil {
		s.log.Error("failed to add to favourites", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to add item to favourites")
	}

	return &favv1.AddToFavouritesResponse{
		Success: true,
		Message: "Item added to favourites successfully",
	}, nil
}

// RemoveFromFavourites implements FavouritesServiceServer.RemoveFromFavourites
func (s *serverAPI) RemoveFromFavourites(
	ctx context.Context,
	req *favv1.RemoveFromFavouritesRequest,
) (*favv1.RemoveFromFavouritesResponse, error) {
	const op = "favourites.RemoveFromFavourites"

	// Validate request
	if req.GetSneakerId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sneaker_id must be positive")
	}

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Remove item
	err = s.favService.RemoveFromFavourite(ctx, userID, int(req.GetSneakerId()))
	if err != nil {
		s.log.Error("failed to remove from favourites", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to remove item from favourites")
	}

	return &favv1.RemoveFromFavouritesResponse{
		Success: true,
		Message: "Item removed from favourites successfully",
	}, nil
}

// GetFavourites implements FavouritesServiceServer.GetFavourites
func (s *serverAPI) GetFavourites(
	ctx context.Context,
	req *favv1.GetFavouritesRequest,
) (*favv1.GetFavouritesResponse, error) {
	const op = "favourites.GetFavourites"

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Get favourites from service
	items, err := s.favService.GetAllFavourites(ctx, userID)
	if err != nil {
		s.log.Error("failed to get favourites", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get favourites")
	}

	// Convert to proto items
	protoItems := make([]*favv1.FavouriteItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, &favv1.FavouriteItem{
			Id:        int64(item.ID),
			UserId:    int64(item.UserSSOID),
			SneakerId: int64(item.SneakerID),
			AddedAt:   item.AddedAt.Unix(),
		})
	}

	return &favv1.GetFavouritesResponse{
		Items: protoItems,
	}, nil
}

// IsFavourite implements FavouritesServiceServer.IsFavourite
func (s *serverAPI) IsFavourite(
	ctx context.Context,
	req *favv1.IsFavouriteRequest,
) (*favv1.IsFavouriteResponse, error) {
	const op = "favourites.IsFavourite"

	// Validate request
	if req.GetSneakerId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sneaker_id must be positive")
	}

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Check if favourite
	isFav, err := s.favService.IsFavourite(ctx, userID, int(req.GetSneakerId()))
	if err != nil {
		s.log.Error("failed to check favourite status", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to check status")
	}

	return &favv1.IsFavouriteResponse{
		IsFavourite: isFav,
	}, nil
}

// GetFavouritesByIDs implements FavouritesServiceServer.GetFavouritesByIDs
func (s *serverAPI) GetFavouritesByIDs(
	ctx context.Context,
	req *favv1.GetFavouritesByIDsRequest,
) (*favv1.GetFavouritesByIDsResponse, error) {
	if len(req.GetIds()) == 0 {
		return &favv1.GetFavouritesByIDsResponse{Items: []*favv1.FavouriteItem{}}, nil
	}

	ids := make([]int, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		ids = append(ids, int(id))
	}

	favourites, err := s.favService.GetByIDs(ctx, ids)
	if err != nil {
		s.log.Error("failed to get favourites by ids", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get favourites")
	}

	items := make([]*favv1.FavouriteItem, 0, len(favourites))
	for _, f := range favourites {
		items = append(items, &favv1.FavouriteItem{
			Id:        int64(f.ID),
			UserId:    int64(f.UserSSOID),
			SneakerId: int64(f.SneakerID),
			AddedAt:   f.AddedAt.Unix(),
		})
	}

	return &favv1.GetFavouritesByIDsResponse{Items: items}, nil
}

// ContextKey — типизированный ключ для значений контекста, избегающий коллизий.
type ContextKey string

// UserIDKey — ключ для хранения user_id в контексте.
const UserIDKey ContextKey = "user_id"

// Helpers
func getUserIDFromContext(ctx context.Context) (int, error) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return 0, fmt.Errorf("user_id не найден в контексте")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("невалидный формат user_id: %w", err)
	}

	return userID, nil
}
