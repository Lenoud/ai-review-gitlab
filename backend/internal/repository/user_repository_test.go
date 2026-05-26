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

func TestUserRepositoryCreatesRoleWithPermissions(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	read := model.SysPermission{Code: "project:read", Name: "查看项目", Category: "project"}
	write := model.SysPermission{Code: "project:write", Name: "管理项目", Category: "project"}
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&write).Error)
	repo := NewUserRepository(db)

	role, err := repo.CreateRole(context.Background(), service.RoleInput{
		Code:          "reviewer",
		Name:          "审查员",
		Description:   "can review projects",
		PermissionIDs: []uint{write.ID, read.ID},
	})

	require.NoError(t, err)
	require.Equal(t, "reviewer", role.Code)
	require.Equal(t, "审查员", role.Name)
	require.Equal(t, "can review projects", role.Description)
	require.ElementsMatch(t, []uint{read.ID, write.ID}, role.PermissionIDs)
}

func TestUserRepositoryUpdatesRoleAndReplacesPermissions(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	read := model.SysPermission{Code: "project:read", Name: "查看项目", Category: "project"}
	write := model.SysPermission{Code: "project:write", Name: "管理项目", Category: "project"}
	role := model.SysRole{Code: "reviewer", Name: "审查员", Description: "old"}
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&write).Error)
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&model.SysRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
	repo := NewUserRepository(db)

	got, err := repo.UpdateRole(context.Background(), role.ID, service.RoleInput{
		Code:          "maintainer",
		Name:          "维护者",
		Description:   "new",
		PermissionIDs: []uint{write.ID},
	})

	require.NoError(t, err)
	require.Equal(t, "maintainer", got.Code)
	require.Equal(t, "维护者", got.Name)
	require.Equal(t, "new", got.Description)
	require.Equal(t, []uint{write.ID}, got.PermissionIDs)
}

func TestUserRepositoryFindRoleByIDIncludesPermissions(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	role := model.SysRole{Code: "reviewer", Name: "审查员", Description: "can review"}
	read := model.SysPermission{Code: "project:read", Name: "查看项目", Category: "project"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&model.SysRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
	repo := NewUserRepository(db)

	got, err := repo.FindRoleByID(context.Background(), role.ID)

	require.NoError(t, err)
	require.Equal(t, role.ID, got.ID)
	require.Equal(t, "reviewer", got.Code)
	require.Equal(t, []uint{read.ID}, got.PermissionIDs)
}

func TestUserRepositoryReturnsRoleConflictForDuplicateCode(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	require.NoError(t, db.Create(&model.SysRole{Code: "reviewer", Name: "审查员"}).Error)
	repo := NewUserRepository(db)

	_, err := repo.CreateRole(context.Background(), service.RoleInput{
		Code: "reviewer",
		Name: "重复",
	})

	require.ErrorIs(t, err, service.ErrRoleCodeExists)
}

func TestUserRepositoryRejectsUnknownRolePermissionIDs(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.CreateRole(context.Background(), service.RoleInput{
		Code:          "reviewer",
		Name:          "审查员",
		PermissionIDs: []uint{999},
	})

	require.ErrorIs(t, err, service.ErrInvalidRBACInput)
}

func TestUserRepositoryDeletesRolesAndPermissionBindings(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	role := model.SysRole{Code: "reviewer", Name: "审查员"}
	read := model.SysPermission{Code: "project:read", Name: "查看项目", Category: "project"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&model.SysRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
	repo := NewUserRepository(db)

	err := repo.DeleteRoles(context.Background(), []uint{role.ID})

	require.NoError(t, err)
	var roleCount int64
	require.NoError(t, db.Model(&model.SysRole{}).Where("id = ?", role.ID).Count(&roleCount).Error)
	require.Zero(t, roleCount)
	var bindingCount int64
	require.NoError(t, db.Model(&model.SysRolePermission{}).Where("role_id = ?", role.ID).Count(&bindingCount).Error)
	require.Zero(t, bindingCount)
}

func TestUserRepositoryRejectsDeletingRoleAssignedToUser(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	user := model.SysUser{Username: "alice", PasswordHash: "hash", Status: service.UserStatusEnabled}
	role := model.SysRole{Code: "reviewer", Name: "审查员"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&model.SysUserRole{UserID: user.ID, RoleID: role.ID}).Error)
	repo := NewUserRepository(db)

	err := repo.DeleteRoles(context.Background(), []uint{role.ID})

	require.ErrorIs(t, err, service.ErrRoleInUse)
}

func TestUserRepositoryRejectsDeletingMissingRole(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	repo := NewUserRepository(db)

	err := repo.DeleteRoles(context.Background(), []uint{999})

	require.ErrorIs(t, err, service.ErrRoleNotFound)
}

func TestUserRepositoryCreatesAdminUserWithRoles(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	admin := model.SysRole{Code: "admin", Name: "管理员"}
	reviewer := model.SysRole{Code: "reviewer", Name: "审查员"}
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&reviewer).Error)
	repo := NewUserRepository(db)

	user, err := repo.CreateUser(context.Background(), service.AdminUserInput{
		Username:     "alice",
		PasswordHash: "hashed-password",
		Nickname:     "Alice",
		Remark:       "team lead",
		RoleIDs:      []uint{reviewer.ID, admin.ID},
	})

	require.NoError(t, err)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "Alice", user.Nickname)
	require.Equal(t, "team lead", user.Remark)
	require.ElementsMatch(t, []uint{admin.ID, reviewer.ID}, user.RoleIDs)
	require.ElementsMatch(t, []service.Role{
		{ID: admin.ID, Code: "admin", Name: "管理员"},
		{ID: reviewer.ID, Code: "reviewer", Name: "审查员"},
	}, user.Roles)
	var record model.SysUser
	require.NoError(t, db.First(&record, user.ID).Error)
	require.Equal(t, "hashed-password", record.PasswordHash)
}

func TestUserRepositoryUpdatesAdminUserAndPreservesPasswordWhenBlank(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	user := model.SysUser{Username: "alice", PasswordHash: "old-hash", Nickname: "Alice", Status: service.UserStatusEnabled}
	oldRole := model.SysRole{Code: "reviewer", Name: "审查员"}
	newRole := model.SysRole{Code: "admin", Name: "管理员"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&oldRole).Error)
	require.NoError(t, db.Create(&newRole).Error)
	require.NoError(t, db.Create(&model.SysUserRole{UserID: user.ID, RoleID: oldRole.ID}).Error)
	repo := NewUserRepository(db)

	got, err := repo.UpdateUser(context.Background(), user.ID, service.AdminUserInput{
		Username: "alice2",
		Nickname: "Alice Liddell",
		Remark:   "new remark",
		RoleIDs:  []uint{newRole.ID},
	})

	require.NoError(t, err)
	require.Equal(t, "alice2", got.Username)
	require.Equal(t, []uint{newRole.ID}, got.RoleIDs)
	var record model.SysUser
	require.NoError(t, db.First(&record, user.ID).Error)
	require.Equal(t, "old-hash", record.PasswordHash)
	require.Equal(t, "new remark", record.Remark)
}

func TestUserRepositoryUpdateAdminUserPreservesUsernameWhenBlank(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	user := model.SysUser{Username: "alice", PasswordHash: "old-hash", Nickname: "Alice", Status: service.UserStatusEnabled}
	role := model.SysRole{Code: "reviewer", Name: "审查员"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&role).Error)
	repo := NewUserRepository(db)

	got, err := repo.UpdateUser(context.Background(), user.ID, service.AdminUserInput{
		Nickname:     "Alice Liddell",
		PasswordHash: "new-hash",
		RoleIDs:      []uint{role.ID},
	})

	require.NoError(t, err)
	require.Equal(t, "alice", got.Username)
	var record model.SysUser
	require.NoError(t, db.First(&record, user.ID).Error)
	require.Equal(t, "alice", record.Username)
	require.Equal(t, "new-hash", record.PasswordHash)
}

func TestUserRepositorySearchesAdminUsersWithRoles(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	adminRole := model.SysRole{Code: "admin", Name: "管理员"}
	reviewerRole := model.SysRole{Code: "reviewer", Name: "审查员"}
	alice := model.SysUser{Username: "alice", PasswordHash: "hash", Nickname: "Alice", Status: service.UserStatusEnabled}
	bob := model.SysUser{Username: "bob", PasswordHash: "hash", Nickname: "Builder", Status: service.UserStatusEnabled}
	require.NoError(t, db.Create(&adminRole).Error)
	require.NoError(t, db.Create(&reviewerRole).Error)
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)
	require.NoError(t, db.Create(&model.SysUserRole{UserID: alice.ID, RoleID: reviewerRole.ID}).Error)
	require.NoError(t, db.Create(&model.SysUserRole{UserID: bob.ID, RoleID: adminRole.ID}).Error)
	repo := NewUserRepository(db)

	page, err := repo.SearchUsers(context.Background(), service.AdminUserSearchQuery{
		Keyword: "ali",
		Page:    1,
		Size:    20,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "alice", page.Items[0].Username)
	require.Equal(t, []uint{reviewerRole.ID}, page.Items[0].RoleIDs)
	require.Equal(t, []service.Role{{ID: reviewerRole.ID, Code: "reviewer", Name: "审查员"}}, page.Items[0].Roles)
}

func TestUserRepositoryRejectsDuplicateUsernameAndUnknownRole(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	role := model.SysRole{Code: "reviewer", Name: "审查员"}
	require.NoError(t, db.Create(&model.SysUser{Username: "alice", PasswordHash: "hash", Status: service.UserStatusEnabled}).Error)
	require.NoError(t, db.Create(&role).Error)
	repo := NewUserRepository(db)

	_, err := repo.CreateUser(context.Background(), service.AdminUserInput{
		Username:     "alice",
		PasswordHash: "new-hash",
		RoleIDs:      []uint{role.ID},
	})
	require.ErrorIs(t, err, service.ErrUsernameExists)

	_, err = repo.CreateUser(context.Background(), service.AdminUserInput{
		Username:     "bob",
		PasswordHash: "new-hash",
		RoleIDs:      []uint{999},
	})
	require.ErrorIs(t, err, service.ErrInvalidRBACInput)
}

func TestUserRepositoryListsRoleOptionsInCodeOrder(t *testing.T) {
	db := openUserRepositoryTestDB(t)
	require.NoError(t, db.Create(&model.SysRole{Code: "reviewer", Name: "审查员"}).Error)
	require.NoError(t, db.Create(&model.SysRole{Code: "admin", Name: "管理员"}).Error)
	repo := NewUserRepository(db)

	roles, err := repo.ListRoleOptions(context.Background())

	require.NoError(t, err)
	require.Equal(t, []service.Role{
		{ID: 2, Code: "admin", Name: "管理员"},
		{ID: 1, Code: "reviewer", Name: "审查员"},
	}, roles)
}

func openUserRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.SysUser{}, &model.SysRole{}, &model.SysPermission{}, &model.SysUserRole{}, &model.SysRolePermission{}))
	return db
}
