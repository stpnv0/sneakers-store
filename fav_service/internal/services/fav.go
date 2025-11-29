package services

import (
	"context"
	"fmt"
	"log"
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
}

func NewFavService(repo FavouritesRepo, cache CacheRepo, ttl time.Duration) *FavService {
	return &FavService{
		repo:     repo,
		cache:    cache,
		cacheTTL: ttl,
	}
}

func (s *FavService) GetAllFavourites(ctx context.Context, userSSOID int) ([]models.Favourite, error) {
	// Пытаемся получить данные из кэша
	favourites, err := s.cache.GetAllFavourites(ctx, userSSOID)
	if err == nil && len(favourites) > 0 {
		log.Printf("INFO: Cache hit for user %d", userSSOID)
		return favourites, nil
	}

	log.Printf("INFO: Cache miss for user %d, loading from DB", userSSOID)

	// Если данных нет в кэше, получаем их из базы данных
	favourites, err = s.repo.GetAllFavourites(ctx, userSSOID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favourites from DB: %w", err)
	}

	// Сохраняем данные в кэш
	if err := s.cache.SetFavourites(ctx, userSSOID, favourites, s.cacheTTL); err != nil {
		log.Printf("WARNING: Failed to cache favourites for user %d: %v", userSSOID, err)
	} else {
		log.Printf("INFO: Cache set for key fav:%d with %d items", userSSOID, len(favourites))
	}

	return favourites, nil
}

func (s *FavService) AddToFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	// Проверяем, не добавлен ли уже товар в избранное
	exists, err := s.repo.IsFavourite(ctx, userSSOID, sneakerID)
	if err != nil {
		return fmt.Errorf("failed to check if favourite exists: %w", err)
	}

	if exists {
		return nil // Товар уже в избранном
	}

	// Добавляем товар в избранное
	if err := s.repo.AddToFavourite(ctx, userSSOID, sneakerID); err != nil {
		return fmt.Errorf("failed to add to favourites: %w", err)
	}

	// Инвалидируем кэш
	if err := s.cache.InvalidateFavourites(ctx, userSSOID); err != nil {
		log.Printf("WARNING: Failed to invalidate cache for user %d: %v", userSSOID, err)
	} else {
		log.Printf("INFO: Cache invalidated for key fav:%d", userSSOID)
	}

	// Сразу обновляем кэш новыми данными
	favourites, err := s.repo.GetAllFavourites(ctx, userSSOID)
	if err != nil {
		log.Printf("WARNING: Failed to get updated favourites for user %d: %v", userSSOID, err)
	} else {
		if err := s.cache.SetFavourites(ctx, userSSOID, favourites, s.cacheTTL); err != nil {
			log.Printf("WARNING: Failed to update cache for user %d: %v", userSSOID, err)
		} else {
			log.Printf("INFO: Cache updated for key fav:%d with %d items", userSSOID, len(favourites))
		}
	}

	return nil
}

func (s *FavService) RemoveFromFavourite(ctx context.Context, userSSOID, sneakerID int) error {
	// Удаляем товар из избранного
	if err := s.repo.RemoveFromFavourite(ctx, userSSOID, sneakerID); err != nil {
		return fmt.Errorf("failed to remove from favourites: %w", err)
	}

	// Инвалидируем кэш
	if err := s.cache.InvalidateFavourites(ctx, userSSOID); err != nil {
		log.Printf("WARNING: Failed to invalidate cache for user %d: %v", userSSOID, err)
	} else {
		log.Printf("INFO: Cache invalidated for key fav:%d", userSSOID)
	}

	// Сразу обновляем кэш новыми данными
	favourites, err := s.repo.GetAllFavourites(ctx, userSSOID)
	if err != nil {
		log.Printf("WARNING: Failed to get updated favourites for user %d: %v", userSSOID, err)
	} else {
		if err := s.cache.SetFavourites(ctx, userSSOID, favourites, s.cacheTTL); err != nil {
			log.Printf("WARNING: Failed to update cache for user %d: %v", userSSOID, err)
		} else {
			log.Printf("INFO: Cache updated for key fav:%d with %d items", userSSOID, len(favourites))
		}
	}

	return nil
}

func (s *FavService) IsFavourite(ctx context.Context, userSSOID, sneakerID int) (bool, error) {
	return s.repo.IsFavourite(ctx, userSSOID, sneakerID)
}

func (s *FavService) ParseIDsString(idsString string) ([]int, error) {
	if idsString == "" {
		return []int{}, nil
	}

	idStrings := strings.Split(idsString, ",")

	// Создаем массив для результата
	ids := make([]int, 0, len(idStrings))

	// Преобразуем каждую строку в число
	for _, idStr := range idStrings {
		// Пропускаем пустые строки
		if idStr == "" {
			continue
		}

		// Преобразуем строку в число
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			return nil, fmt.Errorf("invalid ID format: %w", err)
		}

		// Добавляем число в результат
		ids = append(ids, id)
	}

	return ids, nil
}
