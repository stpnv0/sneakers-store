package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"sso/internal/domain/models"
	"sso/internal/services/auth/mocks"
	"sso/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func newTestAuth(
	saver *mocks.MockUserSaver,
	provider *mocks.MockUserProvider,
	appProvider *mocks.MockAppProvider,
) *Auth {
	return New(testLogger, saver, provider, appProvider, time.Hour)
}

// --- RegisterNewUser ---

func TestRegisterNewUser_Success(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	saver.On("SaveUser", mock.Anything, "test@example.com", mock.Anything).Return(int64(1), nil)

	id, err := svc.RegisterNewUser(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	saver.AssertExpectations(t)
}

func TestRegisterNewUser_UserExists(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	saver.On("SaveUser", mock.Anything, "existing@example.com", mock.Anything).
		Return(int64(0), storage.ErrUserExists)

	_, err := svc.RegisterNewUser(context.Background(), "existing@example.com", "password123")
	assert.Error(t, err)
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	passHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{ID: 1, Email: "test@example.com", PassHash: passHash}
	app := models.App{ID: 1, Name: "test", Secret: "secret123"}

	provider.On("User", mock.Anything, "test@example.com").Return(user, nil)
	appProvider.On("App", mock.Anything, 1).Return(app, nil)

	token, err := svc.Login(context.Background(), "test@example.com", "password123", 1)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	provider.On("User", mock.Anything, "missing@example.com").
		Return(models.User{}, storage.ErrUserNotFound)

	_, err := svc.Login(context.Background(), "missing@example.com", "password123", 1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
}

func TestLogin_WrongPassword(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	passHash, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
	user := models.User{ID: 1, Email: "test@example.com", PassHash: passHash}

	provider.On("User", mock.Anything, "test@example.com").Return(user, nil)

	_, err := svc.Login(context.Background(), "test@example.com", "wrong_password", 1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
}

// --- IsAdmin ---

func TestIsAdmin_True(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	provider.On("IsAdmin", mock.Anything, int64(1)).Return(true, nil)

	result, err := svc.IsAdmin(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestIsAdmin_False(t *testing.T) {
	saver := new(mocks.MockUserSaver)
	provider := new(mocks.MockUserProvider)
	appProvider := new(mocks.MockAppProvider)
	svc := newTestAuth(saver, provider, appProvider)

	provider.On("IsAdmin", mock.Anything, int64(1)).Return(false, nil)

	result, err := svc.IsAdmin(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, result)
}

