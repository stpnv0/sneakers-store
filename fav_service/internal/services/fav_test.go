package services

import (
	"context"
	"errors"
	"log/slog"
	"io"
	"testing"
	"time"

	"fav_service/internal/models"
	"fav_service/internal/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func newTestService(repo *mocks.MockFavouritesRepo, cache *mocks.MockCacheRepo) *FavService {
	return NewFavService(repo, cache, 24*time.Hour, testLogger)
}

// --- GetAllFavourites ---

func TestGetAllFavourites_CacheHit(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	expected := []models.Favourite{
		{ID: 1, UserSSOID: 42, SneakerID: 100},
	}
	cache.On("GetAllFavourites", mock.Anything, 42).Return(expected, nil)

	result, err := svc.GetAllFavourites(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	repo.AssertNotCalled(t, "GetAllFavourites")
	cache.AssertExpectations(t)
}

func TestGetAllFavourites_EmptyCacheHit(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	cache.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite{}, nil)

	result, err := svc.GetAllFavourites(context.Background(), 42)
	require.NoError(t, err)
	assert.Empty(t, result)

	repo.AssertNotCalled(t, "GetAllFavourites")
	cache.AssertExpectations(t)
}

func TestGetAllFavourites_CacheMiss_LoadsFromDB(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	expected := []models.Favourite{
		{ID: 1, UserSSOID: 42, SneakerID: 100},
	}

	cache.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite(nil), errors.New("cache miss"))
	repo.On("GetAllFavourites", mock.Anything, 42).Return(expected, nil)
	cache.On("SetFavourites", mock.Anything, 42, expected, 24*time.Hour).Return(nil)

	result, err := svc.GetAllFavourites(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetAllFavourites_DBError(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	cache.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite(nil), errors.New("cache miss"))
	repo.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite(nil), errors.New("db error"))

	result, err := svc.GetAllFavourites(context.Background(), 42)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
}

func TestGetAllFavourites_CacheSetError_StillReturnsData(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	expected := []models.Favourite{
		{ID: 1, UserSSOID: 42, SneakerID: 100},
	}

	cache.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite(nil), errors.New("miss"))
	repo.On("GetAllFavourites", mock.Anything, 42).Return(expected, nil)
	cache.On("SetFavourites", mock.Anything, 42, expected, 24*time.Hour).Return(errors.New("cache write fail"))

	result, err := svc.GetAllFavourites(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// --- AddToFavourite ---

func TestAddToFavourite_AlreadyExists(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("IsFavourite", mock.Anything, 42, 100).Return(true, nil)

	err := svc.AddToFavourite(context.Background(), 42, 100)
	require.NoError(t, err)

	repo.AssertNotCalled(t, "AddToFavourite")
}

func TestAddToFavourite_Success(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("IsFavourite", mock.Anything, 42, 100).Return(false, nil)
	repo.On("AddToFavourite", mock.Anything, 42, 100).Return(nil)
	cache.On("InvalidateFavourites", mock.Anything, 42).Return(nil)
	repo.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite{{SneakerID: 100, UserSSOID: 42}}, nil)
	cache.On("SetFavourites", mock.Anything, 42, mock.Anything, 24*time.Hour).Return(nil)

	err := svc.AddToFavourite(context.Background(), 42, 100)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestAddToFavourite_RepoError(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("IsFavourite", mock.Anything, 42, 100).Return(false, nil)
	repo.On("AddToFavourite", mock.Anything, 42, 100).Return(errors.New("insert failed"))

	err := svc.AddToFavourite(context.Background(), 42, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

// --- RemoveFromFavourite ---

func TestRemoveFromFavourite_Success(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("RemoveFromFavourite", mock.Anything, 42, 100).Return(nil)
	cache.On("InvalidateFavourites", mock.Anything, 42).Return(nil)
	repo.On("GetAllFavourites", mock.Anything, 42).Return([]models.Favourite{}, nil)
	cache.On("SetFavourites", mock.Anything, 42, mock.Anything, 24*time.Hour).Return(nil)

	err := svc.RemoveFromFavourite(context.Background(), 42, 100)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRemoveFromFavourite_RepoError(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("RemoveFromFavourite", mock.Anything, 42, 100).Return(errors.New("delete failed"))

	err := svc.RemoveFromFavourite(context.Background(), 42, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

// --- IsFavourite ---

func TestIsFavourite_True(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("IsFavourite", mock.Anything, 42, 100).Return(true, nil)

	result, err := svc.IsFavourite(context.Background(), 42, 100)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestIsFavourite_False(t *testing.T) {
	repo := new(mocks.MockFavouritesRepo)
	cache := new(mocks.MockCacheRepo)
	svc := newTestService(repo, cache)

	repo.On("IsFavourite", mock.Anything, 42, 100).Return(false, nil)

	result, err := svc.IsFavourite(context.Background(), 42, 100)
	require.NoError(t, err)
	assert.False(t, result)
}

// --- ParseIDsString ---

func TestParseIDsString_Valid(t *testing.T) {
	svc := newTestService(nil, nil)

	ids, err := svc.ParseIDsString("1,2,3")
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, ids)
}

func TestParseIDsString_Empty(t *testing.T) {
	svc := newTestService(nil, nil)

	ids, err := svc.ParseIDsString("")
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestParseIDsString_Invalid(t *testing.T) {
	svc := newTestService(nil, nil)

	_, err := svc.ParseIDsString("1,abc,3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ID format")
}
