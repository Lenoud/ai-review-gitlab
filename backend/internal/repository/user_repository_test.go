package repository

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

func TestUserRepositoryFindsUserWithRolesAndPermissions(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	user := model.SysUser{
		Username:     "reviewer",
		PasswordHash: "hashed-password",
		Nickname:     "Reviewer",
		Status:       service.UserStatusEnabled,
	}
	require.NoError(t, db.Create(&user).Error)
	role := model.SysRole{Code: "reviewer", Name: "审查员"}
	require.NoError(t, db.Create(&role).Error)
	readPermission := model.SysPermission{Code: "review-log:read", Name: "查看审查日志"}
	writePermission := model.SysPermission{Code: "project:write", Name: "编辑项目"}
	require.NoError(t, db.Create(&readPermission).Error)
	require.NoError(t, db.Create(&writePermission).Error)
	require.NoError(t, db.Create(&model.SysUserRole{UserID: user.ID, RoleID: role.ID}).Error)
	require.NoError(t, db.Create(&model.SysRolePermission{RoleID: role.ID, PermissionID: readPermission.ID}).Error)
	require.NoError(t, db.Create(&model.SysRolePermission{RoleID: role.ID, PermissionID: writePermission.ID}).Error)

	repo := NewUserRepository(db)

	got, err := repo.FindByID(context.Background(), user.ID)

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"reviewer"}, got.Roles)
	require.ElementsMatch(t, []string{"review-log:read", "project:write"}, got.Permissions)
}

func TestUserRepositoryListsRolesInCodeOrder(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	require.NoError(t, db.Create(&model.SysRole{Code: "reviewer", Name: "审查员"}).Error)
	require.NoError(t, db.Create(&model.SysRole{Code: "admin", Name: "管理员"}).Error)

	repo := NewUserRepository(db)

	roles, err := repo.ListRoles(context.Background())

	require.NoError(t, err)
	require.Equal(t, []service.Role{
		{ID: 2, Code: "admin", Name: "管理员"},
		{ID: 1, Code: "reviewer", Name: "审查员"},
	}, roles)
}

func TestUserRepositoryListsPermissionsGroupedByCategory(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	require.NoError(t, db.Create(&model.SysPermission{Code: "project:write", Name: "管理项目", Category: "project"}).Error)
	require.NoError(t, db.Create(&model.SysPermission{Code: "project:read", Name: "查看项目", Category: "project"}).Error)
	require.NoError(t, db.Create(&model.SysPermission{Code: "rbac:read", Name: "查看权限", Category: "rbac"}).Error)

	repo := NewUserRepository(db)

	groups, err := repo.ListPermissionGroups(context.Background())

	require.NoError(t, err)
	require.Equal(t, []service.PermissionGroup{
		{
			Category: "project",
			Permissions: []service.Permission{
				{ID: 2, Code: "project:read", Name: "查看项目", Category: "project"},
				{ID: 1, Code: "project:write", Name: "管理项目", Category: "project"},
			},
		},
		{
			Category: "rbac",
			Permissions: []service.Permission{
				{ID: 3, Code: "rbac:read", Name: "查看权限", Category: "rbac"},
			},
		},
	}, groups)
}

func openUserRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.SysUser{}, &model.SysRole{}, &model.SysPermission{}, &model.SysUserRole{}, &model.SysRolePermission{}))
	return db
}
