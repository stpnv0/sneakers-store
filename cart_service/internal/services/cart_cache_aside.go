package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cart_service/internal/models"
)

// CartCacheAsideService реализует паттерн Cache-Aside для работы с корзиной
type CartCacheAsideService struct {
	repo     CartRepository
	cache    CartCache
	logger   *slog.Logger
	cacheTTL time.Duration
}

func NewCartCacheAsideService(
	repo CartRepository,
	cache CartCache,
	logger *slog.Logger,
	cacheTTL time.Duration,
) *CartCacheAsideService {
	return &CartCacheAsideService{
		repo:     repo,
		cache:    cache,
		logger:   logger,
		cacheTTL: cacheTTL,
	}
}

// GetCart возвращает корзину пользователя, следуя паттерну Cache-Aside
func (s *CartCacheAsideService) GetCart(ctx context.Context, userSSOID int) (*models.Cart, error) {
	const op = "service.GetCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))

	cart, err := s.cache.GetCart(ctx, userSSOID)
	if err == nil {
		log.Debug("cache hit")
		return cart, nil
	}

	log.Debug("cache miss, loading from db")
	cart, err = s.repo.GetCart(ctx, userSSOID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.SetCart(ctx, userSSOID, cart, s.cacheTTL); err != nil {
		log.Warn("failed to cache cart", slog.String("error", err.Error()))
	}

	return cart, nil
}

// AddItemToCart добавляет товар в корзину с обновлением БД и кэша
func (s *CartCacheAsideService) AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error {
	const op = "service.AddToCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))

	item := &models.CartItem{
		UserSSOID: userSSOID,
		SneakerID: sneakerID,
		Quantity:  quantity,
		AddedAt:   time.Now(),
	}

	if err := s.repo.AddCartItem(ctx, item); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.AddToCartItem(ctx, *item); err != nil {
		log.Warn("failed to update cache, invalidating", slog.String("error", err.Error()))
		if invErr := s.cache.InvalidateCart(ctx, userSSOID); invErr != nil {
			log.Warn("failed to invalidate cache", slog.String("error", invErr.Error()))
		}
	}

	log.Info("item added to cart")
	return nil
}

// UpdateCartItemQuantity обновляет количество товара в корзине
func (s *CartCacheAsideService) UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error {
	const op = "service.UpdateCartItemQuantity"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))

	if err := s.repo.UpdateCartItemQuantity(ctx, userSSOID, itemID, quantity); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.UpdateCartItemQuantity(ctx, userSSOID, itemID, quantity); err != nil {
		log.Warn("failed to update cache, invalidating", slog.String("error", err.Error()))
		if invErr := s.cache.InvalidateCart(ctx, userSSOID); invErr != nil {
			log.Warn("failed to invalidate cache", slog.String("error", invErr.Error()))
		}
	}

	log.Info("item quantity updated")
	return nil
}

// RemoveCartItem удаляет товар из корзины
func (s *CartCacheAsideService) RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error {
	const op = "service.RemoveFromCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))

	if err := s.repo.RemoveCartItem(ctx, userSSOID, itemID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.RemoveFromCart(ctx, userSSOID, itemID); err != nil {
		log.Warn("failed to update cache, invalidating", slog.String("error", err.Error()))
		if invErr := s.cache.InvalidateCart(ctx, userSSOID); invErr != nil {
			log.Warn("failed to invalidate cache", slog.String("error", invErr.Error()))
		}
	}

	log.Info("item removed from cart")
	return nil
}

// ClearCart очищает корзину пользователя
func (s *CartCacheAsideService) ClearCart(ctx context.Context, userSSOID int) error {
	const op = "service.ClearCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))

	if err := s.repo.ClearCart(ctx, userSSOID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.InvalidateCart(ctx, userSSOID); err != nil {
		log.Warn("failed to invalidate cache", slog.String("error", err.Error()))
	}

	log.Info("cart cleared")
	return nil
}
