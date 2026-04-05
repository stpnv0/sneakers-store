package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"fav_service/internal/models"
)

type FavouritesRepo interface {
	GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error)
	AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error
	RemoveFromFavourite(ctx context.Context, userSSOID, sneakerID int) error
	IsFavourite(ctx context.Context, userSSOID, sneakerID int) (bool, error)
	GetByIDs(ctx context.Context, ids []int) ([]models.Favourite, error)
}

type CacheRepo interface {
	GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error)
	InvalidateFavourites(ctx context.Context, userSSOID int) error
	SetFavourites(ctx context.Context, userSSOID int, favourites []models.Favourite, ttl time.Duration) error
}

type FavService struct {
	repo     FavouritesRepo
	cache    CacheRepo
	cacheTTL time.Duration
	log      *slog.Logger
}

func NewFavService(repo FavouritesRepo, cache CacheRepo, ttl time.Duration, log *slog.Logger) *FavService {
	return &FavService{
		repo:     repo,
		cache:    cache,
		cacheTTL: ttl,
		log:      log,
	}
}

func (s *FavService) GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error) {
	const op = "service.GetAllFavourites"

	favourites, err := s.cache.GetAllFavourites(ctx, userSSOID)
	if err == nil {
		s.log.Debug("cache hit", slog.String("op", op), slog.Int("user_id", userSSOID))
		return favourites, nil
	}

	s.log.Debug("cache miss, loading from db", slog.String("op", op), slog.Int("user_id", userSSOID))

	favourites, err = s.repo.GetAllFavourites(ctx, userSSOID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.SetFavourites(ctx, userSSOID, favourites, s.cacheTTL); err != nil {
		s.log.Warn("failed to cache favourites",
			slog.String("op", op),
			slog.Int("user_id", userSSOID),
			slog.String("error", err.Error()),
		)
	}

	return favourites, nil
}

func (s *FavService) AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	const op = "service.AddToFavourite"

	exists, err := s.repo.IsFavourite(ctx, userSSOID, sneakerID)
	if err != nil {
		return fmt.Errorf("%s: check exists: %w", op, err)
	}
	if exists {
		return nil
	}

	if err := s.repo.AddToFavourite(ctx, userSSOID, sneakerID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.refreshCache(ctx, op, userSSOID)
	return nil
}

func (s *FavService) RemoveFromFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	const op = "service.RemoveFromFavourite"

	if err := s.repo.RemoveFromFavourite(ctx, userSSOID, sneakerID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.refreshCache(ctx, op, userSSOID)
	return nil
}

func (s *FavService) IsFavourite(ctx context.Context, userSSOID, sneakerID int) (bool, error) {
	return s.repo.IsFavourite(ctx, userSSOID, sneakerID)
}

func (s *FavService) GetByIDs(ctx context.Context, ids []int) ([]models.Favourite, error) {
	return s.repo.GetByIDs(ctx, ids)
}

func (s *FavService) ParseIDsString(idsString string) ([]int, error) {
	if idsString == "" {
		return []int{}, nil
	}

	parts := strings.Split(idsString, ",")
	ids := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid ID format %q: %w", p, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (s *FavService) refreshCache(ctx context.Context, op string, userSSOID int) {
	if err := s.cache.InvalidateFavourites(ctx, userSSOID); err != nil {
		s.log.Warn("failed to invalidate cache",
			slog.String("op", op), slog.Int("user_id", userSSOID),
			slog.String("error", err.Error()),
		)
	}

	favourites, err := s.repo.GetAllFavourites(ctx, userSSOID)
	if err != nil {
		s.log.Warn("failed to reload favourites for cache",
			slog.String("op", op), slog.Int("user_id", userSSOID),
			slog.String("error", err.Error()),
		)
		return
	}

	if err := s.cache.SetFavourites(ctx, userSSOID, favourites, s.cacheTTL); err != nil {
		s.log.Warn("failed to update cache",
			slog.String("op", op), slog.Int("user_id", userSSOID),
			slog.String("error", err.Error()),
		)
	}
}
