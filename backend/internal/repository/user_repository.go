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
	roles := make([]service.Role, 0, len(records))
	for _, record := range records {
		roles = append(roles, service.Role{
			ID:   record.ID,
			Code: record.Code,
			Name: record.Name,
		})
	}
	return roles, nil
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
