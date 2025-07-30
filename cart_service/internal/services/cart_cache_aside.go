package services

import (
	"cart_service/internal/models"
	"cart_service/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"time"
)

// CartCacheAsideService реализует паттерн Cache-Aside для работы с корзиной
type CartCacheAsideService struct {
	pgRepo    *repository.PostgresRepository
	redisRepo *repository.RedisRepository
	logger    *slog.Logger
	cacheTTL  time.Duration
}

func NewCartCacheAsideService(
	pgRepo *repository.PostgresRepository,
	redisRepo *repository.RedisRepository,
	logger *slog.Logger,
	cacheTTL time.Duration,
) *CartCacheAsideService {
	return &CartCacheAsideService{
		pgRepo:    pgRepo,
		redisRepo: redisRepo,
		logger:    logger,
		cacheTTL:  cacheTTL,
	}
}

// GetCart возвращает корзину пользователя, следуя паттерну Cache-Aside
func (s *CartCacheAsideService) GetCart(ctx context.Context, userSSOID int) (*models.Cart, error) {
	const op = "service.GetCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))

	// 1. Пробуем получить из кэша
	cart, err := s.redisRepo.GetCart(ctx, userSSOID)
	if err == nil {
		// Кэш-хит: возвращаем данные из кэша
		if len(cart.Items) > 0 {
			log.Info("cache hit")
			return cart, nil
		}
		log.Info("cache hit but cart is empty, checking DB")
	}

	// 2. Кэш-промах: загружаем из PostgreSQL
	log.Info("cache miss, loading from db")
	cart, err = s.pgRepo.GetCart(ctx, userSSOID)
	if err != nil {
		log.Error("failed to get cart from postgres", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 3. Заполняем кэш новыми данными
	if err = s.redisRepo.SetCart(ctx, userSSOID, cart, s.cacheTTL); err != nil {
		log.Warn("failed to cache cart", slog.String("error", err.Error()))
		// Продолжаем работу даже при ошибке кэширования
	}

	return cart, nil
}

// AddItemToCart добавляет товар в корзину с обновлением БД и кэша
func (s *CartCacheAsideService) AddItemToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error {
	const op = "service.AddToCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID), slog.Int("sneaker_id", sneakerID))

	// 1. Добавляем товар в основную БД
	item := &models.CartItem{
		UserSSOID: userSSOID,
		SneakerID: sneakerID,
		Quantity:  quantity,
		AddedAt:   time.Now(),
	}

	err := s.pgRepo.AddCartItem(ctx, item)
	if err != nil {
		log.Error("failed to add item to postgres", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	// 2. Точечно обновляем кэш - добавляем элемент вместо инвалидации всей корзины
	// Это оптимизация: добавляем товар напрямую в Redis вместо полной инвалидации
	if err = s.redisRepo.AddToCartItem(ctx, *item); err != nil {
		log.Warn("failed to invalidate cart cache", slog.String("error", err.Error()))
		// При ошибке добавления в кэш - инвалидируем кэш для консистентности
		if invalidateErr := s.redisRepo.InvalidateCart(ctx, userSSOID); invalidateErr != nil {
			log.Warn("Also failed to invalidate cache for user", slog.String("error", err.Error()))
		}
	}

	log.Info("item successfully added to cart")
	return nil
}

// UpdateCartItemQuantity обновляет количество товара в корзине
func (s *CartCacheAsideService) UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error {
	const op = "service.UpdateCartItemQuantity"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID), slog.String("item_id", itemID))

	// 1. Обновляем в PostgreSQL
	err := s.pgRepo.UpdateCartItemQuantity(ctx, userSSOID, itemID, quantity)
	if err != nil {
		log.Error("failed to update item in postgres", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	// 2. Точечно обновляем кэш
	if err = s.redisRepo.UpdateCartItemQuantity(ctx, userSSOID, itemID, quantity); err != nil {
		log.Warn("failed to invalidate cart cache", slog.String("error", err.Error()))
		// При ошибке обновления в кэше - инвалидируем кэш для консистентности
		if invalidateErr := s.redisRepo.InvalidateCart(ctx, userSSOID); invalidateErr != nil {
			log.Warn("Also failed to invalidate cache for user", slog.String("error", err.Error()))
		}
	}

	log.Info("item quantity successfully updated")
	return nil
}

// RemoveCartItem удаляет товар из корзины
func (s *CartCacheAsideService) RemoveCartItem(ctx context.Context, userSSOID int, itemID string) error {
	const op = "service.RemoveFromCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID), slog.String("item_id", itemID))
	// 1. Удаляем из PostgreSQL
	err := s.pgRepo.RemoveCartItem(ctx, userSSOID, itemID)
	if err != nil {
		log.Error("failed to remove item from postgres", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	// 2. Точечно удаляем из кэша (вместо полной инвалидации)
	if err = s.redisRepo.RemoveFromCart(ctx, userSSOID, itemID); err != nil {
		log.Warn("failed to invalidate cart cache", slog.String("error", err.Error()))
		// При ошибке удаления из кэша - инвалидируем кэш для консистентности
		if invalidateErr := s.redisRepo.InvalidateCart(ctx, userSSOID); invalidateErr != nil {
			log.Warn("Also failed to invalidate cache for user", slog.String("error", err.Error()))
		}
	}

	log.Info("item successfully removed from cart")
	return nil
}

// ClearCart очищает корзину пользователя
func (s *CartCacheAsideService) ClearCart(ctx context.Context, userSSOID int) error {
	const op = "service.ClearCart"
	log := s.logger.With(slog.String("op", op), slog.Int("user_id", userSSOID))
	// 1. Очищаем в PostgreSQL
	err := s.pgRepo.ClearCart(ctx, userSSOID)
	if err != nil {
		log.Error("failed to clear cart in postgres", slog.String("error", err.Error()))
		return fmt.Errorf("failed to clear cart in PostgreSQL: %w", err)
	}

	// 2. Полностью инвалидируем кэш - здесь это логично, так как очищаем всю корзину
	if err = s.redisRepo.InvalidateCart(ctx, userSSOID); err != nil {
		log.Warn("failed to invalidate cart cache", slog.String("error", err.Error()))
	}

	log.Info("cart successfully cleared")
	return nil
}

// AddToCart - метод для совместимости с интерфейсом CartService
func (s *CartCacheAsideService) AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error {
	return s.AddItemToCart(ctx, userSSOID, sneakerID, quantity)
}

// RemoveFromCart - метод для совместимости с интерфейсом CartService
func (s *CartCacheAsideService) RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error {
	return s.RemoveCartItem(ctx, userSSOID, itemID)
}
