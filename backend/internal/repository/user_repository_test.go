package repository

import (
	"context"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserRepositoryFindsUserByUsername(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	require.NoError(t, db.Create(&model.SysUser{
		Username:     "admin",
		PasswordHash: "hashed-password",
		Nickname:     "Administrator",
		Status:       service.UserStatusEnabled,
	}).Error)

	repo := NewUserRepository(db)

	user, err := repo.FindByUsername(context.Background(), "admin")

	require.NoError(t, err)
	require.Equal(t, "admin", user.Username)
	require.Equal(t, "hashed-password", user.PasswordHash)
	require.Equal(t, "Administrator", user.Nickname)
	require.Equal(t, service.UserStatusEnabled, user.Status)
}

func TestUserRepositoryReturnsDomainNotFound(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByUsername(context.Background(), "missing")

	require.ErrorIs(t, err, service.ErrUserNotFound)
}

func openUserRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.SysUser{}))
	return db
}
