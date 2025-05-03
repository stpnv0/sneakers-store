package services

import (
	"cart_service/internal/models"
	"cart_service/internal/repository"
	"context"
	"fmt"
	"log"
	"time"
)

// CartCacheAsideService реализует паттерн Cache-Aside для работы с корзиной
type CartCacheAsideService struct {
	pgRepo    *repository.PostgresRepository
	redisRepo *repository.RedisRepository
	logger    *log.Logger
	cacheTTL  time.Duration
}

// NewCartCacheAsideService создает новый сервис корзины с Cache-Aside паттерном
func NewCartCacheAsideService(
	pgRepo *repository.PostgresRepository,
	redisRepo *repository.RedisRepository,
	logger *log.Logger,
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
	// 1. Пробуем получить из кэша
	cart, err := s.redisRepo.GetCart(ctx, userSSOID)
	if err == nil {
		// Кэш-хит: возвращаем данные из кэша
		if len(cart.Items) > 0 {
			s.logger.Printf("[INFO] Cache hit for user cart %d", userSSOID)
			return cart, nil
		}
		s.logger.Printf("[INFO] Cache hit for user cart %d but cart is empty, checking DB", userSSOID)
	}

	// 2. Кэш-промах: загружаем из PostgreSQL
	s.logger.Printf("[INFO] Cache miss for user cart %d, loading from DB", userSSOID)
	cart, err = s.pgRepo.GetCart(ctx, userSSOID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart from PostgreSQL: %w", err)
	}

	// 3. Заполняем кэш новыми данными
	if err = s.redisRepo.SetCart(ctx, userSSOID, cart, s.cacheTTL); err != nil {
		s.logger.Printf("[WARNING] Failed to cache cart for user %d: %v", userSSOID, err)
		// Продолжаем работу даже при ошибке кэширования
	}

	return cart, nil
}

// AddItemToCart добавляет товар в корзину с обновлением БД и кэша
func (s *CartCacheAsideService) AddItemToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error {
	// 1. Добавляем товар в основную БД
	item := &models.CartItem{
		UserSSOID: userSSOID,
		SneakerID: sneakerID,
		Quantity:  quantity,
		AddedAt:   time.Now(),
	}

	err := s.pgRepo.AddCartItem(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to add item to PostgreSQL: %w", err)
	}

	// 2. Точечно обновляем кэш - добавляем элемент вместо инвалидации всей корзины
	// Это оптимизация: добавляем товар напрямую в Redis вместо полной инвалидации
	if err = s.redisRepo.AddToCartItem(ctx, *item); err != nil {
		s.logger.Printf("[WARNING] Failed to add item to cache for user %d: %v", userSSOID, err)
		// При ошибке добавления в кэш - инвалидируем кэш для консистентности
		if invalidateErr := s.redisRepo.InvalidateCart(ctx, userSSOID); invalidateErr != nil {
			s.logger.Printf("[WARNING] Also failed to invalidate cache for user %d: %v", userSSOID, invalidateErr)
		}
	}

	return nil
}

// UpdateCartItemQuantity обновляет количество товара в корзине
func (s *CartCacheAsideService) UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, quantity int) error {
	// 1. Обновляем в PostgreSQL
	err := s.pgRepo.UpdateCartItemQuantity(ctx, userSSOID, itemID, quantity)
	if err != nil {
		return fmt.Errorf("failed to update item quantity in PostgreSQL: %w", err)
	}

	// 2. Точечно обновляем кэш
	if err = s.redisRepo.UpdateCartItemQuantity(ctx, userSSOID, itemID, quantity); err != nil {
		s.logger.Printf("[WARNING] Failed to update item quantity in cache for user %d: %v", userSSOID, err)
		// При ошибке обновления в кэше - инвалидируем кэш для консистентности
		if invalidateErr := s.redisRepo.InvalidateCart(ctx, userSSOID); invalidateErr != nil {
			s.logger.Printf("[WARNING] Also failed to invalidate cache for user %d: %v", userSSOID, invalidateErr)
		}
	}

	return nil
}

// RemoveCartItem удаляет товар из корзины
func (s *CartCacheAsideService) RemoveCartItem(ctx context.Context, userSSOID int, itemID string) error {
	// 1. Удаляем из PostgreSQL
	err := s.pgRepo.RemoveCartItem(ctx, userSSOID, itemID)
	if err != nil {
		return fmt.Errorf("failed to remove item from PostgreSQL: %w", err)
	}

	// 2. Точечно удаляем из кэша (вместо полной инвалидации)
	if err = s.redisRepo.RemoveFromCart(ctx, userSSOID, itemID); err != nil {
		s.logger.Printf("[WARNING] Failed to remove item from cache for user %d: %v", userSSOID, err)
		// При ошибке удаления из кэша - инвалидируем кэш для консистентности
		if invalidateErr := s.redisRepo.InvalidateCart(ctx, userSSOID); invalidateErr != nil {
			s.logger.Printf("[WARNING] Also failed to invalidate cache for user %d: %v", userSSOID, invalidateErr)
		}
	}

	return nil
}

// ClearCart очищает корзину пользователя
func (s *CartCacheAsideService) ClearCart(ctx context.Context, userSSOID int) error {
	// 1. Очищаем в PostgreSQL
	err := s.pgRepo.ClearCart(ctx, userSSOID)
	if err != nil {
		return fmt.Errorf("failed to clear cart in PostgreSQL: %w", err)
	}

	// 2. Полностью инвалидируем кэш - здесь это логично, так как очищаем всю корзину
	if err = s.redisRepo.InvalidateCart(ctx, userSSOID); err != nil {
		s.logger.Printf("[WARNING] Failed to invalidate cache for user %d: %v", userSSOID, err)
	}

	return nil
}

// === Методы интерфейса CartService для совместимости ===

// AddToCart - метод для совместимости с интерфейсом CartService
func (s *CartCacheAsideService) AddToCart(ctx context.Context, userSSOID, sneakerID, quantity int) error {
	return s.AddItemToCart(ctx, userSSOID, sneakerID, quantity)
}

// RemoveFromCart - метод для совместимости с интерфейсом CartService
func (s *CartCacheAsideService) RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error {
	return s.RemoveCartItem(ctx, userSSOID, itemID)
}
