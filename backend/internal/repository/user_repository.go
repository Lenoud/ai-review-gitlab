package repository

import (
	"context"
	"errors"

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
