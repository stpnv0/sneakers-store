package services

import (
	"context"
	"time"

	"cart_service/internal/models"
)

// CartService определяет интерфейс для работы с корзиной
type CartService interface {
	GetCart(ctx context.Context, userSSOID int) (*models.Cart, error)
	AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error
	UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error
	RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error
	ClearCart(ctx context.Context, userSSOID int) error
}

// CartRepository — интерфейс основного хранилища данных (PostgreSQL).
//
//go:generate mockery --name=CartRepository --output=mocks --outpkg=mocks --filename=mock_cart_repository.go
type CartRepository interface {
	GetCart(ctx context.Context, userSSOID int) (*models.Cart, error)
	AddCartItem(ctx context.Context, item *models.CartItem) error
	UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error
	RemoveCartItem(ctx context.Context, userSSOID int, itemID string) error
	ClearCart(ctx context.Context, userSSOID int) error
}

// CartCache — интерфейс кэширования (Redis).
//
//go:generate mockery --name=CartCache --output=mocks --outpkg=mocks --filename=mock_cart_cache.go
type CartCache interface {
	GetCart(ctx context.Context, userSSOID int) (*models.Cart, error)
	SetCart(ctx context.Context, userSSOID int, cart *models.Cart, ttl time.Duration) error
	InvalidateCart(ctx context.Context, userSSOID int) error
	AddToCartItem(ctx context.Context, item models.CartItem) error
	UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, newQuantity int) error
	RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error
}
