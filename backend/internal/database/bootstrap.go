package database

import (
	"context"
	"errors"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

var builtinPermissions = []model.SysPermission{
	{Code: "project:read", Name: "查看项目", Category: "project"},
	{Code: "project:write", Name: "管理项目", Category: "project"},
	{Code: "llm-model:read", Name: "查看 LLM 模型", Category: "llm-model"},
	{Code: "llm-model:write", Name: "管理 LLM 模型", Category: "llm-model"},
	{Code: "review-log:read", Name: "查看审查日志", Category: "review-log"},
	{Code: "review-log:write", Name: "管理审查日志", Category: "review-log"},
	{Code: "ai-review-trace:read", Name: "查看 AI 审查链路", Category: "ai-review-trace"},
	{Code: "ai-review-trace:write", Name: "管理 AI 审查链路", Category: "ai-review-trace"},
	{Code: "im-robot:read", Name: "查看 IM 机器人", Category: "im-robot"},
	{Code: "im-robot:write", Name: "管理 IM 机器人", Category: "im-robot"},
	{Code: "member-im-mapping:read", Name: "查看成员 IM 映射", Category: "member-im-mapping"},
	{Code: "member-im-mapping:write", Name: "管理成员 IM 映射", Category: "member-im-mapping"},
	{Code: "project-template:read", Name: "查看项目模板", Category: "project-template"},
	{Code: "project-template:write", Name: "管理项目模板", Category: "project-template"},
	{Code: "project-analysis-plan:read", Name: "查看项目分析计划", Category: "project-analysis-plan"},
	{Code: "project-analysis-plan:write", Name: "管理项目分析计划", Category: "project-analysis-plan"},
	{Code: "rbac:read", Name: "查看用户角色权限", Category: "rbac"},
	{Code: "rbac:write", Name: "管理用户角色权限", Category: "rbac"},
	{Code: "stats:read", Name: "查看统计数据", Category: "stats"},
	{Code: "sys-log:read", Name: "查看系统日志", Category: "sys-log"},
	{Code: "system:read", Name: "查看系统配置", Category: "system"},
	{Code: "system:write", Name: "管理系统配置", Category: "system"},
}

type BootstrapAdminOptions struct {
	Username string
	Password string
	Nickname string
}

func BootstrapAdmin(ctx context.Context, db *gorm.DB, opts BootstrapAdminOptions) error {
	username := strings.TrimSpace(opts.Username)
	if username == "" {
		username = "admin"
	}
	if opts.Password == "" {
		return errors.New("admin password is required")
	}
	nickname := strings.TrimSpace(opts.Nickname)
	if nickname == "" {
		nickname = "Administrator"
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role model.SysRole
		if err := tx.Where(model.SysRole{Code: "admin"}).FirstOrCreate(&role, model.SysRole{
			Code:        "admin",
			Name:        "超级管理员",
			Description: "系统内置管理员角色",
		}).Error; err != nil {
			return err
		}
		permissions, err := seedBuiltinPermissions(ctx, tx)
		if err != nil {
			return err
		}
		if err := bindRolePermissions(ctx, tx, role.ID, permissions); err != nil {
			return err
		}

		hash, err := service.HashPassword(opts.Password)
		if err != nil {
			return err
		}

		var user model.SysUser
		err = tx.Where("username = ?", username).First(&user).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = model.SysUser{
				Username:     username,
				PasswordHash: hash,
				Nickname:     nickname,
				Status:       service.UserStatusEnabled,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&user).Updates(map[string]any{
				"password_hash": hash,
				"nickname":      nickname,
				"status":        service.UserStatusEnabled,
			}).Error; err != nil {
				return err
			}
		}

		return tx.Where(model.SysUserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}).FirstOrCreate(&model.SysUserRole{}, model.SysUserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}).Error
	})
}

func seedBuiltinPermissions(ctx context.Context, tx *gorm.DB) ([]model.SysPermission, error) {
	permissions := make([]model.SysPermission, 0, len(builtinPermissions))
	for _, item := range builtinPermissions {
		var permission model.SysPermission
		err := tx.WithContext(ctx).Where(model.SysPermission{Code: item.Code}).FirstOrCreate(&permission, item).Error
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func bindRolePermissions(ctx context.Context, tx *gorm.DB, roleID uint, permissions []model.SysPermission) error {
	for _, permission := range permissions {
		if err := tx.WithContext(ctx).Where(model.SysRolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		}).FirstOrCreate(&model.SysRolePermission{}, model.SysRolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
