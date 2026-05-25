package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBootstrapAdminCreatesAdminUserAndRole(t *testing.T) {
	db := openBootstrapTestDB(t)

	err := BootstrapAdmin(context.Background(), db, BootstrapAdminOptions{
		Username: "admin",
		Password: "admin-password",
		Nickname: "Administrator",
	})

	require.NoError(t, err)

	var user model.SysUser
	require.NoError(t, db.Where("username = ?", "admin").First(&user).Error)
	require.Equal(t, "Administrator", user.Nickname)
	require.Equal(t, service.UserStatusEnabled, user.Status)
	require.True(t, service.CheckPassword(user.PasswordHash, "admin-password"))

	var role model.SysRole
	require.NoError(t, db.Where("code = ?", "admin").First(&role).Error)

	var userRole model.SysUserRole
	require.NoError(t, db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error)
}

func TestBootstrapAdminUpdatesExistingAdminPassword(t *testing.T) {
	db := openBootstrapTestDB(t)
	oldHash, err := service.HashPassword("old-password")
	require.NoError(t, err)
	require.NoError(t, db.Create(&model.SysUser{
		Username:     "admin",
		PasswordHash: oldHash,
		Nickname:     "Old",
		Status:       service.UserStatusEnabled,
	}).Error)

	err = BootstrapAdmin(context.Background(), db, BootstrapAdminOptions{
		Username: "admin",
		Password: "new-password",
		Nickname: "Administrator",
	})

	require.NoError(t, err)

	var user model.SysUser
	require.NoError(t, db.Where("username = ?", "admin").First(&user).Error)
	require.True(t, service.CheckPassword(user.PasswordHash, "new-password"))
	require.Equal(t, "Administrator", user.Nickname)
}

func openBootstrapTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	return db
}
