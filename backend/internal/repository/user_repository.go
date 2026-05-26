package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*service.User, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	return r.toServiceUser(ctx, &user)
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*service.User, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	return r.toServiceUser(ctx, &user)
}

func (r *UserRepository) ListRoles(ctx context.Context) ([]service.Role, error) {
	var records []model.SysRole
	if err := r.db.WithContext(ctx).Order("code ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return rolesToService(records), nil
}

func (r *UserRepository) ListPermissionGroups(ctx context.Context) ([]service.PermissionGroup, error) {
	var records []model.SysPermission
	if err := r.db.WithContext(ctx).Order("category ASC, code ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	groups := make([]service.PermissionGroup, 0)
	indexByCategory := map[string]int{}
	for _, record := range records {
		category := record.Category
		index, ok := indexByCategory[category]
		if !ok {
			index = len(groups)
			indexByCategory[category] = index
			groups = append(groups, service.PermissionGroup{Category: category})
		}
		groups[index].Permissions = append(groups[index].Permissions, service.Permission{
			ID:       record.ID,
			Code:     record.Code,
			Name:     record.Name,
			Category: record.Category,
		})
	}
	return groups, nil
}

func (r *UserRepository) CreateRole(ctx context.Context, input service.RoleInput) (*service.RoleDetail, error) {
	record := model.SysRole{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return mapRoleWriteError(err)
		}
		return replaceRolePermissions(ctx, tx, record.ID, input.PermissionIDs)
	})
	if err != nil {
		return nil, err
	}
	return r.FindRoleByID(ctx, record.ID)
}

func (r *UserRepository) UpdateRole(ctx context.Context, id uint, input service.RoleInput) (*service.RoleDetail, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.SysRole{}).Where("id = ?", id).Updates(map[string]any{
			"code":        input.Code,
			"name":        input.Name,
			"description": input.Description,
		})
		if result.Error != nil {
			return mapRoleWriteError(result.Error)
		}
		var count int64
		if err := tx.Model(&model.SysRole{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return service.ErrRoleNotFound
		}
		return replaceRolePermissions(ctx, tx, id, input.PermissionIDs)
	})
	if err != nil {
		return nil, err
	}
	return r.FindRoleByID(ctx, id)
}

func (r *UserRepository) FindRoleByID(ctx context.Context, id uint) (*service.RoleDetail, error) {
	var record model.SysRole
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrRoleNotFound
		}
		return nil, err
	}
	permissionIDs, err := r.findRolePermissionIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	return &service.RoleDetail{
		ID:            record.ID,
		Code:          record.Code,
		Name:          record.Name,
		Description:   record.Description,
		PermissionIDs: permissionIDs,
	}, nil
}

func (r *UserRepository) DeleteRoles(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleCount int64
		if err := tx.Model(&model.SysRole{}).Where("id IN ?", ids).Count(&roleCount).Error; err != nil {
			return err
		}
		if roleCount != int64(len(ids)) {
			return service.ErrRoleNotFound
		}
		var assignedCount int64
		if err := tx.Model(&model.SysUserRole{}).Where("role_id IN ?", ids).Count(&assignedCount).Error; err != nil {
			return err
		}
		if assignedCount > 0 {
			return service.ErrRoleInUse
		}
		if err := tx.Where("role_id IN ?", ids).Delete(&model.SysRolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.SysRole{}, ids).Error
	})
}

func (r *UserRepository) CreateUser(ctx context.Context, input service.AdminUserInput) (*service.AdminUser, error) {
	record := model.SysUser{
		Username:     input.Username,
		PasswordHash: input.PasswordHash,
		Nickname:     input.Nickname,
		Remark:       input.Remark,
		Status:       service.UserStatusEnabled,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateRoleIDs(ctx, tx, input.RoleIDs); err != nil {
			return err
		}
		if err := tx.Create(&record).Error; err != nil {
			return mapUserWriteError(err)
		}
		return replaceUserRoles(ctx, tx, record.ID, input.RoleIDs)
	})
	if err != nil {
		return nil, err
	}
	return r.FindAdminUserByID(ctx, record.ID)
}

func (r *UserRepository) UpdateUser(ctx context.Context, id uint, input service.AdminUserInput) (*service.AdminUser, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateRoleIDs(ctx, tx, input.RoleIDs); err != nil {
			return err
		}
		updates := map[string]any{
			"nickname": input.Nickname,
			"remark":   input.Remark,
		}
		if input.Username != "" {
			updates["username"] = input.Username
		}
		if input.PasswordHash != "" {
			updates["password_hash"] = input.PasswordHash
		}
		result := tx.Model(&model.SysUser{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return mapUserWriteError(result.Error)
		}
		var count int64
		if err := tx.Model(&model.SysUser{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return service.ErrUserNotFound
		}
		return replaceUserRoles(ctx, tx, id, input.RoleIDs)
	})
	if err != nil {
		return nil, err
	}
	return r.FindAdminUserByID(ctx, id)
}

func (r *UserRepository) FindAdminUserByID(ctx context.Context, id uint) (*service.AdminUser, error) {
	var record model.SysUser
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	roles, err := r.findUserRoles(ctx, []uint{id})
	if err != nil {
		return nil, err
	}
	return adminUserToService(&record, roles[id]), nil
}

func (r *UserRepository) SearchUsers(ctx context.Context, query service.AdminUserSearchQuery) (*service.AdminUserPage, error) {
	db := r.db.WithContext(ctx).Model(&model.SysUser{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.SysUser
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Limit(query.Size).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0, len(records))
	for _, record := range records {
		userIDs = append(userIDs, record.ID)
	}
	rolesByUserID, err := r.findUserRoles(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	items := make([]service.AdminUser, 0, len(records))
	for i := range records {
		items = append(items, *adminUserToService(&records[i], rolesByUserID[records[i].ID]))
	}
	return &service.AdminUserPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func (r *UserRepository) ListRoleOptions(ctx context.Context) ([]service.Role, error) {
	return r.ListRoles(ctx)
}

func (r *UserRepository) toServiceUser(ctx context.Context, user *model.SysUser) (*service.User, error) {
	roles, err := r.findRoleCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	permissions, err := r.findPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &service.User{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		Status:       user.Status,
		Roles:        roles,
		Permissions:  permissions,
	}, nil
}

func (r *UserRepository) findUserRoles(ctx context.Context, userIDs []uint) (map[uint][]service.Role, error) {
	result := map[uint][]service.Role{}
	if len(userIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		UserID uint
		ID     uint
		Code   string
		Name   string
	}
	err := r.db.WithContext(ctx).
		Table("sys_user_role AS ur").
		Select("ur.user_id, r.id, r.code, r.name").
		Joins("JOIN sys_role AS r ON r.id = ur.role_id").
		Where("ur.user_id IN ?", userIDs).
		Order("r.code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], service.Role{
			ID:   row.ID,
			Code: row.Code,
			Name: row.Name,
		})
	}
	return result, nil
}

func (r *UserRepository) findRoleCodes(ctx context.Context, userID uint) ([]string, error) {
	var roles []string
	err := r.db.WithContext(ctx).
		Table("sys_role AS r").
		Select("r.code").
		Joins("JOIN sys_user_role AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Order("r.code ASC").
		Pluck("r.code", &roles).Error
	return roles, err
}

func (r *UserRepository) findRolePermissionIDs(ctx context.Context, roleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).
		Table("sys_role_permission").
		Select("permission_id").
		Where("role_id = ?", roleID).
		Order("permission_id ASC").
		Pluck("permission_id", &ids).Error
	return ids, err
}

func (r *UserRepository) findPermissionCodes(ctx context.Context, userID uint) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("sys_permission AS p").
		Distinct("p.code").
		Joins("JOIN sys_role_permission AS rp ON rp.permission_id = p.id").
		Joins("JOIN sys_user_role AS ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Order("p.code ASC").
		Pluck("p.code", &permissions).Error
	return permissions, err
}

func replaceRolePermissions(ctx context.Context, tx *gorm.DB, roleID uint, permissionIDs []uint) error {
	if err := tx.WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.SysRolePermission{}).Error; err != nil {
		return err
	}
	if len(permissionIDs) == 0 {
		return nil
	}
	var existingCount int64
	if err := tx.WithContext(ctx).Model(&model.SysPermission{}).Where("id IN ?", permissionIDs).Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount != int64(len(permissionIDs)) {
		return service.ErrInvalidRBACInput
	}
	for _, permissionID := range permissionIDs {
		if err := tx.WithContext(ctx).Create(&model.SysRolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func validateRoleIDs(ctx context.Context, tx *gorm.DB, roleIDs []uint) error {
	if len(roleIDs) == 0 {
		return service.ErrInvalidRBACInput
	}
	var existingCount int64
	if err := tx.WithContext(ctx).Model(&model.SysRole{}).Where("id IN ?", roleIDs).Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount != int64(len(roleIDs)) {
		return service.ErrInvalidRBACInput
	}
	return nil
}

func replaceUserRoles(ctx context.Context, tx *gorm.DB, userID uint, roleIDs []uint) error {
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.SysUserRole{}).Error; err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if err := tx.WithContext(ctx).Create(&model.SysUserRole{
			UserID: userID,
			RoleID: roleID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func mapRoleWriteError(err error) error {
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") || strings.Contains(lower, "constraint failed") {
		return service.ErrRoleCodeExists
	}
	return err
}

func mapUserWriteError(err error) error {
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "username") &&
		(strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") || strings.Contains(lower, "constraint failed")) {
		return service.ErrUsernameExists
	}
	return err
}

func adminUserToService(record *model.SysUser, roles []service.Role) *service.AdminUser {
	roleIDs := make([]uint, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	return &service.AdminUser{
		ID:       record.ID,
		Username: record.Username,
		Nickname: record.Nickname,
		Remark:   record.Remark,
		RoleIDs:  roleIDs,
		Roles:    roles,
	}
}

func rolesToService(records []model.SysRole) []service.Role {
	roles := make([]service.Role, 0, len(records))
	for _, record := range records {
		roles = append(roles, service.Role{
			ID:   record.ID,
			Code: record.Code,
			Name: record.Name,
		})
	}
	return roles
}
