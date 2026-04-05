package repository

import (
	"cart_service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ErrCacheMiss = errors.New("cache miss")

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

func getCartKey(userSSOID int) string {
	return fmt.Sprintf("cart:%d", userSSOID)
}

// AddToCart - старый метод для обратной совместимости
func (r *RedisRepository) AddToCart(ctx context.Context, userSSOID, sneakerID int) error {
	itemID := fmt.Sprintf("%d%d%d", userSSOID, sneakerID, time.Now().UnixNano())

	var CartItem = models.CartItem{
		ID:           itemID,
		UserSSOID:    userSSOID,
		SneakerID:    sneakerID,
		Quantity:     1,
		AddedAt:      time.Now(),
		Synchronized: false,
	}

	return r.AddToCartItem(ctx, CartItem)
}

// AddToCartItem - новый метод, который принимает объект CartItem
func (r *RedisRepository) AddToCartItem(ctx context.Context, item models.CartItem) error {
	key := getCartKey(item.UserSSOID)

	// Проверяем, что у объекта есть ID, если нет - генерируем
	if item.ID == "" {
		item.ID = fmt.Sprintf("%d%d%d", item.UserSSOID, item.SneakerID, time.Now().UnixNano())
	}

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal cart item: %w", err)
	}

	err = r.client.HSet(ctx, key, item.ID, itemJSON).Err()
	if err != nil {
		return fmt.Errorf("set cart item in redis: %w", err)
	}

	if err := r.client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("set cart ttl: %w", err)
	}

	return nil
}

func (r *RedisRepository) GetCart(ctx context.Context, userSSOID int) (*models.Cart, error) {
	key := getCartKey(userSSOID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("check cart key existence: %w", err)
	}
	if exists == 0 {
		return nil, ErrCacheMiss
	}

	values, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get cart from redis: %w", err)
	}

	cart := models.Cart{
		UserSSOID: userSSOID,
		Items:     []models.CartItem{},
		UpdatedAt: time.Now(),
	}

	for _, val := range values {
		var item models.CartItem
		if err := json.Unmarshal([]byte(val), &item); err != nil {
			return nil, fmt.Errorf("unmarshal cart item from redis: %w", err)
		}

		cart.Items = append(cart.Items, item)
	}

	return &cart, nil
}

// SetCart сохраняет всю корзину в Redis
func (r *RedisRepository) SetCart(ctx context.Context, userSSOID int, cart *models.Cart, ttl time.Duration) error {
	key := getCartKey(userSSOID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete old cart from redis: %w", err)
	}

	// Проходим по всем элементам и добавляем их в корзину
	for _, item := range cart.Items {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshal cart item: %w", err)
		}

		err = r.client.HSet(ctx, key, item.ID, itemJSON).Err()
		if err != nil {
			return fmt.Errorf("set cart item in redis: %w", err)
		}
	}

	// Устанавливаем TTL для корзины
	expiry := ttl
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}
	if err := r.client.Expire(ctx, key, expiry).Err(); err != nil {
		return fmt.Errorf("set cart ttl: %w", err)
	}

	return nil
}

// InvalidateCart удаляет корзину из кэша
func (r *RedisRepository) InvalidateCart(ctx context.Context, userSSOID int) error {
	key := getCartKey(userSSOID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisRepository) GetItemFromCart(ctx context.Context, userSSOID int, itemID string) (*models.CartItem, error) {
	key := getCartKey(userSSOID)

	itemJSON, err := r.client.HGet(ctx, key, itemID).Result()
	if err != nil {
		return nil, fmt.Errorf("get item from redis: %w", err)
	}

	var item models.CartItem
	if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
		return nil, fmt.Errorf("unmarshal item from redis: %w", err)
	}

	return &item, nil
}

func (r *RedisRepository) UpdateCartItemQuantity(ctx context.Context, userSSOID int, itemID string, newQuantity int) error {
	key := getCartKey(userSSOID)

	itemJSON, err := r.client.HGet(ctx, key, itemID).Result()
	if err != nil {
		return fmt.Errorf("get cart item from redis: %w", err)
	}

	var item models.CartItem
	if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
		return fmt.Errorf("unmarshal cart item from redis: %w", err)
	}

	item.Quantity = newQuantity
	item.Synchronized = false

	updatedJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal updated item: %w", err)
	}

	return r.client.HSet(ctx, key, itemID, updatedJSON).Err()
}

func (r *RedisRepository) RemoveFromCart(ctx context.Context, userSSOID int, itemID string) error {
	key := getCartKey(userSSOID)
	return r.client.HDel(ctx, key, itemID).Err()
}
