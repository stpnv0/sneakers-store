package services

import (
	"cart_service/internal/models"
	"context"
)

// CartService определяет интерфейс для работы с корзиной
type CartService interface {
	// GetCart возвращает корзину пользователя
	GetCart(ctx context.Context, userSSOID int) (*models.Cart, error)

	// AddToCart добавляет товар в корзину
	AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error

	// UpdateCartItemQuantity обновляет количество товара в корзине
	UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error

	// RemoveFromCart удаляет товар из корзины
	RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error
}
