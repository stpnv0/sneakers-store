package services_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"cart_service/internal/models"
	"cart_service/internal/repository"
	"cart_service/internal/services"
	"cart_service/internal/services/mocks"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

const testTTL = 24 * time.Hour

// ---------------------------------------------------------------------------
// GetCart
// ---------------------------------------------------------------------------

func TestGetCart_CacheHit(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	cached := &models.Cart{
		UserSSOID: 1,
		Items:     []models.CartItem{{ID: "a", SneakerID: 10, Quantity: 2}},
	}
	cache.On("GetCart", mock.Anything, 1).Return(cached, nil)

	result, err := svc.GetCart(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, cached, result)
	repo.AssertNotCalled(t, "GetCart")
}

func TestGetCart_CacheHitEmpty(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	emptyCart := &models.Cart{UserSSOID: 1, Items: []models.CartItem{}}
	cache.On("GetCart", mock.Anything, 1).Return(emptyCart, nil)

	result, err := svc.GetCart(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, result.Items)
	repo.AssertNotCalled(t, "GetCart")
}

func TestGetCart_CacheMiss_LoadsFromDB(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	cache.On("GetCart", mock.Anything, 1).Return(nil, repository.ErrCacheMiss)

	dbCart := &models.Cart{
		UserSSOID: 1,
		Items:     []models.CartItem{{ID: "b", SneakerID: 20, Quantity: 1}},
	}
	repo.On("GetCart", mock.Anything, 1).Return(dbCart, nil)
	cache.On("SetCart", mock.Anything, 1, dbCart, testTTL).Return(nil)

	result, err := svc.GetCart(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, dbCart, result)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestGetCart_CacheMiss_SetCacheFails_StillSucceeds(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	cache.On("GetCart", mock.Anything, 1).Return(nil, repository.ErrCacheMiss)

	dbCart := &models.Cart{UserSSOID: 1, Items: []models.CartItem{}}
	repo.On("GetCart", mock.Anything, 1).Return(dbCart, nil)
	cache.On("SetCart", mock.Anything, 1, dbCart, testTTL).Return(errors.New("redis down"))

	result, err := svc.GetCart(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, dbCart, result)
}

func TestGetCart_DBError(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	cache.On("GetCart", mock.Anything, 1).Return(nil, repository.ErrCacheMiss)
	repo.On("GetCart", mock.Anything, 1).Return(nil, errors.New("db connection lost"))

	result, err := svc.GetCart(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db connection lost")
}

// ---------------------------------------------------------------------------
// AddToCart
// ---------------------------------------------------------------------------

func TestAddToCart_Success(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("AddCartItem", mock.Anything, mock.AnythingOfType("*models.CartItem")).Return(nil)
	cache.On("AddToCartItem", mock.Anything, mock.AnythingOfType("models.CartItem")).Return(nil)

	err := svc.AddToCart(context.Background(), 1, 10, 2)
	require.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestAddToCart_RepoError(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("AddCartItem", mock.Anything, mock.AnythingOfType("*models.CartItem")).
		Return(errors.New("duplicate key"))

	err := svc.AddToCart(context.Background(), 1, 10, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
	cache.AssertNotCalled(t, "AddToCartItem")
}

func TestAddToCart_CacheUpdateFail_Invalidates(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("AddCartItem", mock.Anything, mock.AnythingOfType("*models.CartItem")).Return(nil)
	cache.On("AddToCartItem", mock.Anything, mock.AnythingOfType("models.CartItem")).
		Return(errors.New("redis error"))
	cache.On("InvalidateCart", mock.Anything, 1).Return(nil)

	err := svc.AddToCart(context.Background(), 1, 10, 2)
	require.NoError(t, err)
	cache.AssertCalled(t, "InvalidateCart", mock.Anything, 1)
}

// ---------------------------------------------------------------------------
// RemoveFromCart
// ---------------------------------------------------------------------------

func TestRemoveFromCart_Success(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("RemoveCartItem", mock.Anything, 1, "item-1").Return(nil)
	cache.On("RemoveFromCart", mock.Anything, 1, "item-1").Return(nil)

	err := svc.RemoveFromCart(context.Background(), 1, "item-1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestRemoveFromCart_RepoError(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("RemoveCartItem", mock.Anything, 1, "item-1").Return(errors.New("not found"))

	err := svc.RemoveFromCart(context.Background(), 1, "item-1")
	require.Error(t, err)
	cache.AssertNotCalled(t, "RemoveFromCart")
}

// ---------------------------------------------------------------------------
// ClearCart
// ---------------------------------------------------------------------------

func TestClearCart_Success(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("ClearCart", mock.Anything, 1).Return(nil)
	cache.On("InvalidateCart", mock.Anything, 1).Return(nil)

	err := svc.ClearCart(context.Background(), 1)
	require.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestClearCart_RepoError(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("ClearCart", mock.Anything, 1).Return(errors.New("db error"))

	err := svc.ClearCart(context.Background(), 1)
	require.Error(t, err)
	cache.AssertNotCalled(t, "InvalidateCart")
}

// ---------------------------------------------------------------------------
// UpdateCartItemQuantity
// ---------------------------------------------------------------------------

func TestUpdateCartItemQuantity_Success(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("UpdateCartItemQuantity", mock.Anything, 1, "item-1", 5).Return(nil)
	cache.On("UpdateCartItemQuantity", mock.Anything, 1, "item-1", 5).Return(nil)

	err := svc.UpdateCartItemQuantity(context.Background(), 1, "item-1", 5)
	require.NoError(t, err)
}

func TestUpdateCartItemQuantity_CacheFail_Invalidates(t *testing.T) {
	cache := new(mocks.MockCartCache)
	repo := new(mocks.MockCartRepository)
	svc := services.NewCartCacheAsideService(repo, cache, newTestLogger(), testTTL)

	repo.On("UpdateCartItemQuantity", mock.Anything, 1, "item-1", 5).Return(nil)
	cache.On("UpdateCartItemQuantity", mock.Anything, 1, "item-1", 5).
		Return(errors.New("redis error"))
	cache.On("InvalidateCart", mock.Anything, 1).Return(nil)

	err := svc.UpdateCartItemQuantity(context.Background(), 1, "item-1", 5)
	require.NoError(t, err)
	cache.AssertCalled(t, "InvalidateCart", mock.Anything, 1)
}
