package app

import (
	"context"
	"product_service/internal/model"
	"time"
)

type ProductPostgres interface {
	GetSneakerByID(ctx context.Context, id int64) (*model.Sneaker, error)
	AddSneaker(ctx context.Context, sneaker *model.Sneaker) (int64, error)
	GetAllSneakers(ctx context.Context, limit, offset uint64) ([]*model.Sneaker, error)
	GetSneakersByIDs(ctx context.Context, ids []int64) ([]*model.Sneaker, error)
	DeleteSneaker(ctx context.Context, id int64) error
	UpdateImageKey(ctx context.Context, id int64, imageKey string) error
}

type ProductCache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
}

type FileStore interface {
	GenerateUploadURL(ctx context.Context, key string, contentType string) (string, error)
}
