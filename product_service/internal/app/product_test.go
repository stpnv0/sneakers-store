package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"product_service/internal/model"
	"product_service/internal/app/mocks"
	"product_service/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func newTestService(repo *mocks.MockProductPostgres, cache *mocks.MockProductCache, fs *mocks.MockFileStore) *Service {
	return NewService(repo, cache, fs, testLogger, 10*time.Minute)
}

// --- GetSneakerByID ---

func TestGetSneakerByID_CacheHit(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	sneaker := &model.Sneaker{Id: 1, Title: "Nike", Price: 10000}

	cache.On("Get", mock.Anything, "product:1", mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(2).(*model.Sneaker)
			*dest = *sneaker
		}).Return(nil)

	result, err := svc.GetSneakerByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, sneaker.Title, result.Title)
	repo.AssertNotCalled(t, "GetSneakerByID")
}

func TestGetSneakerByID_CacheMiss(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	sneaker := &model.Sneaker{Id: 1, Title: "Nike", Price: 10000}

	cache.On("Get", mock.Anything, "product:1", mock.Anything).Return(repository.ErrNotFound)
	repo.On("GetSneakerByID", mock.Anything, int64(1)).Return(sneaker, nil)
	cache.On("Set", mock.Anything, "product:1", sneaker, 10*time.Minute).Return(nil)

	result, err := svc.GetSneakerByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Nike", result.Title)
	repo.AssertExpectations(t)
}

func TestGetSneakerByID_DBError(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	cache.On("Get", mock.Anything, "product:1", mock.Anything).Return(repository.ErrNotFound)
	repo.On("GetSneakerByID", mock.Anything, int64(1)).Return((*model.Sneaker)(nil), errors.New("db error"))

	result, err := svc.GetSneakerByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- AddSneaker ---

func TestAddSneaker_Success(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	sneaker := &model.Sneaker{Title: "Nike", Price: 10000}
	repo.On("AddSneaker", mock.Anything, sneaker).Return(int64(42), nil)
	cache.On("DeleteByPrefix", mock.Anything, "products:list:").Return(nil)

	id, err := svc.AddSneaker(context.Background(), sneaker)
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)
}

func TestAddSneaker_DBError(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	sneaker := &model.Sneaker{Title: "Nike", Price: 10000}
	repo.On("AddSneaker", mock.Anything, sneaker).Return(int64(0), errors.New("insert error"))

	id, err := svc.AddSneaker(context.Background(), sneaker)
	assert.Error(t, err)
	assert.Equal(t, int64(0), id)
}

// --- DeleteSneaker ---

func TestDeleteSneaker_Success(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	repo.On("DeleteSneaker", mock.Anything, int64(1)).Return(nil)
	cache.On("Delete", mock.Anything, "product:1").Return(nil)
	cache.On("DeleteByPrefix", mock.Anything, "products:list:").Return(nil)

	err := svc.DeleteSneaker(context.Background(), 1)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

// --- GetAllSneakers ---

func TestGetAllSneakers_CacheMiss(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	sneakers := []*model.Sneaker{{Id: 1, Title: "Nike"}}
	cache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(repository.ErrNotFound)
	repo.On("GetAllSneakers", mock.Anything, uint64(20), uint64(0)).Return(sneakers, nil)
	cache.On("Set", mock.Anything, mock.Anything, sneakers, 5*time.Minute).Return(nil)

	result, err := svc.GetAllSneakers(context.Background(), 20, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// --- UpdateProductImage ---

func TestUpdateProductImage_Success(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	repo.On("UpdateImageKey", mock.Anything, int64(1), "products/img.jpg").Return(nil)
	cache.On("Delete", mock.Anything, "product:1").Return(nil)
	cache.On("DeleteByPrefix", mock.Anything, "products:list:").Return(nil)

	err := svc.UpdateProductImage(context.Background(), 1, "products/img.jpg")
	require.NoError(t, err)
}

func TestUpdateProductImage_NotFound(t *testing.T) {
	repo := new(mocks.MockProductPostgres)
	cache := new(mocks.MockProductCache)
	fs := new(mocks.MockFileStore)
	svc := newTestService(repo, cache, fs)

	repo.On("UpdateImageKey", mock.Anything, int64(1), "products/img.jpg").Return(repository.ErrNotFound)

	err := svc.UpdateProductImage(context.Background(), 1, "products/img.jpg")
	assert.Error(t, err)
}
