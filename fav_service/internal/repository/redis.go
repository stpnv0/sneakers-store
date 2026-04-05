package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"fav_service/internal/models"

	"github.com/go-redis/redis/v8"
)

var ErrCacheMiss = errors.New("cache miss")

type redisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *redisRepo {
	return &redisRepo{client: client}
}

func getKey(userSSOID int) string {
	return fmt.Sprintf("fav:%d", userSSOID)
}

func (r *redisRepo) SetFavourites(ctx context.Context, userSSOID int, favourites []models.Favourite, ttl time.Duration) error {
	key := getKey(userSSOID)

	expiry := ttl
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)

	if len(favourites) > 0 {
		interfaceIDs := make([]interface{}, len(favourites))
		for i, item := range favourites {
			interfaceIDs[i] = item.SneakerID
		}
		pipe.SAdd(ctx, key, interfaceIDs...)
	} else {
		// Store a sentinel value so Exists returns 1 for empty favourites (cache hit).
		pipe.SAdd(ctx, key, "__empty__")
	}

	pipe.Expire(ctx, key, expiry)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set favourites cache: %w", err)
	}

	return nil
}

func (r *redisRepo) InvalidateFavourites(ctx context.Context, userSSOID int) error {
	return r.client.Del(ctx, getKey(userSSOID)).Err()
}

func (r *redisRepo) GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error) {
	key := getKey(userSSOID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("check favourites key existence: %w", err)
	}
	if exists == 0 {
		return nil, ErrCacheMiss
	}

	result, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get favourites from cache: %w", err)
	}

	favourites := make([]models.Favourite, 0, len(result))
	for _, idStr := range result {
		if idStr == "__empty__" {
			continue
		}
		sneakerID, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("parse sneaker_id from cache %q: %w", idStr, err)
		}
		favourites = append(favourites, models.Favourite{
			SneakerID: sneakerID,
			UserSSOID: userSSOID,
		})
	}

	return favourites, nil
}
