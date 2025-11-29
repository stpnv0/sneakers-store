package cart

import (
	"cart_service/internal/models"
	"context"
	"fmt"
	"log/slog"
	"strconv"

	cartv1 "github.com/stpnv0/protos/gen/go/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CartService interface for business logic
type CartService interface {
	AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error
	GetCart(ctx context.Context, userSSOID int) (*models.Cart, error)
	UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error
	RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error
	ClearCart(ctx context.Context, userSSOID int) error
}

// serverAPI implements the gRPC CartServiceServer interface
type serverAPI struct {
	cartv1.UnimplementedCartServiceServer
	cartService CartService
	log         *slog.Logger
}

// Register registers the cart service with the gRPC server
func Register(gRPC *grpc.Server, cartService CartService) {
	cartv1.RegisterCartServiceServer(gRPC, &serverAPI{
		cartService: cartService,
		log:         slog.Default(),
	})
}

// AddToCart implements CartServiceServer.AddToCart
func (s *serverAPI) AddToCart(
	ctx context.Context,
	req *cartv1.AddToCartRequest,
) (*cartv1.AddToCartResponse, error) {
	const op = "cart.AddToCart"

	// Validate request
	if req.GetSneakerId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sneaker_id must be positive")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}

	// Extract user ID from context (set by interceptor)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Call business logic
	err = s.cartService.AddToCart(ctx, userID, int(req.GetSneakerId()), int(req.GetQuantity()))
	if err != nil {
		s.log.Error("failed to add to cart", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to add item to cart")
	}

	return &cartv1.AddToCartResponse{
		Success: true,
		Message: "Item added to cart successfully",
	}, nil
}

// GetCart implements CartServiceServer.GetCart
func (s *serverAPI) GetCart(
	ctx context.Context,
	req *cartv1.GetCartRequest,
) (*cartv1.GetCartResponse, error) {
	const op = "cart.GetCart"

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Get cart from service
	cart, err := s.cartService.GetCart(ctx, userID)
	if err != nil {
		s.log.Error("failed to get cart", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get cart")
	}

	// Convert to proto Cart
	protoCart := convertToProtoCart(cart)

	return &cartv1.GetCartResponse{
		Cart: protoCart,
	}, nil
}

// UpdateCartItemQuantity implements CartServiceServer.UpdateCartItemQuantity
func (s *serverAPI) UpdateCartItemQuantity(
	ctx context.Context,
	req *cartv1.UpdateQuantityRequest,
) (*cartv1.UpdateQuantityResponse, error) {
	const op = "cart.UpdateCartItemQuantity"

	// Validate request
	if req.GetItemId() == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Update quantity
	err = s.cartService.UpdateCartItemQuantity(ctx, userID, req.GetItemId(), int(req.GetQuantity()))
	if err != nil {
		s.log.Error("failed to update quantity", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to update item quantity")
	}

	return &cartv1.UpdateQuantityResponse{
		Success: true,
		Message: "Item quantity updated successfully",
	}, nil
}

// RemoveFromCart implements CartServiceServer.RemoveFromCart
func (s *serverAPI) RemoveFromCart(
	ctx context.Context,
	req *cartv1.RemoveFromCartRequest,
) (*cartv1.RemoveFromCartResponse, error) {
	const op = "cart.RemoveFromCart"

	// Validate request
	if req.GetItemId() == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Remove item
	err = s.cartService.RemoveFromCart(ctx, userID, req.GetItemId())
	if err != nil {
		s.log.Error("failed to remove from cart", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to remove item from cart")
	}

	return &cartv1.RemoveFromCartResponse{
		Success: true,
		Message: "Item removed from cart successfully",
	}, nil
}

// ClearCart implements CartServiceServer.ClearCart
func (s *serverAPI) ClearCart(
	ctx context.Context,
	req *cartv1.ClearCartRequest,
) (*cartv1.ClearCartResponse, error) {
	const op = "cart.ClearCart"

	// Extract user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Clear cart
	err = s.cartService.ClearCart(ctx, userID)
	if err != nil {
		s.log.Error("failed to clear cart", slog.String("op", op), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to clear cart")
	}

	return &cartv1.ClearCartResponse{
		Success: true,
		Message: "Cart cleared successfully",
	}, nil
}

// Helper functions

func getUserIDFromContext(ctx context.Context) (int, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return 0, fmt.Errorf("user_id not found in context")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid user_id format: %w", err)
	}

	return userID, nil
}

func convertToProtoCart(cart *models.Cart) *cartv1.Cart {
	if cart == nil {
		return &cartv1.Cart{
			UserId:    0,
			Items:     []*cartv1.CartItem{},
			UpdatedAt: 0,
		}
	}

	protoItems := make([]*cartv1.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		protoItems = append(protoItems, &cartv1.CartItem{
			Id:        item.ID,
			SneakerId: int32(item.SneakerID),
			Quantity:  int32(item.Quantity),
			AddedAt:   item.AddedAt.Unix(),
		})
	}

	return &cartv1.Cart{
		UserId:    int32(cart.UserSSOID),
		Items:     protoItems,
		UpdatedAt: cart.UpdatedAt.Unix(),
	}
}
