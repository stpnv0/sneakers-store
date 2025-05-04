package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

func (r *redisRepo) SetFavourites(ctx context.Context, userSSOID int, sneakerIDs []int, ttl time.Duration) error {
	key := getKey(userSSOID)

	pipe := r.client.Pipeline()

	pipe.Del(ctx, key)

	if len(sneakerIDs) > 0 {
		interfaceIDs := make([]interface{}, len(sneakerIDs))
		for i, id := range sneakerIDs {
			interfaceIDs[i] = id
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

	fmt.Printf("INFO: Cache set for key %s with %d items\n", key, len(sneakerIDs))
	return nil
}

func (r *redisRepo) InvalidateFavourites(ctx context.Context, userSSOID int) error {
	key := getKey(userSSOID)

	return r.client.Del(ctx, key).Err()
}
func (r *redisRepo) GetAllFavourites(ctx context.Context, userSSOID int) ([]int, error) {
	key := getKey(userSSOID)

	result, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		fmt.Printf("ERROR: Failed to get cache for key %s: %v\n", key, err)
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	fmt.Printf("INFO: Cache retrieved for key %s with %d items\n", key, len(result))

	sneakerIDs := make([]int, len(result))
	for i, idStr := range result {
		sneakerIDs[i], _ = strconv.Atoi(idStr)
	}

	return sneakerIDs, nil
}
