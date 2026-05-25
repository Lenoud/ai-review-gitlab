package database

import (
	"context"
	"errors"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

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
