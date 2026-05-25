package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceLoginIssuesUsableAccessAndRefreshTokens(t *testing.T) {
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	repo := newMemoryUserRepository(&User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hash,
		Nickname:     "Administrator",
		Status:       UserStatusEnabled,
	})
	svc, err := NewAuthService(repo, AuthOptions{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
		Issuer:          "ai-review-test",
		Now:             time.Now,
	})
	require.NoError(t, err)

	tokens, err := svc.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "correct-password",
	})

	require.NoError(t, err)
	require.Equal(t, "Bearer", tokens.TokenType)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Equal(t, int64(60), tokens.ExpiresIn)
	require.Equal(t, int64(3600), tokens.RefreshExpiresIn)

	subject, err := svc.ValidateAccessToken(context.Background(), tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, uint(1), subject.UserID)
	require.Equal(t, "admin", subject.Username)
	require.Equal(t, "Administrator", subject.Nickname)

	refreshed, err := svc.Refresh(context.Background(), tokens.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, refreshed.AccessToken)

	_, err = svc.ValidateAccessToken(context.Background(), tokens.RefreshToken)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthServiceRejectsInvalidCredentials(t *testing.T) {
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	svc, err := NewAuthService(newMemoryUserRepository(&User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hash,
		Status:       UserStatusEnabled,
	}), AuthOptions{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
		Issuer:          "ai-review-test",
	})
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "wrong-password",
	})

	require.ErrorIs(t, err, ErrInvalidCredentials)
}

type memoryUserRepository struct {
	byID       map[uint]*User
	byUsername map[string]*User
}

func newMemoryUserRepository(users ...*User) *memoryUserRepository {
	repo := &memoryUserRepository{
		byID:       map[uint]*User{},
		byUsername: map[string]*User{},
	}
	for _, user := range users {
		repo.byID[user.ID] = user
		repo.byUsername[user.Username] = user
	}
	return repo
}

func (r *memoryUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	user, ok := r.byUsername[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *memoryUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	user, ok := r.byID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}
