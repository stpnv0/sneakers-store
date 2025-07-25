package app

import (
	"context"
	domain "product_service/internal/domain/model"
	"time"
)

type ProductPostgres interface {
	GetSneakerByID(ctx context.Context, id int64) (*domain.Sneaker, error)
	AddSneaker(ctx context.Context, sneaker *domain.Sneaker) (int64, error)
	GetAllSneakers(ctx context.Context, limit, offset uint64) ([]*domain.Sneaker, error)
	GetSneakersByIDs(ctx context.Context, ids []int64) ([]*domain.Sneaker, error)
	DeleteSneaker(ctx context.Context, id int64) error
	UpdateImageKey(ctx context.Context, id int64, imageKey string) error
}

type ProductCache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type FileStore interface {
	GenerateUploadURL(ctx context.Context, key string, contentType string) (string, error)
}
