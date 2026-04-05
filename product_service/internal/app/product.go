package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strings"

	"product_service/internal/model"
	"product_service/internal/repository"
	"time"

	"github.com/google/uuid"
)

// validImageKeyRe соответствует ключам вида "products/<uuid>.<ext>" или "products/<num>.jpg".
var validImageKeyRe = regexp.MustCompile(`^products/[a-zA-Z0-9_-]+\.[a-zA-Z0-9]+$`)

type Service struct {
	repo      ProductPostgres
	cache     ProductCache
	fileStore FileStore
	log       *slog.Logger
	cacheTTL  time.Duration
}

func NewService(repo ProductPostgres, cache ProductCache, fileStore FileStore, log *slog.Logger, cacheTTL time.Duration) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		fileStore: fileStore,
		log:       log,
		cacheTTL:  cacheTTL,
	}
}

func productKeyL1(id int64) string {
	return fmt.Sprintf("product:%d", id)
}

func productsKeyL2(limit, offset uint64) string {
	return fmt.Sprintf("products:list:limit:%d:offset:%d", limit, offset)
}

func (s *Service) GetSneakerByID(ctx context.Context, id int64) (*model.Sneaker, error) {
	const op = "app.Service.GetSneakerByID"
	log := s.log.With(slog.String("op", op), slog.Int64("id", id))

	//cache
	key := productKeyL1(id)
	var cachedSneaker model.Sneaker
	err := s.cache.Get(ctx, key, &cachedSneaker)
	if err == nil {
		log.Info("cache hit")
		return &cachedSneaker, nil
	}

	if !errors.Is(err, repository.ErrNotFound) {
		log.Error("failed to get from cache", slog.String("error", err.Error()))
	}
	log.Info("cache miss")

	//db
	dbSneaker, err := s.repo.GetSneakerByID(ctx, id)
	if err != nil {
		log.Error("failed to get from db", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//fill the cache
	if setErr := s.cache.Set(ctx, key, dbSneaker, s.cacheTTL); setErr != nil {
		log.Error("failed to set cache", slog.String("error", setErr.Error()))
	}

	return dbSneaker, nil
}

func (s *Service) DeleteSneaker(ctx context.Context, id int64) error {
	const op = "app.Service.DeleteSneaker"
	log := s.log.With(slog.String("op", op), slog.Int64("id", id))

	// db
	if err := s.repo.DeleteSneaker(ctx, id); err != nil {
		log.Error("failed to delete from db", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("deleted from db")

	// Invalidate L1 Cache
	if err := s.cache.Delete(ctx, productKeyL1(id)); err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Error("failed to invalidate L1 cache", slog.String("error", err.Error()))
	}

	// Invalidate L2 Cache (все варианты пагинации)
	if err := s.cache.DeleteByPrefix(ctx, "products:list:"); err != nil {
		log.Error("failed to invalidate L2 cache", slog.String("error", err.Error()))
	}

	return nil
}

func (s *Service) GetSneakersByIDs(ctx context.Context, ids []int64) ([]*model.Sneaker, error) {
	const op = "app.Service.GetSneakersByIDs"
	log := s.log.With(slog.String("op", op))

	sneakers, err := s.repo.GetSneakersByIDs(ctx, ids)
	if err != nil {
		log.Error("failed to get sneakers by ids from db", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sneakers, nil
}

func (s *Service) GetAllSneakers(ctx context.Context, limit, offset uint64) ([]*model.Sneaker, error) {
	const op = "app.Service.GetAllSneakers"
	log := s.log.With(slog.String("op", op))

	// Cache
	key := productsKeyL2(limit, offset)
	var cachedSneakers []*model.Sneaker
	err := s.cache.Get(ctx, key, &cachedSneakers)
	if err == nil {
		log.Info("list cache hit")
		return cachedSneakers, nil
	}
	log.Info("list cache miss")

	//db
	dbSneakers, err := s.repo.GetAllSneakers(ctx, limit, offset)
	if err != nil {
		log.Error("failed to get list from db", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// fill L2 Cache
	if len(dbSneakers) > 0 {
		if setErr := s.cache.Set(ctx, key, dbSneakers, s.cacheTTL/2); setErr != nil { // Кэш списков живет меньше
			log.Error("failed to set list cache", slog.String("error", setErr.Error()))
		}
	}

	return dbSneakers, nil
}

func (s *Service) AddSneaker(ctx context.Context, sneaker *model.Sneaker) (int64, error) {
	const op = "app.Service.AddSneaker"
	log := s.log.With(slog.String("op", op))

	//db
	id, err := s.repo.AddSneaker(ctx, sneaker)
	if err != nil {
		log.Error("failed to add sneaker to db", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	sneaker.Id = id
	log.Info("sneaker added", slog.Int64("id", id))

	//invalidate L2 cache
	if delErr := s.cache.DeleteByPrefix(ctx, "products:list:"); delErr != nil {
		log.Error("failed to invalidate list cache", slog.String("error", delErr.Error()))
	}

	return id, nil
}

func (s *Service) GenerateUploadURL(ctx context.Context, originalFilename string, contentType string) (uploadURL string, fileKey string, err error) {
	const op = "app.Service.GenerateUploadURL"
	log := s.log.With(slog.String("op", op))

	fileUUID := uuid.New().String()
	key := fmt.Sprintf("products/%s%s", fileUUID, path.Ext(originalFilename))

	log.Info("generating upload url", slog.String("key", key))

	uploadURL, err = s.fileStore.GenerateUploadURL(ctx, key, contentType)
	if err != nil {
		log.Error("failed to generate url", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return uploadURL, key, nil
}

// ValidateImageKey проверяет, что ключ изображения имеет ожидаемый формат.
func ValidateImageKey(key string) error {
	if !strings.HasPrefix(key, "products/") {
		return fmt.Errorf("image key must start with 'products/'")
	}
	if len(key) > 256 {
		return fmt.Errorf("image key too long (max 256 chars)")
	}
	if !validImageKeyRe.MatchString(key) {
		return fmt.Errorf("image key contains invalid characters")
	}
	return nil
}

func (s *Service) UpdateProductImage(ctx context.Context, productID int64, imageKey string) error {
	const op = "app.Service.UpdateProductImage"
	log := s.log.With(slog.String("op", op), slog.Int64("productID", productID))

	if err := ValidateImageKey(imageKey); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err := s.repo.UpdateImageKey(ctx, productID, imageKey)
	if err != nil {
		log.Error("failed to update product image key in db", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.Delete(ctx, productKeyL1(productID)); err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Warn("failed to invalidate L1 cache for product", slog.String("error", err.Error()))
	}
	if err := s.cache.DeleteByPrefix(ctx, "products:list:"); err != nil {
		log.Warn("failed to invalidate L2 cache", slog.String("error", err.Error()))
	}

	log.Info("product image updated successfully", slog.String("imageKey", imageKey))
	return nil
}
