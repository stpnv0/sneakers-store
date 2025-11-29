package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fav_service/internal/models"

	"github.com/go-redis/redis/v8"
)

type redisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *redisRepo {
	return &redisRepo{
		client: client,
	}
}

func getKey(userSSOID int) string {
	return fmt.Sprintf("fav:%d", userSSOID)
}

func (r *redisRepo) SetFavourites(ctx context.Context, userSSOID int, favourites []models.Favourite, ttl time.Duration) error {
	key := getKey(userSSOID)

	pipe := r.client.Pipeline()

	pipe.Del(ctx, key)

	if len(favourites) > 0 {
		interfaceIDs := make([]interface{}, len(favourites))
		for i, item := range favourites {
			interfaceIDs[i] = item.SneakerID
		}
		pipe.SAdd(ctx, key, interfaceIDs...)
	}

	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	} else {
		pipe.Expire(ctx, key, 24*time.Hour)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to set cache for key %s: %v\n", key, err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	fmt.Printf("INFO: Cache set for key %s with %d items\n", key, len(favourites))
	return nil
}

func (r *redisRepo) InvalidateFavourites(ctx context.Context, userSSOID int) error {
	key := getKey(userSSOID)

	return r.client.Del(ctx, key).Err()
}
func (r *redisRepo) GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error) {
	key := getKey(userSSOID)

	result, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		fmt.Printf("ERROR: Failed to get cache for key %s: %v\n", key, err)
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	fmt.Printf("INFO: Cache retrieved for key %s with %d items\n", key, len(result))

	favourites := make([]models.Favourite, len(result))
	for i, idStr := range result {
		sneakerID, _ := strconv.Atoi(idStr)
		favourites[i] = models.Favourite{
			SneakerID: sneakerID,
			UserSSOID: userSSOID,
			// ID and AddedAt are missing in cache, will be zero values
		}
	}

	return favourites, nil
}
